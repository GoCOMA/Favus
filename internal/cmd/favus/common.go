package favus

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/GoCOMA/Favus/internal/awsutils"
	"github.com/GoCOMA/Favus/internal/config"
	"github.com/GoCOMA/Favus/internal/uploader"
	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	MinPartSizeMB  = 5
	MinConcurrency = 1
	DefaultRegion  = "ap-northeast-2"
)

type ConfigValidator struct {
	RequiredFields []string
	Config         *config.Config
}

func NewConfigValidator(cfg *config.Config) *ConfigValidator {
	return &ConfigValidator{
		Config: cfg,
	}
}

func (cv *ConfigValidator) RequireBucket() *ConfigValidator {
	cv.RequiredFields = append(cv.RequiredFields, "bucket")
	return cv
}

func (cv *ConfigValidator) RequireKey() *ConfigValidator {
	cv.RequiredFields = append(cv.RequiredFields, "key")
	return cv
}

func (cv *ConfigValidator) RequireRegion() *ConfigValidator {
	cv.RequiredFields = append(cv.RequiredFields, "region")
	return cv
}

func (cv *ConfigValidator) IsValid() bool {
	for _, field := range cv.RequiredFields {
		switch field {
		case "bucket":
			if strings.TrimSpace(cv.Config.Bucket) == "" {
				return false
			}
		case "key":
			if strings.TrimSpace(cv.Config.Key) == "" {
				return false
			}
		case "region":
			if strings.TrimSpace(cv.Config.Region) == "" {
				return false
			}
		}
	}
	return true
}

func PromptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func PromptRequired(label string) string {
	for {
		value := PromptInput(label)
		if value != "" {
			return value
		}
		fmt.Println("값이 비어있습니다. 다시 입력해주세요.")
	}
}

func PromptWithDefault(label, defaultValue string) string {
	value := PromptInput(fmt.Sprintf("%s (default: %s)", label, defaultValue))
	if value == "" {
		return defaultValue
	}
	return value
}

func PromptIntWithValidation(label string, defaultValue, minValue int) int {
	// Ensure default is at least the minimum
	if defaultValue < minValue {
		defaultValue = minValue
	}

	prompt := fmt.Sprintf("%s (minimum %d) [%d]", label, minValue, defaultValue)
	input := PromptInput(prompt)

	if input == "" {
		return defaultValue
	}

	if value, err := strconv.Atoi(input); err == nil && value >= minValue {
		return value
	}
	return minValue
}

func PromptYesNoDefault(label string, defaultYes bool) bool {
	defaultHint := "y/N"
	if defaultYes {
		defaultHint = "Y/n"
	}
	for {
		input := PromptInput(fmt.Sprintf("%s [%s]", label, defaultHint))
		if input == "" {
			return defaultYes
		}
		switch strings.ToLower(input) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("y 또는 n으로 응답해주세요.")
		}
	}
}

func LoadConfigWithOverrides(flagBucket, flagKey, flagRegion string) (*config.Config, error) {
	conf := GetLoadedConfig()
	if conf == nil {
		return nil, fmt.Errorf("config not loaded (PersistentPreRunE should have populated it)")
	}

	if flagBucket != "" {
		conf.Bucket = strings.TrimSpace(flagBucket)
	}
	if flagKey != "" {
		conf.Key = strings.TrimSpace(flagKey)
	}
	if flagRegion != "" {
		conf.Region = strings.TrimSpace(flagRegion)
	}

	return conf, nil
}

func CreateUploaderWithAWS(conf *config.Config) (*uploader.Uploader, error) {
	awsCfg, err := awsutils.LoadAWSConfig(profile)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	up, err := uploader.NewUploaderWithAWSConfig(conf, awsCfg)
	if err != nil {
		return nil, fmt.Errorf("init uploader: %w", err)
	}

	return up, nil
}

func ValidateFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}
	return nil
}

func FormatSuccessMessage(action, bucket, key string) string {
	return fmt.Sprintf("✅ %s → s3://%s/%s", action, bucket, key)
}

func PromptForMissingConfig(validator *ConfigValidator) {
	for _, field := range validator.RequiredFields {
		switch field {
		case "bucket":
			if strings.TrimSpace(validator.Config.Bucket) == "" {
				validator.Config.Bucket = PromptRequired("🔧 Enter S3 bucket name")
			}
		case "key":
			if strings.TrimSpace(validator.Config.Key) == "" {
				validator.Config.Key = PromptRequired("📝 Enter S3 object key")
			}
		case "region":
			if strings.TrimSpace(validator.Config.Region) == "" {
				validator.Config.Region = PromptRequired("🌐 Enter AWS Region")
			}
		}
	}
}

func ToStringPtr(s string) *string {
	return aws.String(s)
}

func StringPtrValue(s *string) string {
	return aws.ToString(s)
}
