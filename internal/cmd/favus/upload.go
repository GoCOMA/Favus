package favus

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/uploader"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Load AWS config (profile aware)
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 2) Start from loaded config (file/ENV), then override with flags
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded (PersistentPreRunE should have populated it)")
		}
		if bucket != "" {
			conf.Bucket = strings.TrimSpace(bucket)
		}
		if objectKey != "" {
			conf.Key = strings.TrimSpace(objectKey)
		}

		// 3) Prompt for missing required fields
		reader := bufio.NewReader(os.Stdin)
		if strings.TrimSpace(conf.Bucket) == "" {
			fmt.Print("🔧 Enter S3 bucket name: ")
			in, _ := reader.ReadString('\n')
			conf.Bucket = strings.TrimSpace(in)
		}
		if strings.TrimSpace(conf.Key) == "" {
			fmt.Print("📝 Enter S3 object key: ")
			in, _ := reader.ReadString('\n')
			conf.Key = strings.TrimSpace(in)
		}

		// 4) Validate local file
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}

		// 5) Do upload
		up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
		if err != nil {
			return fmt.Errorf("init uploader: %w", err)
		}
		if err := up.UploadFile(filePath, conf.Key); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}

		fmt.Printf("✅ Upload complete → s3://%s/%s\n", conf.Bucket, conf.Key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the local file to upload (required)")
	uploadCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "Target S3 bucket name (overrides config/ENV)")
	uploadCmd.Flags().StringVarP(&objectKey, "key", "k", "", "S3 object key (overrides config/ENV)")
	_ = uploadCmd.MarkFlagRequired("file")
}
