package favus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GoCOMA/Favus/internal/wsagent"
	"github.com/spf13/cobra"
)

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

var (
	listBucket string // optional; overrides config/ENV
)

var listUploadsCmd = &cobra.Command{
	Use:   "list-uploads",
	Short: "List ongoing multipart uploads in the bucket",
	RunE:  runListUploads,
}

func sendUIEvent(ctx context.Context, bucket string, items []map[string]string) {
	addr := wsagent.DefaultAddr()
	_ = wsagent.SendEvent(ctx, addr, wsagent.Event{
		Type:      "list-uploads",
		RunID:     "",
		Timestamp: time.Now(),
		Payload: mustJSON(map[string]any{
			"bucket": bucket,
			"items":  items,
		}),
	})
}

func runListUploads(cmd *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(listBucket, "", "")
	if err != nil {
		return err
	}

	if conf.Bucket == "" {
		return fmt.Errorf("bucket is required (use --bucket or config/ENV)")
	}

	// Create uploader and list uploads
	up, err := CreateUploaderWithAWS(conf)
	if err != nil {
		return fmt.Errorf("init uploader: %w", err)
	}

	items, err := up.ListMultipartUploads()
	if err != nil {
		return fmt.Errorf("list multipart uploads: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("No ongoing multipart uploads.")
		sendUIEvent(context.Background(), conf.Bucket, []map[string]string{})
		return nil
	}

	fmt.Printf("Ongoing multipart uploads in bucket %s:\n", conf.Bucket)
	uiItems := make([]map[string]string, 0, len(items))

	for _, u := range items {
		// Extract upload information
		key, uploadID, initiated := "", "", "-"

		// Extract fields from MultipartUpload struct
		key = StringPtrValue(u.Key)
		uploadID = StringPtrValue(u.UploadId)
		if u.Initiated != nil {
			initiated = u.Initiated.Format(time.RFC3339)
		}

		// Console output
		fmt.Printf("- UploadID: %s | Key: %s | Initiated: %s\n", uploadID, key, initiated)

		// UI data
		uiItems = append(uiItems, map[string]string{
			"uploadId":  uploadID,
			"key":       key,
			"initiated": initiated,
		})
	}

	sendUIEvent(cmd.Context(), conf.Bucket, uiItems)
	return nil
}

func init() {
	rootCmd.AddCommand(listUploadsCmd)
	listUploadsCmd.Flags().StringVar(&listBucket, "bucket", "", "S3 bucket to inspect (overrides config/ENV)")
}
