package awsutils

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
)

// LoadAWSConfig loads AWS config and exits with message if credentials are missing
func LoadAWSConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		if isMissingCredentials(err) {
			fmt.Println("AWS credentials not found.")
			fmt.Println("Please run `aws configure` to set them up.")
			os.Exit(1)
		}
		return cfg, err
	}
	return cfg, nil
}

// isMissingCredentials checks error type
func isMissingCredentials(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "UnrecognizedClientException", "InvalidClientTokenId", "MissingAuthenticationToken":
			return true
		}
	}
	return false
}
