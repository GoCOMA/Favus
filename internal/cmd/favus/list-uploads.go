package favus

import (
	"fmt"
	"time"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/uploader"
	"github.com/GoCOMA/Favus/internal/wsagent"
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

			// UI에도 전송
			wsagent.SendEvent(cmd.Context(), "list-uploads", map[string]any{
				"bucket": conf.Bucket,
				"items":  []map[string]string{},
			})
			return nil
		}

		fmt.Printf("Ongoing multipart uploads in bucket %s:\n", conf.Bucket)
		uiItems := []map[string]string{} // UI 전송용 데이터 모으기

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

			initiated := "-"
			if u.Initiated != nil {
				initiated = u.Initiated.Format(time.RFC3339)
			}

			// 콘솔 출력
			fmt.Printf("- UploadID: %s | Key: %s | Initiated: %s\n", uploadID, key, initiated)

			// UI 데이터 추가
			uiItems = append(uiItems, map[string]string{
				"uploadId":  uploadID,
				"key":       key,
				"initiated": initiated,
			})
		}

		// UI 브라우저로 전송
		wsagent.SendEvent(cmd.Context(), "list-uploads", map[string]any{
			"bucket": conf.Bucket,
			"items":  uiItems,
		})

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listUploadsCmd)
	listUploadsCmd.Flags().StringVar(&listBucket, "bucket", "", "S3 bucket to inspect (overrides config/ENV)")
}
