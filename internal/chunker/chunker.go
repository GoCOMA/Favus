package chunker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const chunkSize = 4 * 1024 * 1024 // 4MB

/*
author : greensnapback0229
description : core method for split file to chunks
return : error
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

/*
author : greensnapback0229
description : split a single file into chunks using splitFileToChunks
return : error
*/
func SplitSingleFile(filePath string) error {
	return splitFileToChunks(filePath)
}

/*
author : greensnapback0229
description : split all regular files in the given directory into chunks using splitFileToChunks
return : error
*/
func SplitAllFilesInDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			filePath := filepath.Join(dirPath, entry.Name())
			err := splitFileToChunks(filePath)
			if err != nil {
				return fmt.Errorf("failed to split file %s: %w", filePath, err)
			}
		}
	}
	return nil
}
