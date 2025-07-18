package favus

import (
	"errors"
	"fmt"
	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"

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
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		if bucket == "" || objectKey == "" {
			return errors.New("both --bucket and --key must be provided")
		}

		cfg, err := awsutils.LoadAWSConfig()
		if err != nil {
			return err
		}
		s3Client := s3.NewFromConfig(cfg)
		_ = s3Client //임시로 이렇게 처리해둠. 밑에 로직 성공하면 지우자. (선언만하고 쓰이는데없어서 에러남)

		fmt.Println("📤 Starting upload...")
		fmt.Printf("File:   %s\nBucket: %s\nKey:    %s\n\n", filePath, bucket, objectKey)

		// TODO: Use s3Client to perform actual multipart upload
		// e.g., uploader.UploadFile(s3Client, filePath, bucket, objectKey)

		fmt.Println("✅ Upload completed successfully (mock)")
		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the local file to upload (required)")
	uploadCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "Target S3 bucket name (required)")
	uploadCmd.Flags().StringVarP(&objectKey, "key", "k", "", "S3 object key (required)")

	uploadCmd.MarkFlagRequired("file")
	uploadCmd.MarkFlagRequired("bucket")
	uploadCmd.MarkFlagRequired("key")

	rootCmd.AddCommand(uploadCmd)
}
