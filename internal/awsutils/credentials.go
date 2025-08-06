package awsutils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/aws/smithy-go"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// 인증 정보 불러오기: 없으면 프롬프트
func LoadAWSConfig(profile string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// 1차 시도: 기존 config/cached env
	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return cfg, err
	}

	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err == nil {
		fmt.Printf("✅ 인증된 AWS 사용자: %s\n", creds.AccessKeyID)
		return cfg, nil
	}

	// 인증 정보 누락됨
	fmt.Println("❌ AWS 인증 정보가 누락되었습니다.")
	fmt.Println("💬 입력을 통해 인증 정보를 설정합니다.")

	// 🟡 입력 요청
	accessKey, secretKey, region := promptIfEmpty(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_REGION"),
	)

	// 환경변수에 등록
	os.Setenv("AWS_ACCESS_KEY_ID", accessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	os.Setenv("AWS_REGION", region)

	// 2차 시도
	cfg, err = config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return cfg, err
	}

	creds, err = cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return cfg, fmt.Errorf("입력된 인증 정보로도 AWS 인증에 실패했습니다")
	}

	fmt.Printf("✅ 인증된 AWS 사용자: %s\n", creds.AccessKeyID)
	return cfg, nil
}

// 🟡 입력된 값이 없을 때만 프롬프트 출력
func promptIfEmpty(accessKey, secretKey, region string) (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	if strings.TrimSpace(accessKey) == "" {
		fmt.Print("🔑 Enter AWS Access Key ID: ")
		accessKey, _ = reader.ReadString('\n')
	}

	if strings.TrimSpace(secretKey) == "" {
		fmt.Print("🔐 Enter AWS Secret Access Key: ")
		secretKey, _ = reader.ReadString('\n')
	}

	if strings.TrimSpace(region) == "" {
		fmt.Print("🌐 Enter AWS Region (default: ap-northeast-2): ")
		region, _ = reader.ReadString('\n')
	}

	return strings.TrimSpace(accessKey), strings.TrimSpace(secretKey), defaultIfEmpty(region, "ap-northeast-2")
}

func isMissingCredentials(err error) bool {
	var apiErr smithy.APIError
	return errors.As(err, &apiErr)
}

func defaultIfEmpty(val string, def string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return def
	}
	return val
}
