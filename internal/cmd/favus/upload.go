package favus

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CLI flags
var (
	filePath       string
	bucket         string
	objectKey      string
	uploadCompress bool
	autoTune       bool
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

func runUpload(cmd *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(bucket, objectKey, "")
	if err != nil {
		return err
	}

	// Prompt for missing required fields
	validator := NewConfigValidator(conf).RequireBucket().RequireKey()
	PromptForMissingConfig(validator)

	// Validate local file early so we don't measure RTT needlessly
	if err := ValidateFile(filePath); err != nil {
		return err
	}

	// Create uploader (needed before prompts when --auto-tune is used)
	up, err := CreateUploaderWithAWS(conf)
	if err != nil {
		return err
	}

	// Determine default part size and concurrency
	defaultPartSize := conf.PartSizeMB
	if defaultPartSize < MinPartSizeMB {
		defaultPartSize = MinPartSizeMB
	}
	defaultConcurrency := conf.MaxConcurrency
	if defaultConcurrency < MinConcurrency {
		defaultConcurrency = MinConcurrency
	}

	if autoTune {
		fmt.Println("🔍 Measuring RTT to S3...")
		profile, err := up.MeasureNetwork(cmd.Context())
		if err != nil {
			fmt.Printf("⚠️  AutoTune failed (%v), using config defaults\n", err)
		} else {
			fmt.Printf("🔍 %s (RTT %.1fms) — recommended: %d workers, %dMB parts\n",
				profile.Label, float64(profile.RTT.Milliseconds()), profile.Workers, profile.PartSizeMB)
			defaultPartSize = profile.PartSizeMB
			defaultConcurrency = profile.Workers
		}
	}

	conf.PartSizeMB = PromptIntWithValidation("📦 Enter part size in MB", defaultPartSize, MinPartSizeMB)
	conf.MaxConcurrency = PromptIntWithValidation("🔁 Enter max concurrency", defaultConcurrency, MinConcurrency)

	// Compression prompt (unless explicitly set via flag)
	if cmd.Flags().Changed("compress") {
		conf.Compress = uploadCompress
	} else {
		conf.Compress = PromptYesNoDefault("🗜  압축해서 업로드할까요?", conf.Compress)
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
	uploadCmd.Flags().BoolVar(&uploadCompress, "compress", false, "Compress the file with gzip before uploading")
	uploadCmd.Flags().Lookup("compress").NoOptDefVal = "true"
	uploadCmd.Flags().BoolVar(&autoTune, "auto-tune", false, "Measure RTT to S3 and automatically set workers and part size")
	_ = uploadCmd.MarkFlagRequired("file")
}
