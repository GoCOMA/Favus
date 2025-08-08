package config

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Bucket         string `mapstructure:"bucket"`
	Key            string `mapstructure:"key"`
	Region         string `mapstructure:"region"`
	PartSizeMB     int    `mapstructure:"partSizeMB"`
	MaxConcurrency int    `mapstructure:"maxConcurrency"`
	UploadID       string
}

// --- File Loader ---
func LoadConfig(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var conf Config
	if err := v.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	fmt.Printf("[DEBUG] config loaded from file: %+v\n", conf)
	return &conf, nil
}

// --- Upload Prompt ---
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
		Region:         defaultIfEmpty(region, "ap-northeast-2"),
		PartSizeMB:     parseIntWithMin(partStr, 10, 5),
		MaxConcurrency: parseIntWithMin(conStr, 4, 1),
	}
}

// --- Resume Prompt ---
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
		Region:   defaultIfEmpty(region, "ap-northeast-2"),
		UploadID: strings.TrimSpace(uploadID),
	}
}

// --- ls-orphans Prompt ---
func PromptForSimpleBucket(bucketInput, regionInput string) *Config {
	reader := bufio.NewReader(os.Stdin)

	// bucket
	bucket := strings.TrimSpace(bucketInput)
	if bucket == "" {
		fmt.Print("🔧 Enter S3 bucket name: ")
		input, _ := reader.ReadString('\n')
		bucket = strings.TrimSpace(input)
	}

	// region
	region := strings.TrimSpace(regionInput)
	if region == "" {
		fmt.Print("🌐 Enter AWS region (default: ap-northeast-2): ")
		input, _ := reader.ReadString('\n')
		region = strings.TrimSpace(input)
	}

	return &Config{
		Bucket: bucket,
		Region: defaultIfEmpty(region, "ap-northeast-2"),
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
