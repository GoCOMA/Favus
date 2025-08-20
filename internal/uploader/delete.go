package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	// 1. 명령줄 인수가 올바른지 확인
	if len(os.Args) < 2 {
		fmt.Println("사용법: go run <파일이름>.go <버킷-이름>")
		// os.Exit(1)를 사용하여 프로그램 비정상 종료
		os.Exit(1)
	}
	// 명령줄에서 버킷 이름을 가져옴
	bucketName := os.Args[1]

	// 2. 기본 AWS 설정 로드
	// config.LoadDefaultConfig를 사용하면 AWS 환경 변수, 공유 자격 증명 파일,
	// IAM 역할을 자동으로 찾아서 로드합니다.
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Printf("AWS 설정 로드 오류: %v\n", err)
		os.Exit(1)
	}

	// 3. S3 클라이언트 생성
	// 불필요한 awsutils 패키지 사용을 제거하고, 기본적인 S3 클라이언트를 생성합니다.
	s3Client := s3.NewFromConfig(cfg)

	fmt.Printf("버킷 '%s'에서 불완전한 멀티파트 업로드를 검색하는 중...\n", bucketName)

	var uploads []types.MultipartUpload
	var nextKeyMarker *string
	var nextUploadIdMarker *string

	// 4. ListMultipartUploads API를 사용하여 모든 불완전한 업로드 목록을 가져옴
	for {
		resp, err := s3Client.ListMultipartUploads(context.TODO(), &s3.ListMultipartUploadsInput{
			Bucket:         aws.String(bucketName),
			KeyMarker:      nextKeyMarker,
			UploadIdMarker: nextUploadIdMarker,
		})
		if err != nil {
			fmt.Printf("멀티파트 업로드 목록 가져오기 오류: %v\n", err)
			os.Exit(1)
		}

		uploads = append(uploads, resp.Uploads...)

		// 응답이 잘렸으면 다음 페이지를 계속해서 가져옴
		// *resp.IsTruncated는 포인터 변수의 값을 역참조하여 부울 값으로 사용
		if resp.IsTruncated == nil || !*resp.IsTruncated {
			break
		}
		nextKeyMarker = resp.NextKeyMarker
		nextUploadIdMarker = resp.NextUploadIdMarker
	}

	if len(uploads) == 0 {
		fmt.Println("불완전한 멀티파트 업로드를 찾지 못했습니다.")
		return
	}

	fmt.Printf("불완전한 업로드 %d개를 찾았습니다. 삭제하는 중...\n", len(uploads))

	// 5. AbortMultipartUpload API를 사용하여 각 업로드를 삭제
	for _, upload := range uploads {
		_, err := s3Client.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucketName),
			Key:      upload.Key,
			UploadId: upload.UploadId,
		})
		if err != nil {
			fmt.Printf("키 '%s'와 ID '%s'를 가진 업로드 삭제 오류: %v\n", *upload.Key, *upload.UploadId, err)
		} else {
			fmt.Printf("키 '%s'와 ID '%s'를 가진 업로드를 성공적으로 삭제했습니다.\n", *upload.Key, *upload.UploadId)
		}
	}

	fmt.Println("삭제 작업이 완료되었습니다.")
}