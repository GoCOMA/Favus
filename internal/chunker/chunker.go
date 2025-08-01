package chunker

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"favus/internal/config"
)

const DefaultChunkSize = config.DefaultChunkSize

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

// Chunks returns a slice of Chunks for the file.
func (fc *FileChunker) Chunks() []Chunk {
	var chunks []Chunk
	for i := 0; ; i++ {
		offset := int64(i) * fc.chunkSize
		remaining := fc.fileSize - offset
		if remaining <= 0 {
			break
		}

		chunkSize := fc.chunkSize
		if remaining < chunkSize {
			chunkSize = remaining
		}

		chunks = append(chunks, Chunk{
			Index:    i + 1, // S3 part numbers start from 1
			Offset:   offset,
			Size:     chunkSize,
			FilePath: fc.filePath,
		})
	}
	return chunks
}

// ChunkReader implements io.ReadSeekCloser for a specific file chunk.
type ChunkReader struct {
	file   *os.File
	offset int64
	size   int64
	read   int64

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
// It opens the file for each chunk to ensure proper seeking and closing.
func (fc *FileChunker) GetChunkReader(chunk Chunk) (io.ReadSeekCloser, error) {
	file, err := os.Open(fc.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for chunk reader: %w", err)
	}

	// Create a new ChunkReader instance
	cr := &ChunkReader{
		file:   file,
		offset: chunk.Offset,
		size:   chunk.Size,
		read:   0,
	}

	// Seek to the start of the chunk immediately after opening the file
	// This ensures that subsequent Reads from the ChunkReader start from the correct position.
	_, err = cr.file.Seek(chunk.Offset, io.SeekStart)
	if err != nil {
		cr.Close() // Close the file if seek fails
		return nil, fmt.Errorf("failed to seek to chunk offset %d: %w", chunk.Offset, err)
	}

	return cr, nil
}
