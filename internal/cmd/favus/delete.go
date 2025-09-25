package favus

import (
	"fmt"

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
	RunE: runDelete,
}

func runDelete(_ *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(delBucket, delKey, "")
	if err != nil {
		return err
	}

	// Prompt for missing required fields
	validator := NewConfigValidator(conf).RequireBucket().RequireKey()
	PromptForMissingConfig(validator)

	// Create uploader and perform deletion
	up, err := CreateUploaderWithAWS(conf)
	if err != nil {
		return err
	}

	if err := up.DeleteFile(conf.Key); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Println(FormatSuccessMessage("Deleted", conf.Bucket, conf.Key))
	return nil
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&delBucket, "bucket", "", "S3 bucket (overrides config/ENV)")
	deleteCmd.Flags().StringVar(&delKey, "key", "", "S3 object key to delete (overrides config/ENV)")
}
