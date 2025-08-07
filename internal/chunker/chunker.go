package chunker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const chunkSize = 4 * 1024 * 1024 // 4MB

/*
*

	split chunk from single file
*/
func splitFileToChunks(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, chunkSize)
	chunkNum := 0

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunkFileName := fmt.Sprintf("%s_chunk%d", filepath.Base(filePath), chunkNum)
			err := writeChunk(chunkFileName, buf[:n])
			if err != nil {
				return fmt.Errorf("failed to write chunk %d: %w", chunkNum, err)
			}
			fmt.Printf("Wrote chunk %d: %s (%d bytes)\n", chunkNum, chunkFileName, n)
			chunkNum++
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}
	return nil
}

func writeChunk(filename string, data []byte) error {
	//Todo: write chunk to s3
}

func main() {
	filePath := "video.mp4" // 분할할 파일명으로 변경하세요
	err := splitFileToChunks(filePath)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
