package favus

import (
	"context"
	"fmt"
	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	lsOrphansBucket string
	lsOrphansRegion string
)

var lsOrphansCmd = &cobra.Command{
	Use:   "ls-orphans",
	Short: "List incomplete multipart uploads in a bucket",
	Long: `Scans the specified S3 bucket for incomplete multipart uploads
that may be wasting storage space and prints their metadata.`,
	Example: `
  favus ls-orphans --bucket my-bucket
  favus ls-orphans --config config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("failed to load config")
		}

		// 1. CLI 인자가 우선
		targetBucket := lsOrphansBucket
		if targetBucket == "" {
			targetBucket = conf.Bucket
		}
		if targetBucket == "" {
			return fmt.Errorf("S3 bucket name is required")
		}

		// 2. AWS 인증 및 region 설정
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return err
		}

		// 3. S3 Client 생성 및 로직 실행
		endpoint := os.Getenv("AWS_ENDPOINT_URL")
		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if endpoint != "" {
				o.UsePathStyle = true
			}
		})

		// 4) 페이지네이션으로 진행 중 멀티파트 업로드 나열
		fmt.Println("🔍 Scanning for incomplete uploads in:", targetBucket)

		ctx := context.Background()
		var (
			keyMarker      *string
			uploadIDMarker *string
			totalFound     int
		)

		for {
			out, err := s3Client.ListMultipartUploads(ctx, &s3.ListMultipartUploadsInput{
				Bucket:         &targetBucket,
				KeyMarker:      keyMarker,
				UploadIdMarker: uploadIDMarker,
				MaxUploads:     aws.Int32(1000),
			})
			if err != nil {
				return fmt.Errorf("list multipart uploads: %w", err)
			}

			for _, up := range out.Uploads {
				key := aws.ToString(up.Key)
				uid := aws.ToString(up.UploadId)

				initiated := "-"
				if up.Initiated != nil {
					initiated = up.Initiated.UTC().Format(time.RFC3339)
				}

				initiator := "-"
				if up.Initiator != nil {
					if up.Initiator.DisplayName != nil && *up.Initiator.DisplayName != "" {
						initiator = aws.ToString(up.Initiator.DisplayName)
					} else if up.Initiator.ID != nil && *up.Initiator.ID != "" {
						initiator = aws.ToString(up.Initiator.ID)
					}
				}

				storageClass := string(up.StorageClass)
				if storageClass == "" {
					storageClass = "-"
				}

				fmt.Printf("- UploadID: %s | Key: %s | Initiated: %s | Initiator: %s | StorageClass: %s\n",
					uid, key, initiated, initiator, storageClass)

				totalFound++
			}

			// 페이징
			if out.IsTruncated != nil && *out.IsTruncated {
				keyMarker = out.NextKeyMarker
				uploadIDMarker = out.NextUploadIdMarker
				continue
			}
			break
		}

		if totalFound == 0 {
			fmt.Println("✅ Found 0 orphan uploads")
		} else {
			fmt.Printf("✅ Found %d incomplete multipart upload(s)\n", totalFound)
		}
		return nil
	},
}

func init() {
	lsOrphansCmd.Flags().StringVarP(&lsOrphansBucket, "bucket", "b", "", "Target S3 bucket name")
	rootCmd.AddCommand(lsOrphansCmd)
}
