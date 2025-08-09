package uploader

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"favus/internal/chunker"
	"favus/internal/config"
	"favus/pkg/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Uploader manages file uploads, deletions, and multipart upload operations for S3.
type Uploader struct {
	s3Client *s3.S3
	Config   *config.Config
}

func (u *Uploader) ResumeUpload(statusFilePath string) error {
	ru := NewResumeUploader(u.s3Client)
	return ru.ResumeUpload(statusFilePath)
}

// checkBucket verifies that the bucket exists and that the caller has permissions.
func (u *Uploader) checkBucket(bucket string) error {
	_, err := u.s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("bucket %s does not exist", bucket)
			case "Forbidden", "AccessDenied":
				return fmt.Errorf("bucket %s exists but access is denied", bucket)
			}
		}
		return fmt.Errorf("failed to check bucket %s: %w", bucket, err)
	}
	return nil
}

// NewUploader creates and returns a new Uploader instance with an initialized AWS S3 client.
func NewUploader(cfg *config.Config) (*Uploader, error) {
	awsConfig := &aws.Config{
		Region: aws.String(cfg.AWSRegion),
	}

	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		awsConfig.Endpoint = aws.String(endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true) // LocalStack S3는 path-style 필요
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &Uploader{
		s3Client: s3.New(sess),
		Config:   cfg,
	}, nil
}

// UploadFile performs a multipart upload of a local file to S3.
// A temporary status file is created and cleaned up upon successful upload.
func (u *Uploader) UploadFile(filePath, s3Key string) error {
	utils.Info(fmt.Sprintf("Starting multipart upload for file: %s to s3://%s/%s", filePath, u.Config.S3BucketName, s3Key)) // Use S3BucketName from config

	if err := u.checkBucket(u.Config.S3BucketName); err != nil {
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

	fileChunker, err := chunker.NewFileChunker(filePath, u.Config.ChunkSize)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create file chunker for %s: %v", filePath, err))
		return fmt.Errorf("failed to create file chunker: %w", err)
	}
	chunks := fileChunker.Chunks()

	// Initiate multipart upload with S3.
	initiateOutput, err := u.s3Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(u.Config.S3BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to initiate multipart upload for %s: %v", s3Key, err))
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}
	uploadID := *initiateOutput.UploadId
	utils.Info(fmt.Sprintf("Initiated multipart upload with UploadID: %s", uploadID))

	// Prepare a status tracker to save upload progress.
	// Status file name includes part of the UploadID for uniqueness and consistent extension.
	statusFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.upload_status", filepath.Base(filePath), uploadID[:8]))
	status := NewUploadStatus(filePath, u.Config.S3BucketName, s3Key, uploadID, len(chunks))

	var completedParts []*s3.CompletedPart
	// Upload each file chunk.
	for _, ch := range chunks {
		reader, err := fileChunker.GetChunkReader(ch)
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to get chunk reader for part %d: %v", ch.Index, err))
			u.AbortMultipartUpload(s3Key, uploadID)
			return fmt.Errorf("failed to get chunk reader for part %d: %w", ch.Index, err)
		}

		utils.Info(fmt.Sprintf("Uploading part %d (offset %d, size %d) for file %s", ch.Index, ch.Offset, ch.Size, filePath))

		var uploadOutput *s3.UploadPartOutput
		// Retry part upload on transient errors.
		err = utils.Retry(5, 2*time.Second, func() error {
			var partErr error
			uploadOutput, partErr = u.s3Client.UploadPart(&s3.UploadPartInput{
				Body:          aws.ReadSeekCloser(reader),
				Bucket:        aws.String(u.Config.S3BucketName),
				Key:           aws.String(s3Key),
				PartNumber:    aws.Int64(int64(ch.Index)),
				UploadId:      aws.String(uploadID),
				ContentLength: aws.Int64(ch.Size),
			})
			if partErr != nil {
				utils.Error(fmt.Sprintf("Failed to upload part %d: %v", ch.Index, partErr))
				return partErr
			}
			return nil
		})

		if err != nil {
			utils.Error(fmt.Sprintf("Failed to upload part %d after retries: %v", ch.Index, err))
			u.AbortMultipartUpload(s3Key, uploadID)
			return fmt.Errorf("failed to upload part %d after retries: %w", ch.Index, err)
		}

		// Record completed part and save status.
		if uploadOutput.ETag != nil {
			status.AddCompletedPart(ch.Index, *uploadOutput.ETag)
			if err := status.SaveStatus(statusFilePath); err != nil {
				utils.Error(fmt.Sprintf("Failed to save status after completing part %d: %v", ch.Index, err))
				// Log the error but continue, as status save failure is non-fatal for current upload
			}
			completedParts = append(completedParts, &s3.CompletedPart{
				PartNumber: aws.Int64(int64(ch.Index)),
				ETag:       uploadOutput.ETag,
			})
			utils.Info(fmt.Sprintf("Successfully uploaded part %d. ETag: %s", ch.Index, *uploadOutput.ETag))
		} else {
			utils.Error(fmt.Sprintf("ETag for part %d is nil. Aborting upload.", ch.Index))
			u.AbortMultipartUpload(s3Key, uploadID)
			return fmt.Errorf("ETag for part %d is nil", ch.Index)
		}
	}

	// Complete the multipart upload in S3.
	utils.Info(fmt.Sprintf("Completing multipart upload for file: %s", filePath))
	_, err = u.s3Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(u.Config.S3BucketName),
		Key:      aws.String(s3Key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to complete multipart upload: %v", err))
		u.AbortMultipartUpload(s3Key, uploadID) // Abort if completion fails to clean up S3 resources
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	utils.Info(fmt.Sprintf("Multipart upload completed successfully for %s", filePath))

	// Remove the temporary status file.
	if err := os.Remove(statusFilePath); err != nil {
		utils.Error(fmt.Sprintf("Failed to remove status file %s: %v", statusFilePath, err))
	}

	return nil
}

// DeleteFile deletes a specific object from the configured S3 bucket.
func (u *Uploader) DeleteFile(s3Key string) error { // Consistent receiver type: S3Uploader -> Uploader
	utils.Info(fmt.Sprintf("Deleting file s3://%s/%s", u.Config.S3BucketName, s3Key))
	_, err := u.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(u.Config.S3BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to delete file %s from S3: %v", s3Key, err))
		return fmt.Errorf("failed to delete file %s from S3: %w", s3Key, err)
	}
	utils.Info(fmt.Sprintf("Successfully deleted file s3://%s/%s", u.Config.S3BucketName, s3Key))
	return nil
}

// AbortMultipartUpload aborts an ongoing multipart upload in S3.
// This is crucial for cleaning up incomplete uploads.
func (u *Uploader) AbortMultipartUpload(s3Key, uploadID string) error {
	utils.Info(fmt.Sprintf("Aborting multipart upload for key: %s, UploadID: %s", s3Key, uploadID))
	_, err := u.s3Client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
		Bucket:   aws.String(u.Config.S3BucketName),
		Key:      aws.String(s3Key),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to abort multipart upload for key %s, UploadID %s: %v", s3Key, uploadID, err))
		return fmt.Errorf("failed to abort multipart upload: %w", err)
	}
	utils.Info(fmt.Sprintf("Multipart upload aborted successfully for key: %s, UploadID: %s", s3Key, uploadID))
	return nil
}

// ListMultipartUploads lists all ongoing (incomplete) multipart uploads for the configured S3 bucket.
func (u *Uploader) ListMultipartUploads() ([]*s3.MultipartUpload, error) {
	utils.Info(fmt.Sprintf("Listing ongoing multipart uploads for bucket: %s", u.Config.S3BucketName))
	output, err := u.s3Client.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: aws.String(u.Config.S3BucketName),
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to list multipart uploads: %v", err))
		return nil, fmt.Errorf("failed to list multipart uploads: %w", err)
	}
	return output.Uploads, nil
}
