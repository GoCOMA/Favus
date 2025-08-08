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

var (
	resumeFilePath string
	resumeBucket   string
	resumeKey      string
	uploadID       string
)

func promptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume an interrupted multipart upload to S3",
	Long: `Resume an S3 multipart upload using a previously initiated Upload ID.
Use this command when an upload was interrupted and you want to continue from where it left off.`,
	Example: `
  favus resume --file ./video.mp4 --bucket my-bucket --key uploads/video.mp4 --upload-id ABC123XYZ
  favus resume --file ./video.mp4 --upload-id ABC123XYZ --config config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// ✅ 1. 인증 먼저
		cfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return err
		}

		// ✅ 2. config.yaml 우선 적용
		conf := GetLoadedConfig()
		if resumeBucket == "" && conf != nil {
			resumeBucket = conf.Bucket
		}
		if resumeKey == "" && conf != nil {
			resumeKey = conf.Key
		}
		if uploadID == "" && conf != nil {
			uploadID = conf.UploadID
		}

		// ✅ 3. 누락된 값만 프롬프트
		if resumeBucket == "" {
			resumeBucket = promptInput("🔧 Enter S3 bucket name")
		}
		if resumeKey == "" {
			resumeKey = promptInput("📝 Enter S3 object key")
		}
		if uploadID == "" {
			uploadID = promptInput("🔁 Enter Upload ID")
		}

		// ✅ 4. 로컬 파일 체크
		if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
			return fmt.Errorf("❌ file not found: %s", resumeFilePath)
		}

		// ✅ 5. 최종 정보 출력
		fmt.Println("✅ Final values:")
		fmt.Printf("File     : %s\n", resumeFilePath)
		fmt.Printf("Bucket   : %s\n", resumeBucket)
		fmt.Printf("Key      : %s\n", resumeKey)
		fmt.Printf("UploadID : %s\n", uploadID)

		// ✅ 6. 업로드 재개 (모의)
		s3Client := s3.NewFromConfig(cfg)
		_ = s3Client
		fmt.Println("🔄 Resuming upload (mock)...")
		fmt.Println("✅ Resume completed (mock)")
		return nil
	},
}

func init() {
	resumeCmd.Flags().StringVarP(&resumeFilePath, "file", "f", "", "Path to local file (required)")
	resumeCmd.Flags().StringVarP(&resumeBucket, "bucket", "b", "", "S3 bucket name")
	resumeCmd.Flags().StringVarP(&resumeKey, "key", "k", "", "S3 object key")
	resumeCmd.Flags().StringVarP(&uploadID, "upload-id", "u", "", "Upload ID to resume (required)")

	resumeCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(resumeCmd)
}
