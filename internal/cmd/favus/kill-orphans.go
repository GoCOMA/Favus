package favus

import (
	"context"
	"fmt"
	"os"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var killBucket string

var killOrphansCmd = &cobra.Command{
	Use:   "kill-orphans",
	Short: "Abort ALL incomplete multipart uploads in a bucket",
	Long: `Scans the given S3 bucket and aborts every in-progress multipart upload.
This is destructive and may interrupt ongoing uploads.`,
	Example: `
  favus kill-orphans --bucket my-bucket`,
	RunE: runKillOrphans,
}

type AbortStats struct {
	Total   int
	Aborted int
	Failed  int
}

func (s AbortStats) HasFailures() bool {
	return s.Failed > 0
}

func (s AbortStats) Print() {
	if s.Total == 0 {
		fmt.Println("✅ 미완성 멀티파트 업로드가 없습니다.")
		return
	}
	fmt.Printf("완료: 대상 %d, 성공 %d, 실패 %d\n", s.Total, s.Aborted, s.Failed)
}

func abortSingleUpload(ctx context.Context, client *s3.Client, bucket string, key, uploadID *string) error {
	_, err := client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   ToStringPtr(bucket),
		Key:      key,
		UploadId: uploadID,
	})
	return err
}

func runKillOrphans(_ *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(killBucket, "", "")
	if err != nil {
		return err
	}

	// Prompt for bucket if missing
	validator := NewConfigValidator(conf).RequireBucket()
	PromptForMissingConfig(validator)

	// Setup AWS config and S3 client
	awsCfg, err := awsutils.LoadAWSConfig(profile)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if os.Getenv("AWS_ENDPOINT_URL") != "" {
			o.UsePathStyle = true
		}
	})

	fmt.Printf("🔍 Scanning bucket '%s' for incomplete multipart uploads...\n", conf.Bucket)

	// Paginate through and abort all incomplete uploads
	ctx := context.Background()
	paginator := s3.NewListMultipartUploadsPaginator(client, &s3.ListMultipartUploadsInput{
		Bucket:     ToStringPtr(conf.Bucket),
		MaxUploads: aws.Int32(1000),
	})

	stats := AbortStats{}
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list multipart uploads: %w", err)
		}

		for _, up := range out.Uploads {
			stats.Total++
			key := StringPtrValue(up.Key)
			uid := StringPtrValue(up.UploadId)

			if err := abortSingleUpload(ctx, client, conf.Bucket, up.Key, up.UploadId); err != nil {
				stats.Failed++
				fmt.Printf("❌ abort 실패: key=%s uploadId=%s err=%v\n", key, uid, err)
			} else {
				stats.Aborted++
				fmt.Printf("✅ abort 성공: key=%s uploadId=%s\n", key, uid)
			}
		}
	}

	stats.Print()
	if stats.HasFailures() {
		return fmt.Errorf("some uploads could not be aborted")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(killOrphansCmd)
	killOrphansCmd.Flags().StringVar(&killBucket, "bucket", "", "S3 bucket (overrides config/ENV)")
}
