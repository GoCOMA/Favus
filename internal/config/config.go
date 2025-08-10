package config

import (
	"fmt"
	"os"
	"strconv"
)

var DefaultChunkSize int64 = 5 * 1024 * 1024

type Config struct {
	AWSRegion    string
	S3BucketName string
	ChunkSize    int64
}

func LoadConfig() (*Config, error) {
	// CHUNK_SIZE
	if chunkSizeStr := os.Getenv("CHUNK_SIZE"); chunkSizeStr != "" {
		if parsed, err := strconv.ParseInt(chunkSizeStr, 10, 64); err != nil || parsed <= 0 {
			fmt.Printf("Warning: invalid CHUNK_SIZE '%s'. Using default %d bytes.\n", chunkSizeStr, DefaultChunkSize)
		} else {
			DefaultChunkSize = parsed
		}
	}

	// S3_BUCKET_NAME (required)
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	// AWS_REGION (optional; default to ap-northeast-2)
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-2"
	}

	cfg := &Config{
		AWSRegion:    region,
		S3BucketName: bucket,
		ChunkSize:    DefaultChunkSize,
	}

	fmt.Printf("Using AWS Region: %s\n", cfg.AWSRegion)
	fmt.Printf("Using S3 Bucket: %s\n", cfg.S3BucketName)
	fmt.Printf("Using Chunk Size: %d bytes\n", cfg.ChunkSize)

	return cfg, nil
}
