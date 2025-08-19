package uploader

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/GoCOMA/Favus/internal/chunker"
	"github.com/GoCOMA/Favus/internal/wsagent"
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

	// === [추가] 서버 상태와 동기화(ListParts) ===
	{
		ctx := context.Background()
		srvCompleted, err := ru.fetchServerCompletedParts(ctx, status.Bucket, status.Key, status.UploadID)
		if err != nil {
			utils.Error(fmt.Sprintf("ListParts failed for %s/%s (UploadID=%s): %v", status.Bucket, status.Key, status.UploadID, err))
			return fmt.Errorf("list parts: %w", err)
		}
		for pn, et := range srvCompleted {
			status.AddCompletedPart(pn, et) // 서버가 진실원천
		}
		if err := status.SaveStatus(statusFilePath); err != nil {
			utils.Error(fmt.Sprintf("Failed to save status after server sync: %v", err))
			return fmt.Errorf("save status after sync: %w", err)
		}
	}
	// === [추가 끝] ===

	fileChunker, err := chunker.NewFileChunker(status.FilePath, status.PartSizeBytes)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create file chunker for resume for %s: %v", status.FilePath, err))
		return fmt.Errorf("failed to create file chunker for resume: %w", err)
	}
	chunks := fileChunker.Chunks()

	// 총 파트 개수 일치 여부 확인
	if len(chunks) != status.TotalParts {
		utils.Error(fmt.Sprintf(
			"Mismatch in total parts for %s: expected %d, got %d from status. Aborting resume.",
			status.FilePath, len(chunks), status.TotalParts,
		))
		return fmt.Errorf("mismatch in total parts: expected %d, got %d from status", len(chunks), status.TotalParts)
	}

	// 이미 완료된 파트 정보를 CompletedPart 형식으로 수집 (동기화 결과 반영됨)
	completedParts := make([]s3types.CompletedPart, 0, len(status.CompletedParts))
	for partNum, eTag := range status.CompletedParts {
		completedParts = append(completedParts, s3types.CompletedPart{
			PartNumber: aws.Int32(int32(partNum)),
			ETag:       aws.String(eTag),
		})
	}
	// 정렬(안전)
	sort.Slice(completedParts, func(i, j int) bool {
		return aws.ToInt32(completedParts[i].PartNumber) < aws.ToInt32(completedParts[j].PartNumber)
	})

	// === 진행률 바 설정: 총 바이트 기준 ===
	fi, _ := os.Stat(status.FilePath)
	totalBar := progressbar.NewOptions64(
		fi.Size(),
		progressbar.OptionSetDescription("total"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		//progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWriter(os.Stdout),
	)

	// 이미 완료된 바이트 합산
	var already int64
	for _, ch := range chunks {
		if status.IsPartCompleted(ch.Index) {
			already += ch.Size
		}
	}
	_ = totalBar.Add64(already)

	// === WS Reporter: 세션 시작(Resumed) ===
	r := newWSReporter(fi.Size())
	// UI 초기화용 preCompleted 목록 구성(파트/크기/etag)
	preCompleted := make([]map[string]any, 0, len(completedParts))
	for _, cp := range completedParts {
		part := int(aws.ToInt32(cp.PartNumber))
		var sz int64
		// 파트 크기 추정: chunk 목록에 동일 파트 번호 존재
		if part >= 1 && part <= len(chunks) {
			sz = chunks[part-1].Size
		}
		preCompleted = append(preCompleted, map[string]any{
			"part": part,
			"size": sz,
			"etag": aws.ToString(cp.ETag),
		})
	}
	r.start(status.Bucket, status.Key, status.UploadID, status.PartSizeBytes, map[string]any{
		"resumed":       true,
		"alreadyBytes":  already,
		"preCompleted":  preCompleted,
		"totalParts":    status.TotalParts,
		"partSizeBytes": status.PartSizeBytes,
	})
	// 진행률 기준을 맞추기 위해 내부 누적값 초기화
	r.uploadedBytes = already // 같은 패키지이므로 필드 접근 가능

	// === 남은 파트 업로드 ===
	for _, ch := range chunks {
		if status.IsPartCompleted(ch.Index) {
			utils.Info(fmt.Sprintf("Part %d already completed, skipping.", ch.Index))
			continue
		}

		reader, err := fileChunker.GetChunkReader(ch) // io.ReadSeekCloser
		if err != nil {
			utils.Error(fmt.Sprintf("Failed to get chunk reader for part %d of %s: %v", ch.Index, status.FilePath, err))
			r.error(fmt.Sprintf("get chunk reader (part %d): %v", ch.Index, err), &ch.Index)
			r.done(false, status.UploadID)
			return fmt.Errorf("failed to get chunk reader for part %d: %w", ch.Index, err)
		}

		// WS: 파트 시작
		r.partStart(ch.Index, ch.Size, ch.Offset)

		// 파트 진행률 바
		partBar := progressbar.NewOptions64(
			ch.Size,
			progressbar.OptionSetDescription(fmt.Sprintf("part %d", ch.Index)),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetWidth(30),
			progressbar.OptionThrottle(65*time.Millisecond),
			//progressbar.OptionClearOnFinish(),
			progressbar.OptionSetWriter(os.Stdout),
		)

		// 진행률 래퍼 (되감기/재시도 고려)
		pr := NewReadSeekCloserProgress(reader, func(n int64) {
			_ = partBar.Add64(n)
			_ = totalBar.Add64(n)
			r.progressAdd(n)
			r.partProgressAdd(ch.Index, n)

			// wsagent 이벤트도 여기서
			ev := wsagent.Event{
				Type:      "progress",
				RunID:     r.runID,
				Timestamp: time.Now(),
				Payload:   []byte(fmt.Sprintf(`{"bytes":%d}`, n)),
			}
			_ = wsagent.SendEvent(context.Background(), wsagent.DefaultAddr(), ev)
		})

		utils.Info(fmt.Sprintf("Uploading part %d (offset %d, size %d) for file %s",
			ch.Index, ch.Offset, ch.Size, status.FilePath))

		var uploadOutput *s3.UploadPartOutput
		err = utils.Retry(5, 2*time.Second, func() error {
			var partErr error
			uploadOutput, partErr = ru.S3Client.UploadPart(context.Background(), &s3.UploadPartInput{
				Body:          pr,
				Bucket:        &status.Bucket,
				Key:           &status.Key,
				PartNumber:    aws.Int32(int32(ch.Index)),
				UploadId:      &status.UploadID,
				ContentLength: aws.Int64(ch.Size),
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
			r.error(fmt.Sprintf("upload part %d failed after retries: %v", ch.Index, err), &ch.Index)
			r.done(false, status.UploadID)
			return fmt.Errorf("failed to upload part %d after retries: %w", ch.Index, err)
		}

		if uploadOutput.ETag == nil {
			utils.Error(fmt.Sprintf("ETag for part %d is nil. Aborting resume.", ch.Index))
			r.error(fmt.Sprintf("nil ETag on part %d", ch.Index), &ch.Index)
			r.done(false, status.UploadID)
			return fmt.Errorf("ETag for part %d is nil", ch.Index)
		}

		// 상태 저장 + 누적 파트 목록 갱신
		status.AddCompletedPart(ch.Index, *uploadOutput.ETag)
		if err := status.SaveStatus(statusFilePath); err != nil {
			utils.Error(fmt.Sprintf("Failed to save status after completing part %d for %s: %v", ch.Index, status.FilePath, err))
		}
		_ = partBar.Finish()
		utils.Info(fmt.Sprintf("Successfully uploaded part %d. ETag: %s", ch.Index, *uploadOutput.ETag))

		completedParts = append(completedParts, s3types.CompletedPart{
			PartNumber: aws.Int32(int32(ch.Index)),
			ETag:       uploadOutput.ETag,
		})

		// WS: 파트 완료
		r.partDone(ch.Index, ch.Size, *uploadOutput.ETag)
	}

	// Complete 전에 정렬(안전)
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
		r.error(fmt.Sprintf("complete multipart: %v", err), nil)
		r.done(false, status.UploadID)
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	utils.Info(fmt.Sprintf("Multipart upload completed successfully for %s", status.FilePath))
	r.done(true, status.UploadID)

	// Clean up status file
	if err := os.Remove(statusFilePath); err != nil {
		utils.Error(fmt.Sprintf("Failed to remove status file %s: %v", statusFilePath, err))
	}

	return nil
}

// fetchServerCompletedParts lists completed parts on S3 and returns a map[partNumber]ETag.
// It handles pagination via PartNumberMarker/NextPartNumberMarker.
func (ru *ResumeUploader) fetchServerCompletedParts(
	ctx context.Context, bucket, key, uploadID string,
) (map[int]string, error) {
	result := make(map[int]string)

	var partMarkerStr *string

	for {
		in := &s3.ListPartsInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String(key),
			UploadId: aws.String(uploadID),
			MaxParts: aws.Int32(1000),
		}
		if partMarkerStr != nil {
			in.PartNumberMarker = partMarkerStr
		}

		out, err := ru.S3Client.ListParts(ctx, in)
		if err != nil {
			return nil, err
		}

		for _, p := range out.Parts {
			pn := int(aws.ToInt32(p.PartNumber))
			if p.ETag != nil {
				result[pn] = aws.ToString(p.ETag) // 따옴표 포함 그대로 유지
			}
		}

		// 페이징
		if out.IsTruncated != nil && *out.IsTruncated {
			if out.NextPartNumberMarker != nil && aws.ToString(out.NextPartNumberMarker) != "" {
				// 그대로 다음 마커로 사용
				partMarkerStr = out.NextPartNumberMarker
				continue
			}
			// 혹시 널/빈 문자열이면 마지막 파트로부터 유추
			if n := len(out.Parts); n > 0 {
				last := int(aws.ToInt32(out.Parts[n-1].PartNumber))
				s := strconv.Itoa(last)
				partMarkerStr = aws.String(s)
				continue
			}
		}
		break
	}
	return result, nil
}
