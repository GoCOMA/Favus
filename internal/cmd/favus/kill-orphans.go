package favus

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) config/ENV → flag overlay
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded (PersistentPreRunE should have populated it)")
		}
		if killBucket != "" {
			conf.Bucket = strings.TrimSpace(killBucket)
		}
		if strings.TrimSpace(conf.Bucket) == "" {
			conf.Bucket = promptInput("🔧 Enter S3 bucket name")
		}

		// 2) AWS config
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 3) S3 client
		cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if os.Getenv("AWS_ENDPOINT_URL") != "" {
				o.UsePathStyle = true
			}
		})

		fmt.Printf("🔍 Scanning bucket '%s' for incomplete multipart uploads...\n", conf.Bucket)

		// 4) 페이지네이션으로 전체 Abort
		ctx := context.Background()
		p := s3.NewListMultipartUploadsPaginator(cli, &s3.ListMultipartUploadsInput{
			Bucket:     aws.String(conf.Bucket),
			MaxUploads: aws.Int32(1000),
		})

		total, aborted, failed := 0, 0, 0
		for p.HasMorePages() {
			out, err := p.NextPage(ctx)
			if err != nil {
				return fmt.Errorf("list multipart uploads: %w", err)
			}
			for _, up := range out.Uploads {
				total++
				key := aws.ToString(up.Key)
				uid := aws.ToString(up.UploadId)

				_, err := cli.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
					Bucket:   aws.String(conf.Bucket),
					Key:      up.Key,
					UploadId: up.UploadId,
				})
				if err != nil {
					failed++
					fmt.Printf("❌ abort 실패: key=%s uploadId=%s err=%v\n", key, uid, err)
				} else {
					aborted++
					fmt.Printf("✅ abort 성공: key=%s uploadId=%s\n", key, uid)
				}
			}
		}

		if total == 0 {
			fmt.Println("✅ 미완성 멀티파트 업로드가 없습니다.")
			return nil
		}

		fmt.Printf("완료: 대상 %d, 성공 %d, 실패 %d\n", total, aborted, failed)
		if failed > 0 {
			return fmt.Errorf("some uploads could not be aborted")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(killOrphansCmd)
	killOrphansCmd.Flags().StringVar(&killBucket, "bucket", "", "S3 bucket (overrides config/ENV)")
}
