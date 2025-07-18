package favus

import (
	"fmt"
	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"

	"github.com/spf13/cobra"
)

var (
	resumeFilePath string
	resumeBucket   string
	resumeKey      string
	uploadID       string
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume an interrupted multipart upload to S3",
	Long: `Favus resume allows you to continue a previously interrupted multipart upload using an upload ID.
It checks which parts have already been uploaded and continues the rest.`,
	Example: `
  favus resume --file ./video.mp4 --bucket my-bucket --key uploads/video.mp4 --upload-id xyz123`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat(resumeFilePath); os.IsNotExist(err) {
			fmt.Printf(" File not found: %s\n", resumeFilePath)
			return
		}

		cfg, err := awsutils.LoadAWSConfig()
		if err != nil {
			fmt.Println("AWS credential error:", err)
			return
		}
		s3Client := s3.NewFromConfig(cfg)
		_ = s3Client //임시로 이렇게 처리해둠. 밑에 로직 성공하면 지우자. (선언만하고 쓰이는데없어서 에러남)

		fmt.Println("🔄 Resuming upload...")
		fmt.Printf("File: %s\nBucket: %s\nKey: %s\nUploadID: %s\n", resumeFilePath, resumeBucket, resumeKey, uploadID)

		// TODO: Call resume logic with s3Client
		// e.g., uploader.ResumeUpload(s3Client, resumeFilePath, resumeBucket, resumeKey, uploadID)

		fmt.Println("✅ Resume completed (mock)")
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)

	resumeCmd.Flags().StringVarP(&resumeFilePath, "file", "f", "", "Path to local file")
	resumeCmd.Flags().StringVarP(&resumeBucket, "bucket", "b", "", "S3 bucket name")
	resumeCmd.Flags().StringVarP(&resumeKey, "key", "k", "", "S3 object key")
	resumeCmd.Flags().StringVarP(&uploadID, "upload-id", "u", "", "Upload ID to resume")

	_ = resumeCmd.MarkFlagRequired("file")
	_ = resumeCmd.MarkFlagRequired("bucket")
	_ = resumeCmd.MarkFlagRequired("key")
	_ = resumeCmd.MarkFlagRequired("upload-id")
}
