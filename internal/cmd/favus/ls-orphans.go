package favus

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var (
	lsOrphansBucket string
	lsOrphansRegion string
)

var lsOrphansCmd = &cobra.Command{
	Use:   "ls-orphans",
	Short: "List incomplete multipart uploads in a bucket",
	Long: `Scans the specified S3 bucket for incomplete multipart uploads
that may be wasting storage space and prints their metadata.`,
	Example: `
  favus ls-orphans --bucket my-bucket
  favus ls-orphans --config config.yaml`,
	RunE: runLsOrphans,
}

func formatUploadInfo(up *types.MultipartUpload) (string, string, string, string, string) {
	key := StringPtrValue(up.Key)
	uid := StringPtrValue(up.UploadId)

	initiated := "-"
	if up.Initiated != nil {
		initiated = up.Initiated.UTC().Format(time.RFC3339)
	}

	initiator := "-"
	if up.Initiator != nil {
		if up.Initiator.DisplayName != nil && *up.Initiator.DisplayName != "" {
			initiator = StringPtrValue(up.Initiator.DisplayName)
		} else if up.Initiator.ID != nil && *up.Initiator.ID != "" {
			initiator = StringPtrValue(up.Initiator.ID)
		}
	}

	storageClass := string(up.StorageClass)
	if storageClass == "" {
		storageClass = "-"
	}

	return uid, key, initiated, initiator, storageClass
}

func createS3ClientWithPathStyle(awsCfg aws.Config) *s3.Client {
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if os.Getenv("AWS_ENDPOINT_URL") != "" {
			o.UsePathStyle = true
		}
	})
}

func runLsOrphans(_ *cobra.Command, _ []string) error {
	// Load and validate config
	conf, err := LoadConfigWithOverrides(lsOrphansBucket, "", lsOrphansRegion)
	if err != nil {
		return err
	}

	// Validate required bucket
	if conf.Bucket == "" {
		return fmt.Errorf("S3 bucket name is required")
	}

	// Setup AWS config and S3 client
	awsCfg, err := awsutils.LoadAWSConfig(profile)
	if err != nil {
		return err
	}

	s3Client := createS3ClientWithPathStyle(awsCfg)

	// Scan for incomplete uploads
	fmt.Println("🔍 Scanning for incomplete uploads in:", conf.Bucket)

	ctx := context.Background()
	var (
		keyMarker      *string
		uploadIDMarker *string
		totalFound     int
	)

	for {
		out, err := s3Client.ListMultipartUploads(ctx, &s3.ListMultipartUploadsInput{
			Bucket:         ToStringPtr(conf.Bucket),
			KeyMarker:      keyMarker,
			UploadIdMarker: uploadIDMarker,
			MaxUploads:     aws.Int32(1000),
		})
		if err != nil {
			return fmt.Errorf("list multipart uploads: %w", err)
		}

		for _, up := range out.Uploads {
			uid, key, initiated, initiator, storageClass := formatUploadInfo(&up)
			fmt.Printf("- UploadID: %s | Key: %s | Initiated: %s | Initiator: %s | StorageClass: %s\n",
				uid, key, initiated, initiator, storageClass)
			totalFound++
		}

		// Handle pagination
		if out.IsTruncated != nil && *out.IsTruncated {
			keyMarker = out.NextKeyMarker
			uploadIDMarker = out.NextUploadIdMarker
			continue
		}
		break
	}

	if totalFound == 0 {
		fmt.Println("✅ Found 0 orphan uploads")
	} else {
		fmt.Printf("✅ Found %d incomplete multipart upload(s)\n", totalFound)
	}
	return nil
}

func init() {
	lsOrphansCmd.Flags().StringVarP(&lsOrphansBucket, "bucket", "b", "", "Target S3 bucket name")
	rootCmd.AddCommand(lsOrphansCmd)
}
