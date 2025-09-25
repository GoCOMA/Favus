package chunker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoCOMA/Favus/internal/config"
	"github.com/GoCOMA/Favus/pkg/utils"
)

var DefaultChunkSize = config.DefaultChunkSize
var chunksDir = config.ChunksDir

// Error type constants
const (
	ErrorNone       = ""
	ErrorFileOpen   = "FILE_OPEN_FAILED"
	ErrorChunkWrite = "CHUNK_WRITE_FAILED"
	ErrorFileRead   = "FILE_READ_FAILED"
)

// ChunkResult represents the result of file chunking operation
type ChunkResult struct {
	Chunks []Chunk
	Error  string // String to distinguish error types
}

// Chunk represents a part of a file to be uploaded.
type Chunk struct {
	Index    int    // Part number
	Offset   int64  // Starting offset in the file
	Size     int64  // Size of the chunk
	FilePath string // Path to the original file
}

// FileChunker provides methods to chunk a file.
type FileChunker struct {
	filePath  string
	fileSize  int64
	chunkSize int64
}

// NewFileChunker creates a new FileChunker.
func NewFileChunker(filePath string, chunkSize int64) (*FileChunker, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// If chunkSize is 0 or negative, use the default chunk size.
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}

	return &FileChunker{
		filePath:  filePath,
		fileSize:  fileInfo.Size(),
		chunkSize: chunkSize,
	}, nil
}

/*
author : popeye
description : returns a slice of Chunks for the file
return : []Chunk
*/
func (fc *FileChunker) Chunks() []Chunk {
	result := fc.splitFileToChunks(fc.filePath)
	if result.Error == ErrorFileOpen {
		utils.LogError("Failed to open file: %s", fc.filePath)
		return nil
	} else if result.Error == ErrorChunkWrite {
		utils.LogError("Failed to write chunk: %s", fc.filePath)
		return nil
	} else if result.Error == ErrorFileRead {
		utils.LogError("Failed to read file: %s", fc.filePath)
		return nil
	} else if len(result.Chunks) == 0 {
		utils.LogWarning("No chunks were generated from file: %s", fc.filePath)
		return nil
	} else {
		utils.LogInfo("Successfully generated %d chunks from file: %s", len(result.Chunks), fc.filePath)
		return result.Chunks
	}
}

/*
author : greensnapback0229, popeye
description : core method for split file to chunks
return : ChunkResult
*/
func (fc *FileChunker) splitFileToChunks(filePath string) ChunkResult {
	utils.LogInfo("Starting file chunking for: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		utils.LogError("Failed to open file: %s, error: %v", filePath, err)
		return ChunkResult{Chunks: nil, Error: ErrorFileOpen}
	}
	defer file.Close()

	utils.LogInfo("Successfully opened file: %s", filePath)

	buf := make([]byte, fc.chunkSize)
	chunkNum := 0
	var chunks []Chunk

	for {
		n, err := file.Read(buf)
		if n > 0 {
			chunkFileName := fmt.Sprintf("%s_chunk%d", filepath.Base(filePath), chunkNum)
			utils.LogInfo("Creating chunk %d: %s (%d bytes)", chunkNum, chunkFileName, n)

			err := writeChunk(chunkFileName, buf[:n])
			if err != nil {
				utils.LogError("Failed to write chunk %d: %s, error: %v", chunkNum, chunkFileName, err)
				utils.LogWarning("Rolling back %d previously created chunks", chunkNum)
				rollbackChunks(filePath, chunkNum)
				return ChunkResult{Chunks: nil, Error: ErrorChunkWrite}
			}

			utils.LogInfo("Successfully created chunk %d: %s", chunkNum, chunkFileName)

			chunks = append(chunks, Chunk{
				Index:    chunkNum + 1,
				Offset:   int64(chunkNum) * fc.chunkSize,
				Size:     int64(n),
				FilePath: filePath,
			})
			chunkNum++
		}
		if err == io.EOF {
			utils.LogInfo("Reached end of file, total chunks created: %d", chunkNum)
			break
		}
		if err != nil {
			utils.LogError("Failed to read file: %s, error: %v", filePath, err)
			utils.LogWarning("Rolling back %d previously created chunks", chunkNum)
			rollbackChunks(filePath, chunkNum)
			return ChunkResult{Chunks: nil, Error: ErrorFileRead}
		}
	}

	utils.LogInfo("File chunking completed successfully: %s, total chunks: %d", filePath, len(chunks))
	return ChunkResult{Chunks: chunks, Error: ErrorNone}
}

/*
author : popeye
description : write chunk to file
return : error
*/
func writeChunk(filename string, data []byte) error {
	// Create chunks directory inside chunker directory if it doesn't exist
	if err := createChunksDirectory(chunksDir); err != nil {
		return err
	}

	// Create full path for chunk file
	chunkFilePath := filepath.Join(chunksDir, filename)

	// Write chunk data to file
	if err := os.WriteFile(chunkFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk file %s: %w", chunkFilePath, err)
	}

	return nil
}

// createChunksDirectory creates the chunks directory if it doesn't exist
func createChunksDirectory(chunksDir string) error {
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunks directory: %w", err)
	}
	return nil
}

/*
author : popeye
description : removes all chunk files that were created during the failed operation
return : void
*/
func rollbackChunks(filePath string, chunkCount int) {
	chunksDir := config.ChunksDir
	baseFileName := filepath.Base(filePath)

	utils.LogWarning("Rolling back %d chunk files for %s", chunkCount, filePath)

	for i := 0; i < chunkCount; i++ {
		chunkFileName := fmt.Sprintf("%s_chunk%d", baseFileName, i)
		chunkFilePath := filepath.Join(chunksDir, chunkFileName)

		if err := os.Remove(chunkFilePath); err != nil {
			utils.LogError("Failed to remove chunk file %s: %v", chunkFilePath, err)
		} else {
			utils.LogInfo("Removed chunk file: %s", chunkFileName)
		}
	}

	// If no chunks were created, remove the chunks directory
	if chunkCount == 0 {
		if err := os.RemoveAll(chunksDir); err != nil {
			utils.LogError("Failed to remove chunks directory %s: %v", chunksDir, err)
		} else {
			utils.LogInfo("Removed chunks directory: %s", chunksDir)
		}
	}
}

/*
author : greensnapback0229
description : split a single file into chunks using splitFileToChunks
return : error
func SplitSingleFile(filePath string) error {
	result := splitFileToChunks(filePath)
	if result.Error != ErrorNone {
		return fmt.Errorf("failed to split file %s: %s", filePath, result.Error)
	}
	return nil
} */

/*
author : greensnapback0229
description : split all regular files in the given directory into chunks using splitFileToChunks
return : error

func SplitAllFilesInDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			filePath := filepath.Join(dirPath, entry.Name())
			result := splitFileToChunks(filePath)
			if result.Error != ErrorNone {
				return fmt.Errorf("failed to split file %s: %s", filePath, result.Error)
			}
		}
	}
	return nil
} */

// ChunkReader implements io.ReadSeekCloser for a specific file chunk.
type ChunkReader struct {
	file   *os.File
	offset int64
	size   int64
	read   int64
}

// Read reads up to len(p) bytes into p.
// It stops reading when the chunk's size limit is reached.
func (cr *ChunkReader) Read(p []byte) (n int, err error) {
	if cr.read >= cr.size {
		return 0, io.EOF // End of this chunk
	}

	remaining := cr.size - cr.read
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err = cr.file.Read(p)
	cr.read += int64(n)
	return n, err
}

// Seek sets the offset for the next Read or Write on the file.
// It ensures seeking stays within the bounds of the chunk.
func (cr *ChunkReader) Seek(offset int64, whence int) (int64, error) {
	newOffsetInChunk := cr.read

	switch whence {
	case io.SeekStart:
		newOffsetInChunk = offset
	case io.SeekCurrent:
		newOffsetInChunk += offset
	case io.SeekEnd:
		newOffsetInChunk = cr.size + offset
	}

	if newOffsetInChunk < 0 || newOffsetInChunk > cr.size {
		return 0, fmt.Errorf("seek offset %d out of bounds for chunk of size %d", offset, cr.size)
	}

	// Calculate the absolute file position
	absoluteFilePos := cr.offset + newOffsetInChunk

	// Seek the underlying file
	_, err := cr.file.Seek(absoluteFilePos, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("failed to seek underlying file to absolute position %d: %w", absoluteFilePos, err)
	}

	cr.read = newOffsetInChunk
	return newOffsetInChunk, nil
}

// Close closes the underlying file.
func (cr *ChunkReader) Close() error {
	return cr.file.Close()
}

// GetChunkReader returns an io.ReadSeekCloser for a specific chunk.
/*
author : popeye
description : returns an io.ReadSeekCloser for a specific chunk
return : io.ReadSeekCloser, error
*/
func (fc *FileChunker) GetChunkReader(chunk Chunk) (io.ReadSeekCloser, error) {
	// Create chunk file path based on the chunk index
	chunkFileName := fmt.Sprintf("%s_chunk%d", filepath.Base(chunk.FilePath), chunk.Index-1)
	chunkFilePath := filepath.Join(chunksDir, chunkFileName)

	// Open the chunk file
	file, err := os.Open(chunkFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open chunk file %s: %w", chunkFilePath, err)
	}

	// Create a new ChunkReader instance
	cr := &ChunkReader{
		file:   file,
		offset: 0, // Chunk files start from offset 0
		size:   chunk.Size,
		read:   0,
	}

	return cr, nil
}

/*
author : popeye
description : removes all chunk files that were created for the current file
return : error
*/
func (fc *FileChunker) CleanupChunks() error {
	utils.LogInfo("Starting cleanup of chunk files for: %s", fc.filePath)

	baseFileName := filepath.Base(fc.filePath)
	chunksDir := chunksDir

	// Get the list of chunk files
	entries, err := os.ReadDir(chunksDir)
	if err != nil {
		utils.LogError("Failed to read chunks directory %s: %v", chunksDir, err)
		return fmt.Errorf("failed to read chunks directory: %w", err)
	}

	// Count and remove chunk files for this specific file
	removedCount := 0
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			fileName := entry.Name()
			// Check if this chunk file belongs to our file
			if strings.HasPrefix(fileName, baseFileName+"_chunk") {
				chunkFilePath := filepath.Join(chunksDir, fileName)
				if err := os.Remove(chunkFilePath); err != nil {
					utils.LogError("Failed to remove chunk file %s: %v", chunkFilePath, err)
					return fmt.Errorf("failed to remove chunk file %s: %w", chunkFilePath, err)
				}
				utils.LogInfo("Removed chunk file: %s", fileName)
				removedCount++
			}
		}
	}

	utils.LogInfo("Successfully removed %d chunk files for: %s", removedCount, fc.filePath)

	// If no chunk files remain in the directory, remove the chunks directory itself
	remainingEntries, err := os.ReadDir(chunksDir)
	if err != nil {
		utils.LogWarning("Failed to check remaining entries in chunks directory: %v", err)
		return nil // Don't fail cleanup if we can't check remaining files
	}

	// Check if only directories remain (no regular files)
	hasRegularFiles := false
	for _, entry := range remainingEntries {
		if entry.Type().IsRegular() {
			hasRegularFiles = true
			break
		}
	}

	if !hasRegularFiles {
		if err := os.RemoveAll(chunksDir); err != nil {
			utils.LogWarning("Failed to remove empty chunks directory %s: %v", chunksDir, err)
		} else {
			utils.LogInfo("Removed empty chunks directory: %s", chunksDir)
		}
	}

	return nil
}
