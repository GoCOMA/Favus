package uploader

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CheckS3BucketAccess checks if a given S3 bucket exists and is accessible.
func CheckS3BucketAccess(ctx context.Context, cfg aws.Config, bucket string) error {
	client := s3.NewFromConfig(cfg)

	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("⚠️ bucket '%s' exists but access is denied", bucket)
		}
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return fmt.Errorf("❌ bucket '%s' does not exist", bucket)
		}
		return fmt.Errorf("❌ unexpected error checking bucket: %w", err)
	}

	fmt.Printf("✅ bucket '%s' is accessible\n", bucket)
	return nil
}
