package duplicate

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GoCOMA/Favus/internal/config"
	"github.com/GoCOMA/Favus/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsv2cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/redis/go-redis/v9"
)

// DuplicateChecker handles duplicate file detection using Redis-based Cuckoo Filter and Count-Min Sketch
type DuplicateChecker struct {
	rdb      *redis.Client
	s3Client *s3.Client
}

// NewDuplicateChecker creates a new duplicate checker instance
func NewDuplicateChecker(cfg *config.Config) (*DuplicateChecker, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	rdb := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create S3 client
	s3Client, err := createS3Client(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	dc := &DuplicateChecker{
		rdb:      rdb,
		s3Client: s3Client,
	}

	// Initialize Redis data structures
	if err := dc.initializeRedisStructures(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis structures: %w", err)
	}

	return dc, nil
}

// initializeRedisStructures sets up Cuckoo Filter and Count-Min Sketch in Redis
func (dc *DuplicateChecker) initializeRedisStructures(ctx context.Context) error {
	// Initialize Cuckoo Filter for file hash tracking
	// Capacity: 1000000, Bucket size: 4, Max iterations: 20
	_, err := dc.rdb.Do(ctx, "CF.RESERVE", "file_hashes", 1000000).Result()
	if err != nil {
		errMsg := err.Error()
		// Ignore if key already exists or if item exists (both mean structure is already created)
		if errMsg != "ERR key already exists" && errMsg != "ERR item exists" {
			return fmt.Errorf("failed to create Cuckoo Filter: %w", err)
		}
		utils.Info("Cuckoo Filter already exists, reusing existing structure")
	} else {
		utils.Info("Cuckoo Filter created successfully")
	}

	// Initialize Count-Min Sketch for frequency tracking
	// Width: 1000, Depth: 10
	_, err = dc.rdb.Do(ctx, "CMS.INITBYDIM", "upload_frequency", 1000, 10).Result()
	if err != nil {
		errMsg := err.Error()
		// Check for various "already exists" error messages
		if errMsg != "ERR key already exists" && errMsg != "CMS: key already exists" {
			return fmt.Errorf("failed to create Count-Min Sketch: %w", err)
		}
		utils.Info("Count-Min Sketch already exists, reusing existing structure")
	} else {
		utils.Info("Count-Min Sketch created successfully")
	}

	utils.Info("Redis data structures initialized successfully")
	return nil
}

// CheckDuplicate checks if a file should be uploaded based on hash and frequency
// Returns: (shouldUpload, reason, error)
func (dc *DuplicateChecker) CheckDuplicate(ctx context.Context, filePath string, cfg *config.Config) (bool, string, error) {
	// Calculate file hash
	hash, err := dc.calculateFileHash(filePath)
	if err != nil {
		return true, "hash calculation failed", fmt.Errorf("failed to calculate file hash: %w", err)
	}

	utils.Info(fmt.Sprintf("Checking duplicate for file: %s, hash: %s", filePath, hash))

	// Step 1: Check Cuckoo Filter
	exists, err := dc.rdb.Do(ctx, "CF.EXISTS", "file_hashes", hash).Bool()
	if err != nil {
		return true, "cuckoo filter check failed", fmt.Errorf("failed to check Cuckoo Filter: %w", err)
	}

	if !exists {
		// File hash not seen before, definitely upload
		utils.Info(fmt.Sprintf("File hash %s not found in Cuckoo Filter, proceeding with upload", hash))
		return true, "new file", nil
	}

	// Step 2: Check upload frequency using Count-Min Sketch
	result, err := dc.rdb.Do(ctx, "CMS.QUERY", "upload_frequency", hash).Result()
	if err != nil {
		return true, "frequency check failed", fmt.Errorf("failed to query Count-Min Sketch: %w", err)
	}

	// Extract frequency from result (CMS.QUERY returns array with single value)
	var frequency int
	switch v := result.(type) {
	case []interface{}:
		if len(v) > 0 {
			if freq, ok := v[0].(int64); ok {
				frequency = int(freq)
			}
		}
	case int64:
		frequency = int(v)
	case int:
		frequency = v
	default:
		utils.Error(fmt.Sprintf("Unexpected CMS.QUERY response type: %T", result))
		return true, "frequency check failed", fmt.Errorf("unexpected CMS.QUERY response type: %T", result)
	}

	utils.Info(fmt.Sprintf("File hash %s found %d times in Count-Min Sketch", hash, frequency))

	// If frequency is high (more than 10 times), skip upload
	if frequency > 10 {
		utils.Info(fmt.Sprintf("File hash %s uploaded %d times, skipping upload", hash, frequency))
		return false, fmt.Sprintf("file uploaded %d times already", frequency), nil
	}

	// Step 3: For low frequency, check S3 directly
	utils.Info(fmt.Sprintf("File hash %s has low frequency (%d), checking S3 directly", hash, frequency))
	existsInS3, err := dc.checkS3Exists(ctx, cfg, filePath)
	if err != nil {
		// If S3 check fails, proceed with upload to be safe
		utils.Error(fmt.Sprintf("S3 check failed for %s: %v", filePath, err))
		return true, "s3 check failed", nil
	}

	if existsInS3 {
		utils.Info(fmt.Sprintf("File %s already exists in S3, skipping upload", filePath))
		return false, "file exists in s3", nil
	}

	utils.Info(fmt.Sprintf("File %s not found in S3, proceeding with upload", filePath))
	return true, "not in s3", nil
}

// RecordUpload records a successful upload in both Cuckoo Filter and Count-Min Sketch
func (dc *DuplicateChecker) RecordUpload(ctx context.Context, filePath string) error {
	hash, err := dc.calculateFileHash(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash for recording: %w", err)
	}

	// Add to Cuckoo Filter
	_, err = dc.rdb.Do(ctx, "CF.ADD", "file_hashes", hash).Result()
	if err != nil {
		return fmt.Errorf("failed to add hash to Cuckoo Filter: %w", err)
	}

	// Increment frequency in Count-Min Sketch
	_, err = dc.rdb.Do(ctx, "CMS.INCRBY", "upload_frequency", hash, 1).Result()
	if err != nil {
		return fmt.Errorf("failed to increment frequency in Count-Min Sketch: %w", err)
	}

	utils.Info(fmt.Sprintf("Recorded upload for file: %s, hash: %s", filePath, hash))
	return nil
}

// calculateFileHash calculates SHA256 hash of a file
func (dc *DuplicateChecker) calculateFileHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// checkS3Exists checks if a file already exists in S3 using HEAD request
func (dc *DuplicateChecker) checkS3Exists(ctx context.Context, cfg *config.Config, filePath string) (bool, error) {
	// Extract key from file path or use provided key
	key := filepath.Base(filePath)
	if cfg.Key != "" {
		key = cfg.Key
	}

	// Perform HEAD request to check if object exists
	_, err := dc.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &cfg.Bucket,
		Key:    &key,
	})

	if err != nil {
		// If object doesn't exist, AWS returns a 404 error
		return false, nil
	}

	return true, nil
}

// GetStats returns statistics about the duplicate checker
func (dc *DuplicateChecker) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get Cuckoo Filter info
	cfInfo, err := dc.rdb.Do(ctx, "CF.INFO", "file_hashes").Result()
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to get Cuckoo Filter info: %v", err))
	} else {
		stats["cuckoo_filter"] = cfInfo
	}

	// Get Count-Min Sketch info
	cmsInfo, err := dc.rdb.Do(ctx, "CMS.INFO", "upload_frequency").Result()
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to get Count-Min Sketch info: %v", err))
	} else {
		stats["count_min_sketch"] = cmsInfo
	}

	return stats, nil
}

// Close closes the Redis connection
func (dc *DuplicateChecker) Close() error {
	return dc.rdb.Close()
}

// createS3Client creates an AWS S3 client using the provided configuration
func createS3Client(cfg *config.Config) (*s3.Client, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")

	var (
		awsCfg aws.Config
		err    error
	)
	if endpoint != "" {
		resolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					HostnameImmutable: true,
				}, nil
			})
		awsCfg, err = awsv2cfg.LoadDefaultConfig(context.Background(),
			awsv2cfg.WithRegion(cfg.Region),
			awsv2cfg.WithEndpointResolverWithOptions(resolver),
		)
	} else {
		awsCfg, err = awsv2cfg.LoadDefaultConfig(context.Background(),
			awsv2cfg.WithRegion(cfg.Region),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true // For LocalStack or custom S3-compatible endpoints
		}
	})

	return cli, nil
}
