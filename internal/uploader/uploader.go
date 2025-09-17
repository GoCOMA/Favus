package uploader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/GoCOMA/Favus/internal/chunker"
	"github.com/GoCOMA/Favus/internal/config"
	"github.com/GoCOMA/Favus/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/schollz/progressbar/v3"
)

// Uploader manages file uploads, deletions, and multipart upload operations for S3.
type Uploader struct {
	s3Client *s3.Client
	Config   *config.Config
}

// ResumeUpload proxies to ResumeUploader so main can call on *Uploader.
func (u *Uploader) ResumeUpload(statusFilePath string) error {
	ru := NewResumeUploader(u.s3Client)
	return ru.ResumeUpload(statusFilePath)
}

// checkBucket verifies that the bucket exists and that the caller has permissions.
func (u *Uploader) checkBucket(bucket string) error {
	_, err := u.s3Client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound", "NoSuchBucket":
				return fmt.Errorf("bucket %s does not exist", bucket)
			case "Forbidden", "AccessDenied":
				return fmt.Errorf("bucket %s exists but access is denied", bucket)
			}
		}
		return fmt.Errorf("failed to check bucket %s: %w", bucket, err)
	}
	return nil
}

// NewUploader creates and returns a new Uploader instance with an initialized AWS S3 client (v2).
func NewUploader(cfgApp *config.Config) (*Uploader, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")

	var (
		awsCfg aws.Config
		err    error
	)
	if endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					HostnameImmutable: true,
				}, nil
			})
		awsCfg, err = awsv2cfg.LoadDefaultConfig(context.Background(),
			awsv2cfg.WithRegion(cfgApp.Region),
			awsv2cfg.WithEndpointResolverWithOptions(resolver),
		)
	} else {
		awsCfg, err = awsv2cfg.LoadDefaultConfig(context.Background(),
			awsv2cfg.WithRegion(cfgApp.Region),
		)
	}
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to load AWS v2 config: %v", err))
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true // For LocalStack or custom S3-compatible endpoints
		}
	})

	return &Uploader{
		s3Client: cli,
		Config:   cfgApp,
	}, nil
}

// UploadFile performs a multipart upload of a local file to S3.
func (u *Uploader) UploadFile(filePath, s3Key string) error {
	utils.Info(fmt.Sprintf("Starting multipart upload for file: %s to s3://%s/%s", filePath, u.Config.Bucket, s3Key))

	// Bucket verification
	if err := u.checkBucket(u.Config.Bucket); err != nil {
		utils.Error(fmt.Sprintf("%v", err))
		return err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to get file info for %s: %v", filePath, err))
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		utils.Info(fmt.Sprintf("File %s is empty, skipping upload", filePath))
		return nil
	}

	// WS reporter (에이전트가 떠있을 때만 실제로 전송)
	r := newWSReporter(fileInfo.Size())

	fileChunker, err := chunker.NewFileChunker(filePath, u.Config.PartSizeBytes())
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create file chunker for %s: %v", filePath, err))
		r.error(fmt.Sprintf("create chunker: %v", err), nil)
		return fmt.Errorf("failed to create file chunker: %w", err)
	}
	chunks := fileChunker.Chunks()

	// Progress bars: total + per-part
	totalBar := progressbar.NewOptions64(
		fileInfo.Size(),
		progressbar.OptionSetDescription("total"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		// progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWriter(os.Stdout),
	)

	// Initiate multipart upload
	initiateOutput, err := u.s3Client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: &u.Config.Bucket,
		Key:    &s3Key,
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to initiate multipart upload for %s: %v", s3Key, err))
		r.error(fmt.Sprintf("initiate multipart: %v", err), nil)
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}
	uploadID := aws.ToString(initiateOutput.UploadId)
	utils.Info(fmt.Sprintf("Initiated multipart upload with UploadID: %s", uploadID))

	// WS: 세션 시작
	r.start(u.Config.Bucket, s3Key, uploadID, u.Config.PartSizeBytes(), nil)

	// Prepare status tracker
	statusFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.upload_status", filepath.Base(filePath), uploadID[:8]))
	status := NewWSTracker(
		NewUploadStatus(filePath, u.Config.Bucket, s3Key, uploadID, len(chunks), u.Config.PartSizeBytes()),
	)

	var completedParts []s3types.CompletedPart
	// Concurrently upload chunks
	maxConcurrency := u.Config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	jobs := make(chan chunker.Chunk, len(chunks))
	results := make(chan s3types.CompletedPart, len(chunks))
	errs := make(chan error, len(chunks))

	for w := 1; w <= maxConcurrency; w++ {
		go func(workerID int) {
			for ch := range jobs {
				utils.Info(fmt.Sprintf("[Worker %d] Uploading part %d for file %s", workerID, ch.Index, filePath))

				reader, err := fileChunker.GetChunkReader(ch)
				if err != nil {
					errs <- fmt.Errorf("[Worker %d] failed to get chunk reader for part %d: %w", workerID, ch.Index, err)
					return
				}

				r.partStart(ch.Index, ch.Size, ch.Offset)

				// Retry logic for each part
				var uploadOutput *s3.UploadPartOutput
				err = utils.Retry(5, 2*time.Second, func() error {
					// Reset reader to the beginning of the chunk for each retry
					_, seekErr := reader.Seek(0, 0)
					if seekErr != nil {
						return fmt.Errorf("failed to seek chunk reader for part %d: %w", ch.Index, seekErr)
					}

					pr := NewReadSeekCloserProgress(reader, func(n int64) {
						_ = totalBar.Add64(n)
						r.progressAdd(n)
						r.partProgressAdd(ch.Index, n)
					})

					var partErr error
					uploadOutput, partErr = u.s3Client.UploadPart(context.Background(), &s3.UploadPartInput{
						Body:          pr,
						Bucket:        &u.Config.Bucket,
						Key:           &s3Key,
						PartNumber:    aws.Int32(int32(ch.Index)),
						UploadId:      &uploadID,
						ContentLength: aws.Int64(ch.Size),
					})
					if partErr != nil {
						utils.Error(fmt.Sprintf("[Worker %d] Failed to upload part %d: %v", workerID, ch.Index, partErr))
						return partErr
					}
					return nil
				})
				_ = reader.Close()

				if err != nil {
					errs <- fmt.Errorf("[Worker %d] failed to upload part %d after retries: %w", workerID, ch.Index, err)
					return
				}

				if uploadOutput.ETag == nil {
					errs <- fmt.Errorf("[Worker %d] ETag for part %d is nil", workerID, ch.Index)
					return
				}

				part := s3types.CompletedPart{
					PartNumber: aws.Int32(int32(ch.Index)),
					ETag:       uploadOutput.ETag,
				}
				results <- part
				status.AddCompletedPart(ch.Index, *uploadOutput.ETag)
				if err := status.SaveStatus(statusFilePath); err != nil {
					utils.Error(fmt.Sprintf("[Worker %d] Failed to save status for part %d: %v", workerID, ch.Index, err))
				}
				utils.Info(fmt.Sprintf("[Worker %d] Successfully uploaded part %d. ETag: %s", workerID, ch.Index, *uploadOutput.ETag))
				r.partDone(ch.Index, ch.Size, *uploadOutput.ETag)
			}
		}(w)
	}

	for _, ch := range chunks {
		jobs <- ch
	}
	close(jobs)

	for i := 0; i < len(chunks); i++ {
		select {
		case part := <-results:
			completedParts = append(completedParts, part)
		case err := <-errs:
			utils.Error(fmt.Sprintf("An error occurred during upload: %v", err))
			_ = u.AbortMultipartUpload(s3Key, uploadID)
			r.done(false, uploadID)
			return err
		}
	}

	// Complete 전에 파트 오름차순 정렬(안전)
	sort.Slice(completedParts, func(i, j int) bool {
		return aws.ToInt32(completedParts[i].PartNumber) < aws.ToInt32(completedParts[j].PartNumber)
	})

	// Complete the multipart upload
	utils.Info(fmt.Sprintf("Completing multipart upload for file: %s", filePath))
	_, err = u.s3Client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   &u.Config.Bucket,
		Key:      &s3Key,
		UploadId: &uploadID,
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to complete multipart upload: %v", err))
		r.error(fmt.Sprintf("complete multipart: %v", err), nil)
		_ = u.AbortMultipartUpload(s3Key, uploadID)
		r.done(false, uploadID)
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	utils.Info(fmt.Sprintf("Multipart upload completed successfully for %s", filePath))
	r.done(true, uploadID)

	// Add a small delay to ensure WebSocket messages are sent before exiting.
	time.Sleep(1 * time.Second)

	// Clean up status file
	if err := os.Remove(statusFilePath); err != nil {
		utils.Error(fmt.Sprintf("Failed to remove status file %s: %v", statusFilePath, err))
	}

	return nil
}

// DeleteFile deletes a specific object from the configured S3 bucket.
func (u *Uploader) DeleteFile(s3Key string) error {
	utils.Info(fmt.Sprintf("Deleting file s3://%s/%s", u.Config.Bucket, s3Key))
	_, err := u.s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &u.Config.Bucket,
		Key:    &s3Key,
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to delete file %s from S3: %v", s3Key, err))
		return fmt.Errorf("failed to delete file %s from S3: %w", s3Key, err)
	}
	utils.Info(fmt.Sprintf("Successfully deleted file s3://%s/%s", u.Config.Bucket, s3Key))
	return nil
}

// AbortMultipartUpload aborts an ongoing multipart upload in S3.
func (u *Uploader) AbortMultipartUpload(s3Key, uploadID string) error {
	utils.Info(fmt.Sprintf("Aborting multipart upload for key: %s, UploadID: %s", s3Key, uploadID))
	_, err := u.s3Client.AbortMultipartUpload(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   &u.Config.Bucket,
		Key:      &s3Key,
		UploadId: &uploadID,
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to abort multipart upload for key %s, UploadID %s: %v", s3Key, uploadID, err))
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}
	utils.Info(fmt.Sprintf("Multipart upload aborted successfully for key: %s, UploadID: %s", s3Key, uploadID))
	return nil
}

// ListMultipartUploads lists all ongoing multipart uploads for the configured S3 bucket.
func (u *Uploader) ListMultipartUploads() ([]s3types.MultipartUpload, error) {
	utils.Info(fmt.Sprintf("Listing ongoing multipart uploads for bucket: %s", u.Config.Bucket))
	output, err := u.s3Client.ListMultipartUploads(context.Background(), &s3.ListMultipartUploadsInput{
		Bucket: &u.Config.Bucket,
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to list multipart uploads: %v", err))
		return nil, fmt.Errorf("failed to list multipart uploads: %w", err)
	}
	return output.Uploads, nil
}

// NewUploaderWithAWSConfig lets callers provide a pre-built aws.Config (e.g., from awsutils.LoadAWSConfig).
func NewUploaderWithAWSConfig(cfgApp *config.Config, awsCfg aws.Config) (*Uploader, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true // LocalStack / custom S3-compatible endpoints
		}
	})
	return &Uploader{
		s3Client: cli,
		Config:   cfgApp,
	}, nil
}

// ListObjects lists completed objects (not multipart sessions) in the configured bucket.
// Optionally filter by a key prefix and limit results with maxKeys (>0).
func (u *Uploader) ListObjects(prefix string, maxKeys int32) ([]s3types.Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: &u.Config.Bucket,
	}
	if prefix != "" {
		input.Prefix = &prefix
	}
	if maxKeys > 0 {
		input.MaxKeys = aws.Int32(maxKeys)
	}

	var results []s3types.Object
	var token *string
	for {
		input.ContinuationToken = token
		out, err := u.s3Client.ListObjectsV2(context.Background(), input)
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}
		results = append(results, out.Contents...)
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
		if maxKeys > 0 && int32(len(results)) >= maxKeys {
			// Trim to maxKeys if we overshot
			if int32(len(results)) > maxKeys {
				results = results[:maxKeys]
			}
			break
		}
	}
	return results, nil
}
