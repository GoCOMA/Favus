package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSCredentials loads AWS configuration using SDK v2.
// It automatically checks environment variables, shared credentials file, and more.
func LoadAWSCredentials(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("❌ failed to load AWS config: %w", err)
	}

	// Just for debugging / confirmation
	fmt.Println("✔ AWS config loaded (v2)")
	fmt.Println("AccessKeyID (from credentials):", cfg.Credentials)

	return cfg, nil
}
