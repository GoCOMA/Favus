package favus

import (
	"errors"
	"fmt"
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

		fmt.Printf("Starting upload...\n")
		fmt.Printf("File:   %s\n", filePath)
		fmt.Printf("Bucket: %s\n", bucket)
		fmt.Printf("Key:    %s\n\n", objectKey)

		// TODO: Replace this with actual logic
		// chunks := chunker.SplitFile(filePath)
		// err := uploader.UploadFile(chunks, bucket, objectKey)

		// Simulate result for now
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
