package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultRegion     = "ap-northeast-2"
	minPartSizeMB     = 5
	defaultPartSizeMB = 5
)

// Backward-compatibility for develop branch users of config.DefaultChunkSize (bytes)
var DefaultChunkSize int64 = int64(defaultPartSizeMB) * 1024 * 1024

type Config struct {
	Bucket         string `mapstructure:"bucket"`
	Key            string `mapstructure:"key"`
	Region         string `mapstructure:"region"`
	PartSizeMB     int    `mapstructure:"partSizeMB"`
	MaxConcurrency int    `mapstructure:"maxConcurrency"`
	UploadID       string
}

// --- File Loader + ENV Overlay (develop compatibility) ---
func LoadConfig(path string) (*Config, error) {
	// Default values
	conf := &Config{
		Region:         defaultRegion,
		PartSizeMB:     defaultPartSizeMB,
		MaxConcurrency: 4,
	}

	// Load from config file (main branch logic)
	if path != "" {
		v := viper.New()
		v.SetConfigFile(path)
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err := v.Unmarshal(&conf); err != nil {
			return nil, fmt.Errorf("unmarshal config: %w", err)
		}
	}

	// Apply environment variable overrides (develop branch compatibility)
	applyEnvOverrides(conf)

	// Ensure minimum values
	if conf.Region == "" {
		conf.Region = defaultRegion
	}
	if conf.PartSizeMB < minPartSizeMB {
		conf.PartSizeMB = minPartSizeMB
	}

	// Keep develop's global in sync (bytes)
	DefaultChunkSize = int64(conf.PartSizeMB) * 1024 * 1024

	fmt.Printf("[DEBUG] effective config: %+v\n", *conf)
	return conf, nil
}

// Map develop's ENV variables into main's Config struct
// - S3_BUCKET_NAME -> Bucket
// - AWS_REGION     -> Region
// - CHUNK_SIZE (in bytes) -> PartSizeMB (ceil, min 5MB)
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("S3_BUCKET_NAME"); v != "" {
		c.Bucket = strings.TrimSpace(v)
	}
	if v := os.Getenv("AWS_REGION"); v != "" {
		c.Region = strings.TrimSpace(v)
	}
	if v := os.Getenv("CHUNK_SIZE"); v != "" {
		if b, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil && b > 0 {
			mb := int((b + (1024*1024 - 1)) / (1024 * 1024)) // ceil to MB
			if mb < minPartSizeMB {
				mb = minPartSizeMB
			}
			c.PartSizeMB = mb
		} else {
			fmt.Printf("Warning: invalid CHUNK_SIZE '%s'. Using %dMB.\n", v, minPartSizeMB)
		}
	}
}

// Convenience: get part size in bytes
func (c *Config) PartSizeBytes() int64 {
	mb := c.PartSizeMB
	if mb < minPartSizeMB {
		mb = minPartSizeMB
	}
	return int64(mb) * 1024 * 1024
}

// --- Upload Prompt (main branch logic) ---
func PromptForUploadConfig(existingBucket, existingKey string) *Config {
	reader := bufio.NewReader(os.Stdin)

	bucket := strings.TrimSpace(existingBucket)
	if bucket == "" {
		fmt.Print("🔧 Enter S3 bucket name: ")
		b, _ := reader.ReadString('\n')
		bucket = strings.TrimSpace(b)
	}

	key := strings.TrimSpace(existingKey)
	if key == "" {
		fmt.Print("📝 Enter S3 object key: ")
		k, _ := reader.ReadString('\n')
		key = strings.TrimSpace(k)
	}

	fmt.Print("🌐 Enter AWS Region (default: ap-northeast-2): ")
	region, _ := reader.ReadString('\n')

	fmt.Print("📦 Enter part size in MB (minimum 5): ")
	partStr, _ := reader.ReadString('\n')

	fmt.Print("🔁 Enter max concurrency (minimum 1): ")
	conStr, _ := reader.ReadString('\n')

	return &Config{
		Bucket:         bucket,
		Key:            key,
		Region:         defaultIfEmpty(region, defaultRegion),
		PartSizeMB:     parseIntWithMin(partStr, defaultPartSizeMB, minPartSizeMB),
		MaxConcurrency: parseIntWithMin(conStr, 4, 1),
	}
}

// --- Resume Prompt (main branch logic) ---
func PromptForResumeConfig() *Config {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("🔧 Enter S3 bucket name: ")
	bucket, _ := reader.ReadString('\n')

	fmt.Print("📝 Enter S3 object key: ")
	key, _ := reader.ReadString('\n')

	fmt.Print("🌐 Enter AWS region (default: ap-northeast-2): ")
	region, _ := reader.ReadString('\n')

	fmt.Print("🔁 Enter Upload ID: ")
	uploadID, _ := reader.ReadString('\n')

	return &Config{
		Bucket:   strings.TrimSpace(bucket),
		Key:      strings.TrimSpace(key),
		Region:   defaultIfEmpty(region, defaultRegion),
		UploadID: strings.TrimSpace(uploadID),
	}
}

// --- Simple Bucket Prompt (main branch logic) ---
func PromptForSimpleBucket(bucketInput, regionInput string) *Config {
	reader := bufio.NewReader(os.Stdin)

	bucket := strings.TrimSpace(bucketInput)
	if bucket == "" {
		fmt.Print("🔧 Enter S3 bucket name: ")
		input, _ := reader.ReadString('\n')
		bucket = strings.TrimSpace(input)
	}

	region := strings.TrimSpace(regionInput)
	if region == "" {
		fmt.Print("🌐 Enter AWS region (default: ap-northeast-2): ")
		input, _ := reader.ReadString('\n')
		region = strings.TrimSpace(input)
	}

	return &Config{
		Bucket: bucket,
		Region: defaultIfEmpty(region, defaultRegion),
	}
}

// --- Helpers ---
func parseIntWithMin(input string, defaultVal, minVal int) int {
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(input)
	if err != nil || val < minVal {
		return minVal
	}
	return val
}

func defaultIfEmpty(val string, def string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	return val
}
