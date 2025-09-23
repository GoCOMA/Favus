package favus

import (
	"fmt"

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
Handles chunking, retries, resume support, and progress visualization automatically.`,
	Example: `
  favus upload --file ./bigfile.mp4 --bucket my-bucket --key uploads/bigfile.mp4
  favus upload -f ./bigfile.mp4 -c config.yaml`,
	RunE: runUpload,
}

func runUpload(_ *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(bucket, objectKey, "")
	if err != nil {
		return err
	}

	// Prompt for missing required fields
	validator := NewConfigValidator(conf).RequireBucket().RequireKey()
	PromptForMissingConfig(validator)

	// Prompt for upload parameters with proper defaults
	defaultPartSize := conf.PartSizeMB
	if defaultPartSize < MinPartSizeMB {
		defaultPartSize = MinPartSizeMB
	}

	defaultConcurrency := conf.MaxConcurrency
	if defaultConcurrency < MinConcurrency {
		defaultConcurrency = MinConcurrency
	}

	conf.PartSizeMB = PromptIntWithValidation("📦 Enter part size in MB", defaultPartSize, MinPartSizeMB)
	conf.MaxConcurrency = PromptIntWithValidation("🔁 Enter max concurrency", defaultConcurrency, MinConcurrency)

	// Validate local file
	if err := ValidateFile(filePath); err != nil {
		return err
	}

	// Create uploader and perform upload
	up, err := CreateUploaderWithAWS(conf)
	if err != nil {
		return err
	}

	if err := up.UploadFile(filePath, conf.Key); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Println(FormatSuccessMessage("Upload complete", conf.Bucket, conf.Key))
	return nil
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the local file to upload (required)")
	uploadCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "Target S3 bucket name (overrides config/ENV)")
	uploadCmd.Flags().StringVarP(&objectKey, "key", "k", "", "S3 object key (overrides config/ENV)")
	_ = uploadCmd.MarkFlagRequired("file")
}
