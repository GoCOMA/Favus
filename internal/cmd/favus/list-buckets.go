package favus

import (
	"context"
	"fmt"
	"time"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var listBucketsCmd = &cobra.Command{
	Use:   "list-buckets",
	Short: "List all S3 buckets in the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1) AWS config (LocalStack or real AWS)
		awsCfg, err := awsutils.LoadAWSConfig(profile)
		if err != nil {
			return fmt.Errorf("load aws config: %w", err)
		}

		// 2) Create S3 client
		s3Client := s3.NewFromConfig(awsCfg)

		// 3) List buckets
		result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{}) 
		if err != nil {
			return fmt.Errorf("could not list buckets: %w", err)
		}

		if len(result.Buckets) == 0 {
			fmt.Println("No buckets found in the account.")
			return nil
		}

		fmt.Println("Buckets:")
		for _, bucket := range result.Buckets {
			bucketName := ""
			if bucket.Name != nil {
				bucketName = *bucket.Name
			}
			
			creationDate := "N/A"
			if bucket.CreationDate != nil {
				creationDate = bucket.CreationDate.Format(time.RFC3339)
			}
			
			fmt.Printf("- %s (Created: %s)\n", bucketName, creationDate)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listBucketsCmd)
}
