package favus

import (
	"fmt"
	"strings"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/uploader"
	"github.com/spf13/cobra"
)

var (
	delBucket string
	delKey    string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an object from S3",
	Long:  "Deletes a single object from the configured S3 bucket.",
	Example: `
  favus delete --key uploads/video.mp4
  favus delete --bucket my-bucket --key uploads/video.mp4 -c config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) AWS config
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 2) Base config (file/ENV via PersistentPreRunE)
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded (PersistentPreRunE should have populated it)")
		}

		// 3) Overlay flags
		if delBucket != "" {
			conf.Bucket = strings.TrimSpace(delBucket)
		}
		if delKey != "" {
			conf.Key = strings.TrimSpace(delKey)
		}

		// 4) Prompt missing fields
		if strings.TrimSpace(conf.Bucket) == "" {
			conf.Bucket = promptInput("🔧 Enter S3 bucket name")
		}
		if strings.TrimSpace(conf.Key) == "" {
			conf.Key = promptInput("🗑️  Enter S3 object key to delete")
		}

		// 5) Execute deletion
		up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
		if err != nil {
			return fmt.Errorf("init uploader: %w", err)
		}
		if err := up.DeleteFile(conf.Key); err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}

		fmt.Printf("✅ Deleted s3://%s/%s\n", conf.Bucket, conf.Key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&delBucket, "bucket", "", "S3 bucket (overrides config/ENV)")
	deleteCmd.Flags().StringVar(&delKey, "key", "", "S3 object key to delete (overrides config/ENV)")
}
