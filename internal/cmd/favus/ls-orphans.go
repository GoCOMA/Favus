package favus

import (
	"fmt"
	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
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
		cfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return err
		}

		// 3. S3 Client 생성 및 로직 실행
		s3Client := s3.NewFromConfig(cfg)
		_ = s3Client // TODO: ListMultipartUploads 로직 구현

		fmt.Println("🔍 Scanning for incomplete uploads in:", targetBucket)
		fmt.Println("✅ Found 0 orphan uploads (mock)")
		return nil
	},
}

func init() {
	lsOrphansCmd.Flags().StringVarP(&lsOrphansBucket, "bucket", "b", "", "Target S3 bucket name")
	rootCmd.AddCommand(lsOrphansCmd)
}
