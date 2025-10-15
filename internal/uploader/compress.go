package uploader

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/GoCOMA/Favus/pkg/utils"
)

// compressToTempGzip creates a gzipped copy of src under ~/.favus/compressed and returns the new path.
func compressToTempGzip(srcPath string) (string, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("stat source file: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}

	destDir := filepath.Join(home, ".favus", "compressed")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create compressed directory: %w", err)
	}

	base := filepath.Base(srcPath)
	timestamp := time.Now().UnixNano()
	destPath := filepath.Join(destDir, fmt.Sprintf("%s.%d.gz", base, timestamp))

	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create compressed file: %w", err)
	}

	gw := gzip.NewWriter(dest)
	gw.Name = base
	gw.ModTime = info.ModTime()

	utils.Info(fmt.Sprintf("Compressing %s → %s (gzip)", srcPath, destPath))
	if _, err := io.Copy(gw, src); err != nil {
		_ = gw.Close()
		_ = dest.Close()
		_ = os.Remove(destPath)
		return "", fmt.Errorf("gzip copy failed: %w", err)
	}

	if err := gw.Close(); err != nil {
		_ = dest.Close()
		_ = os.Remove(destPath)
		return "", fmt.Errorf("finalize gzip writer: %w", err)
	}

	if err := dest.Close(); err != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("close compressed file: %w", err)
	}

	utils.Info(fmt.Sprintf("Compression finished: %s (size %d bytes)", destPath, fileSize(destPath)))
	return destPath, nil
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
