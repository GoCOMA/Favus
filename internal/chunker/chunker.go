package favus

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SplitFileWithRetry는 재시도 로직이 포함된 파일 분할 함수
// maxRetries: 최대 재시도 횟수
// resumeOnError: true면 중단된 지점부터 재시도, false면 처음부터 재시도
func SplitFileWithRetry(srcPath, dstDir string, partSize int64, maxRetries int, resumeOnError bool, progress func(done, total int)) ([]string, error) {
	base := filepath.Base(srcPath)
	LogInfo("원본파일명 '%s' 분할 시작 (재시도 로직 포함)", base)

	var lastCompletedPart int
	var parts []string
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			LogInfo("재시도 %d/%d 시작... (마지막 완료 파트: %d)", attempt, maxRetries, lastCompletedPart)
		}

		if resumeOnError && attempt > 0 && lastCompletedPart > 0 {
			// 중단된 지점부터 재시도
			parts, err = splitFileFromPart(srcPath, dstDir, partSize, lastCompletedPart+1, &lastCompletedPart, progress)
		} else {
			// 처음부터 재시도
			parts, err = splitFileWithProgressTracking(srcPath, dstDir, partSize, &lastCompletedPart, progress)
		}

		if err == nil {
			LogInfo("원본파일명 '%s' 분할 성공: 총 %d개 파트 생성", base, len(parts))
			return parts, nil
		}

		LogError("시도 %d 실패: %v", attempt+1, err)

		if attempt < maxRetries {
			// 잠시 대기 후 재시도
			waitTime := time.Duration(attempt+1) * time.Second
			LogInfo("%v 후 재시도...", waitTime)
			time.Sleep(waitTime)
		}
	}

	// 최대 재시도 횟수 초과 시 분할된 파일들 삭제
	LogWarning("원본파일명 '%s' 분할 실패: 최대 재시도 횟수(%d) 초과", base, maxRetries)
	LogInfo("분할된 파일들을 정리합니다...")
	if len(parts) > 0 {
		if removeErr := RemoveParts(parts); removeErr != nil {
			LogError("파일 정리 중 오류: %v", removeErr)
		} else {
			LogInfo("분할된 파일들 삭제 완료")
		}
	}

	return nil, fmt.Errorf("최대 재시도 횟수(%d) 초과: %w", maxRetries, err)
}

// splitFileWithProgressTracking은 진행률을 추적하면서 파일을 분할
func splitFileWithProgressTracking(srcPath, dstDir string, partSize int64, lastCompletedPart *int, progress func(done, total int)) ([]string, error) {
	base := filepath.Base(srcPath)
	LogInfo("원본파일명 '%s' 분할 시작 (진행률 추적)", base)

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		LogError("분할 디렉토리 생성 실패: %v", err)
		return nil, fmt.Errorf("분할 디렉토리 생성 실패: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			LogError("원본 파일이 존재하지 않음: %s", srcPath)
		} else if os.IsPermission(err) {
			LogError("원본 파일 접근 권한 없음: %s", srcPath)
		} else {
			LogError("원본 파일 열기 실패: %v", err)
		}
		return nil, fmt.Errorf("원본 파일 열기 실패: %w", err)
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		LogError("파일 정보 조회 실패: %v", err)
		return nil, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	fileSize := fileInfo.Size()
	totalParts := int((fileSize + partSize - 1) / partSize)
	base = filepath.Base(srcPath)

	LogInfo("진행률 추적 - 파일 크기: %d bytes, 총 파트 수: %d", fileSize, totalParts)

	const maxWorkers = 4
	workerPool := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// Context를 사용하여 에러 발생 시 즉시 중단
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, totalParts) // 모든 고루틴의 에러를 받을 수 있도록
	parts := make([]string, totalParts)

	var completedParts int
	var mu sync.Mutex

	for partNum := 1; partNum <= totalParts; partNum++ {
		// Context가 취소되었는지 확인
		select {
		case <-ctx.Done():
			// 이미 에러가 발생했으므로 대기 후 종료
			wg.Wait()
			RemoveParts(parts) // 에러 시에는 삭제
			return nil, <-errChan
		default:
		}

		workerPool <- struct{}{}
		startPos := int64(partNum-1) * partSize
		currentPartSize := partSize
		if startPos+partSize > fileSize {
			currentPartSize = fileSize - startPos
		}
		partName := fmt.Sprintf("%s.part%03d", base, partNum)
		partPath := filepath.Join(dstDir, partName)
		parts[partNum-1] = partPath

		wg.Add(1)
		go func(partNum int, startPos, currentPartSize int64, partPath string) {
			defer wg.Done()
			defer func() { <-workerPool }()

			// Context 취소 확인
			select {
			case <-ctx.Done():
				return
			default:
			}

			src, err := os.Open(srcPath)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 원본 파일 열기 실패: %w", partNum, err):
					cancel() // 다른 고루틴들 중단
				case <-ctx.Done():
				}
				return
			}
			defer src.Close()

			if _, err := src.Seek(startPos, io.SeekStart); err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 파일 포인터 이동 실패: %w", partNum, err):
					cancel()
				case <-ctx.Done():
				}
				return
			}

			out, err := os.Create(partPath)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 출력 파일 생성 실패: %w", partNum, err):
					cancel()
				case <-ctx.Done():
				}
				return
			}
			defer out.Close()

			bufferedWriter := bufio.NewWriterSize(out, 64*1024)
			defer bufferedWriter.Flush()
			buf := make([]byte, 64*1024)
			written := int64(0)

			for written < currentPartSize {
				// Context 취소 확인
				select {
				case <-ctx.Done():
					return
				default:
				}

				readSize := int64(len(buf))
				if remain := currentPartSize - written; remain < readSize {
					readSize = remain
				}
				n, err := io.ReadFull(src, buf[:readSize])
				if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
					select {
					case errChan <- fmt.Errorf("파트 %d: 파일 읽기 실패: %w", partNum, err):
						cancel()
					case <-ctx.Done():
					}
					return
				}
				if n == 0 {
					break
				}
				if _, err := bufferedWriter.Write(buf[:n]); err != nil {
					select {
					case errChan <- fmt.Errorf("파트 %d: 파일 쓰기 실패: %w", partNum, err):
						cancel()
					case <-ctx.Done():
					}
					return
				}
				written += int64(n)
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					break
				}
			}

			mu.Lock()
			completedParts++
			*lastCompletedPart = completedParts // 마지막 완료된 파트 업데이트
			if progress != nil {
				progress(completedParts, totalParts)
			}
			mu.Unlock()
		}(partNum, startPos, currentPartSize, partPath)
	}

	wg.Wait()
	close(errChan) // 에러 채널 닫기

	// 에러가 있었는지 확인
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
			LogError("원본파일명 '%s' 분할 실패 (진행률 추적): %v", base, err)
			LogInfo("진행률 추적 - 분할된 파일들 정리 중...")
			RemoveParts(parts) // 에러 시에는 삭제
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	LogInfo("원본파일명 '%s' 분할 성공 (진행률 추적): 총 %d개 파트 생성됨", base, len(parts))
	return parts, nil
}

// splitFileFromPart는 특정 파트부터 파일 분할을 시작
func splitFileFromPart(srcPath, dstDir string, partSize int64, startPart int, lastCompletedPart *int, progress func(done, total int)) ([]string, error) {
	base := filepath.Base(srcPath)
	LogInfo("원본파일명 '%s' 분할 시작 (특정 파트부터: %d번째 파트)", base, startPart)

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		LogError("분할 디렉토리 생성 실패: %v", err)
		return nil, fmt.Errorf("분할 디렉토리 생성 실패: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			LogError("원본 파일이 존재하지 않음: %s", srcPath)
		} else if os.IsPermission(err) {
			LogError("원본 파일 접근 권한 없음: %s", srcPath)
		} else {
			LogError("원본 파일 열기 실패: %v", err)
		}
		return nil, fmt.Errorf("원본 파일 열기 실패: %w", err)
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		LogError("파일 정보 조회 실패: %v", err)
		return nil, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	fileSize := fileInfo.Size()
	totalParts := int((fileSize + partSize - 1) / partSize)
	base = filepath.Base(srcPath)

	LogInfo("특정 파트부터 분할 - 파일 크기: %d bytes, 총 파트 수: %d, 시작 파트: %d", fileSize, totalParts, startPart)

	// 이미 완료된 파트들 확인
	existingParts := make([]string, totalParts)
	for i := 1; i <= totalParts; i++ {
		partName := fmt.Sprintf("%s.part%03d", base, i)
		partPath := filepath.Join(dstDir, partName)
		existingParts[i-1] = partPath
	}

	const maxWorkers = 4
	workerPool := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, totalParts) // 모든 고루틴의 에러를 받을 수 있도록
	parts := make([]string, totalParts)
	copy(parts, existingParts)

	var completedParts int
	var mu sync.Mutex

	// 진행률 초기화 (이미 완료된 파트들)
	completedParts = startPart - 1
	*lastCompletedPart = completedParts // 초기값 설정
	if progress != nil {
		progress(completedParts, totalParts)
	}

	for partNum := startPart; partNum <= totalParts; partNum++ {
		select {
		case <-ctx.Done():
			wg.Wait()
			RemoveParts(parts)
			return nil, <-errChan
		default:
		}

		workerPool <- struct{}{}
		startPos := int64(partNum-1) * partSize
		currentPartSize := partSize
		if startPos+partSize > fileSize {
			currentPartSize = fileSize - startPos
		}
		partName := fmt.Sprintf("%s.part%03d", base, partNum)
		partPath := filepath.Join(dstDir, partName)
		parts[partNum-1] = partPath

		wg.Add(1)
		go func(partNum int, startPos, currentPartSize int64, partPath string) {
			defer wg.Done()
			defer func() { <-workerPool }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			src, err := os.Open(srcPath)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 원본 파일 열기 실패: %w", partNum, err):
					cancel()
				case <-ctx.Done():
				}
				return
			}
			defer src.Close()

			if _, err := src.Seek(startPos, io.SeekStart); err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 파일 포인터 이동 실패: %w", partNum, err):
					cancel()
				case <-ctx.Done():
				}
				return
			}

			out, err := os.Create(partPath)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("파트 %d: 출력 파일 생성 실패: %w", partNum, err):
					cancel()
				case <-ctx.Done():
				}
				return
			}
			defer out.Close()

			bufferedWriter := bufio.NewWriterSize(out, 64*1024)
			defer bufferedWriter.Flush()
			buf := make([]byte, 64*1024)
			written := int64(0)

			for written < currentPartSize {
				select {
				case <-ctx.Done():
					return
				default:
				}

				readSize := int64(len(buf))
				if remain := currentPartSize - written; remain < readSize {
					readSize = remain
				}
				n, err := io.ReadFull(src, buf[:readSize])
				if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
					select {
					case errChan <- fmt.Errorf("파트 %d: 파일 읽기 실패: %w", partNum, err):
						cancel()
					case <-ctx.Done():
					}
					return
				}
				if n == 0 {
					break
				}
				if _, err := bufferedWriter.Write(buf[:n]); err != nil {
					select {
					case errChan <- fmt.Errorf("파트 %d: 파일 쓰기 실패: %w", partNum, err):
						cancel()
					case <-ctx.Done():
					}
					return
				}
				written += int64(n)
				if err == io.ErrUnexpectedEOF || err == io.EOF {
					break
				}
			}

			mu.Lock()
			completedParts++
			*lastCompletedPart = completedParts // 마지막 완료된 파트 업데이트
			if progress != nil {
				progress(completedParts, totalParts)
			}
			mu.Unlock()
		}(partNum, startPos, currentPartSize, partPath)
	}

	wg.Wait()
	close(errChan) // 에러 채널 닫기

	// 에러가 있었는지 확인
	var firstErr error
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
			LogError("원본파일명 '%s' 분할 실패 (특정 파트부터): %v", base, err)
			LogInfo("특정 파트부터 분할 - 분할된 파일들 정리 중...")
			RemoveParts(parts) // 에러 시에는 삭제
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	LogInfo("원본파일명 '%s' 분할 성공 (특정 파트부터): 총 %d개 파트 생성됨", base, len(parts))
	return parts, nil
}

func RemoveParts(parts []string) error {
	LogInfo("분할된 파일 %d개 삭제 시작", len(parts))
	var errs []string
	for _, p := range parts {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			LogError("파일 삭제 실패: %s - %v", p, err)
			errs = append(errs, fmt.Sprintf("%s: %v", p, err))
		}
	}
	if len(errs) > 0 {
		LogError("분할 파일 삭제 중 일부 실패: %s", strings.Join(errs, ", "))
		return fmt.Errorf("분할 파일 삭제 중 일부 실패: %s", strings.Join(errs, ", "))
	}
	LogInfo("분할된 파일 %d개 삭제 완료", len(parts))
	return nil
}
