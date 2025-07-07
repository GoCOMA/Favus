package favus

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	orphansBucket string
)

var orphansCmd = &cobra.Command{
	Use:   "ls-orphans",
	Short: "List incomplete multipart uploads (orphaned) in a bucket",
	Long: `Scan an S3 bucket and list ongoing multipart uploads that were not completed.
These uploads may consume storage without being visible as regular objects.`,
	Example: `
  favus ls-orphans --bucket my-bucket`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Listing orphan multipart uploads in bucket: %s\n\n", orphansBucket)

		// TODO: 실제 S3 ListMultipartUploads API 호출로 대체
		mockOrphans := []struct {
			UploadID string
			Key      string
			Date     string
		}{
			{"abc123uploadid", "uploads/bigfile1.mp4", "2025-07-06 21:40:12"},
			{"def456uploadid", "uploads/video_chunk.mov", "2025-07-05 18:13:47"},
		}

		fmt.Println("UPLOAD ID        | KEY                      | INITIATED AT")
		fmt.Println("-----------------|--------------------------|------------------------")
		for _, orphan := range mockOrphans {
			fmt.Printf("%-17s| %-25s| %s\n", orphan.UploadID, orphan.Key, orphan.Date)
		}
	},
}

func init() {
	rootCmd.AddCommand(orphansCmd)

	orphansCmd.Flags().StringVarP(&orphansBucket, "bucket", "b", "", "Target S3 bucket name")
	_ = orphansCmd.MarkFlagRequired("bucket")
}
