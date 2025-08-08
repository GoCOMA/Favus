package favus

import (
	"bufio"
	"fmt"
	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CLI flags
var (
	filePath  string
	bucket    string
	objectKey string
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to S3 using multipart upload",
	Long: `Initiates a multipart upload for a large file and uploads all parts to the specified S3 bucket.
Handles chunking, retries, and upload tracking automatically.`,
	Example: `
  favus upload --file ./bigfile.mp4 --bucket my-bucket --key uploads/bigfile.mp4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. AWS 인증 (인증 없으면 내부에서 프롬프트)
		cfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return err
		}

		// 2. config.yaml 우선 적용 (있다면)
		conf := GetLoadedConfig()
		if bucket == "" && conf != nil {
			bucket = conf.Bucket
		}
		if objectKey == "" && conf != nil {
			objectKey = conf.Key
		}

		// 3. 누락된 값에 대해 프롬프트
		reader := bufio.NewReader(os.Stdin)
		if bucket == "" {
			fmt.Print("🔧 Enter S3 bucket name: ")
			input, _ := reader.ReadString('\n')
			bucket = strings.TrimSpace(input)
		}
		if objectKey == "" {
			fmt.Print("📝 Enter S3 object key: ")
			input, _ := reader.ReadString('\n')
			objectKey = strings.TrimSpace(input)
		}

		// 4. 파일 존재 여부 체크
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}

		// 5. 실행 로그
		fmt.Printf("✅ Final values → file: %s, bucket: %s, key: %s\n", filePath, bucket, objectKey)

		// 6. 업로드 로직 (mock)
		s3Client := s3.NewFromConfig(cfg)
		_ = s3Client
		fmt.Println("📤 Starting upload...")
		fmt.Println("✅ Upload completed successfully (mock)")
		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the local file to upload (required)")
	uploadCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "Target S3 bucket name (required)")
	uploadCmd.Flags().StringVarP(&objectKey, "key", "k", "", "S3 object key (required)")

	uploadCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(uploadCmd)
}
