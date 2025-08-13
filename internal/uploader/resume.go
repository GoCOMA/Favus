package uploader

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/GoCOMA/Favus/internal/chunker"
	"github.com/GoCOMA/Favus/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/schollz/progressbar/v3"
)

// ResumeUploader allows resuming a multipart upload (AWS SDK v2).
type ResumeUploader struct {
	S3Client *s3.Client
}

// NewResumeUploader creates a new ResumeUploader.
func NewResumeUploader(s3Client *s3.Client) *ResumeUploader {
	return &ResumeUploader{S3Client: s3Client}
}

// ResumeUpload resumes a multipart upload from a saved status.
func (ru *ResumeUploader) ResumeUpload(statusFilePath string) error {
	status, err := LoadStatus(statusFilePath)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to load upload status for resume from %s: %v", statusFilePath, err))
		return fmt.Errorf("failed to load upload status for resume: %w", err)
	}

	utils.Info(fmt.Sprintf("Resuming upload for file: %s with UploadID: %s", status.FilePath, status.UploadID))

	fileChunker, err := chunker.NewFileChunker(status.FilePath, status.PartSizeBytes)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create file chunker for resume for %s: %v", status.FilePath, err))
		return fmt.Errorf("failed to create file chunker for resume: %w", err)
	}
	chunks := fileChunker.Chunks()

	// Ensure the total parts match
	if len(chunks) != status.TotalParts {
		utils.Error(fmt.Sprintf(
			"Mismatch in total parts for %s: expected %d, got %d from status. Aborting resume.",
			status.FilePath, len(chunks), status.TotalParts,
		))
		return fmt.Errorf("mismatch in total parts: expected %d, got %d from status", len(chunks), status.TotalParts)
	}

	completedParts := make([]s3types.CompletedPart, 0, len(status.CompletedParts))
	for partNum, eTag := range status.CompletedParts {
		completedParts = append(completedParts, s3types.CompletedPart{
			PartNumber: aws.Int32(int32(partNum)), // *int32
			ETag:       aws.String(eTag),          // *string
		})
	}

	// Sort completed parts by part number to ensure correct order
	sort.Slice(completedParts, func(i, j int) bool {
		return aws.ToInt32(completedParts[i].PartNumber) < aws.ToInt32(completedParts[j].PartNumber)
	})

	// Upload remaining parts
	// === 진행률 바 설정: 총 바이트 기준 ===
	fi, _ := os.Stat(status.FilePath)
	totalBar := progressbar.NewOptions64(
		fi.Size(),
		progressbar.OptionSetDescription("total"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionClearOnFinish(),
	)
	// 이미 완료된 파트만큼 선가산
	var already int64
	for _, ch := range chunks {
		if status.IsPartCompleted(ch.Index) {
			already += ch.Size
		}
	}
	_ = totalBar.Add64(already)

	for _, ch := range chunks {
		if status.IsPartCompleted(ch.Index) {
			utils.Info(fmt.Sprintf("Part %d already completed, skipping.", ch.Index))
			continue
		}

		reader, err := fileChunker.GetChunkReader(ch) // implements io.ReadSeekCloser
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to get chunk reader for part %d of %s: %v", ch.Index, status.FilePath, err))
			return fmt.Errorf("failed to get chunk reader for part %d: %w", ch.Index, err)
		}

		utils.Info(fmt.Sprintf("Uploading part %d (offset %d, size %d) for file %s", ch.Index, ch.Offset, ch.Size, status.FilePath))

		//파트 진행률 바
		partBar := progressbar.NewOptions64(
			ch.Size,
			progressbar.OptionSetDescription(fmt.Sprintf("part %d", ch.Index)),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(30),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionClearOnFinish(),
		)
		// 진행률 래퍼로 감싸서 SDK가 읽는 만큼 바를 올림 (재시도/되감기 보정 포함)
		pr := NewReadSeekCloserProgress(reader, func(n int64) {
			_ = partBar.Add64(n)
			_ = totalBar.Add64(n)
		})
		utils.Info(fmt.Sprintf("Uploading part %d (offset %d, size %d) for file %s", ch.Index, ch.Offset, ch.Size, status.FilePath))

		var uploadOutput *s3.UploadPartOutput
		err = utils.Retry(5, 2*time.Second, func() error {
			var partErr error
			uploadOutput, partErr = ru.S3Client.UploadPart(context.Background(), &s3.UploadPartInput{
				Body:          pr, // Read+Seek
				Bucket:        &status.Bucket,
				Key:           &status.Key,
				PartNumber:    aws.Int32(int32(ch.Index)), // *int32
				UploadId:      &status.UploadID,
				ContentLength: aws.Int64(ch.Size), // *int64
			})
			if partErr != nil {
				utils.Error(fmt.Sprintf("Failed to upload part %d for %s: %v", ch.Index, status.FilePath, partErr))
				return partErr
			}
			return nil
		})
		_ = pr.Close()

		if err != nil {
			utils.Error(fmt.Sprintf("Failed to upload part %d for %s after retries: %v", ch.Index, status.FilePath, err))
			return fmt.Errorf("failed to upload part %d after retries: %w", ch.Index, err)
		}

		status.AddCompletedPart(ch.Index, *uploadOutput.ETag)
		if err := status.SaveStatus(statusFilePath); err != nil {
			utils.Error(fmt.Sprintf("Failed to save status after completing part %d for %s: %v", ch.Index, status.FilePath, err))
		}
		utils.Info(fmt.Sprintf("Successfully uploaded part %d. ETag: %s", ch.Index, *uploadOutput.ETag))

		completedParts = append(completedParts, s3types.CompletedPart{
			PartNumber: aws.Int32(int32(ch.Index)),
			ETag:       uploadOutput.ETag,
		})
	}

	// Complete 전에 파트 오름차순 정렬(완료 파트 + 새 파트 전체)
	sort.Slice(completedParts, func(i, j int) bool {
		return aws.ToInt32(completedParts[i].PartNumber) < aws.ToInt32(completedParts[j].PartNumber)
	})
	// Complete the multipart upload
	utils.Info(fmt.Sprintf("Completing multipart upload for file: %s", status.FilePath))
	_, err = ru.S3Client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   &status.Bucket,
		Key:      &status.Key,
		UploadId: &status.UploadID,
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to complete multipart upload for %s: %v", status.FilePath, err))
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	utils.Info(fmt.Sprintf("Multipart upload completed successfully for %s", status.FilePath))

	// Clean up status file
	if err := os.Remove(statusFilePath); err != nil {
		utils.Error(fmt.Sprintf("Failed to remove status file %s: %v", statusFilePath, err))
	}

	return nil
}
