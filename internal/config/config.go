package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

func LoadAWSCredentials() (*credentials.Credentials, error) {
	envCreds := credentials.NewEnvCredentials()
	_, err := envCreds.Get()
	if err == nil {
		fmt.Println("✔ AWS credentials from environment variables")
		return envCreds, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}
	credFile := filepath.Join(homeDir, ".aws", "credentials")
	fileCreds := credentials.NewSharedCredentials(credFile, "default")
	_, err = fileCreds.Get()
	if err == nil {
		fmt.Println("AWS credentials from shared credentials file")
		return fileCreds, nil
	}

	return nil, fmt.Errorf("AWS credentials not found (env or file)")
}
