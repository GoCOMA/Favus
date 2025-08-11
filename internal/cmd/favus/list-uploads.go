package favus

import (
	"fmt"
	"time"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/uploader"
	"github.com/spf13/cobra"
)

var (
	listBucket string // optional; overrides config/ENV
)

var listUploadsCmd = &cobra.Command{
	Use:   "list-uploads",
	Short: "List ongoing multipart uploads in the bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) Load effective config prepared by PersistentPreRunE
		conf := GetLoadedConfig()
		if conf == nil {
			return fmt.Errorf("config not loaded")
		}
		if listBucket != "" {
			conf.Bucket = listBucket
		}
		if conf.Bucket == "" {
			return fmt.Errorf("bucket is required (use --bucket or config/ENV)")
		}

		// 2) AWS config (LocalStack or real AWS)
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 3) Create uploader and list
		up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
		if err != nil {
			return fmt.Errorf("init uploader: %w", err)
		}

		items, err := up.ListMultipartUploads()
		if err != nil {
			return fmt.Errorf("list multipart uploads: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("No ongoing multipart uploads.")
			return nil
		}
		fmt.Printf("Ongoing multipart uploads in bucket %s:\n", conf.Bucket)
		for _, u := range items {
			// pointers may be nil; guard
			key := ""
			uploadID := ""
			if u.Key != nil {
				key = *u.Key
			}
			if u.UploadId != nil {
				uploadID = *u.UploadId
			}
			initTime := u.Initiated
			fmt.Printf("- UploadID: %s | Key: %s | Initiated: %s\n",
				uploadID, key, initTime.Format(time.RFC3339))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listUploadsCmd)
	listUploadsCmd.Flags().StringVar(&listBucket, "bucket", "", "S3 bucket to inspect (overrides config/ENV)")
}
