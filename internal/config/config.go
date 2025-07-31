package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

const DefaultChunkSize = 5 * 1024 * 1024

type Config struct {
	AWSRegion string
	S3BucketName string
	ChunkSize int64
}

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

func LoadConfig() (*Config, error) {
	creds, err := LoadAWSCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS credentials: %w", err)
	}

	chunkSizeStr := os.Getenv("CHUNK_SIZE")
	
	// If CHUNK_SIZE is set, parse it; otherwise, use the default
	if chunkSizeStr != "" {
		parsedSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
		if err != nil {
			fmt.Printf("Warning: CHUNK_SIZE environment variable '%s' is not a valid number. Using default chunk size (%d bytes).\n", chunkSizeStr, DefaultChunkSize)
		} else if parsedSize <= 0 {
			fmt.Printf("Warning: CHUNK_SIZE environment variable '%s' must be greater than 0. Using default chunk size (%d bytes).\n", chunkSizeStr, DefaultChunkSize)
		} else {
			DefaultChunkSize = parsedSize
		}
	}

	// Check if S3_BUCKET_NAME is set in the environment variables
	S3BucketName := os.Getenv("S3_BUCKET_NAME")
	if S3BucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	// Default configuration
	config := &Config{
		AWSRegion: "ap-northeast-2", // Default to Seoul region
		S3BucketName: S3BucketName,
		ChunkSize: DefaultChunkSize,
	}

	fmt.Printf("Using AWS Region: %s\n", config.AWSRegion)
	fmt.Printf("Using S3 Bucket: %s\n", config.S3BucketName)
	fmt.Printf("Using Chunk Size: %d bytes\n", config.ChunkSize)
	
	return config, nil
}