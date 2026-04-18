package stream

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileWriter writes streaming results to a file.
type FileWriter struct {
	mu       sync.Mutex
	closed   bool
	file     *os.File
	writeErr error
}

// NewFileWriter creates a new FileWriter that writes to the specified file path.
// Creates or truncates the file if it exists.
// Validates file path to prevent directory traversal attacks.
func NewFileWriter(filePath string) (*FileWriter, error) {
	// Validate and clean path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)

	// For relative paths, ensure they don't escape the current directory
	if !filepath.IsAbs(cleanPath) && !filepath.IsLocal(cleanPath) {
		return nil, fmt.Errorf("invalid file path: attempted directory traversal")
	}

	file, err := os.Create(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return &FileWriter{
		file: file,
	}, nil
}

// WriteChunk writes a single chunk of data to the file.
// Respects context cancellation and returns an error if writer is closed.
func (fw *FileWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return err
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.closed {
		return errors.New("file writer is closed")
	}

	if fw.writeErr != nil {
		return fw.writeErr
	}

	// Write chunk to file
	_, err := fw.file.Write(chunk)
	if err != nil {
		fw.writeErr = fmt.Errorf("failed to write chunk to file: %w", err)
		return fw.writeErr
	}

	return nil
}

// Close closes the writer and the underlying file.
// It's idempotent and can be called multiple times.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.closed {
		return nil // Already closed, idempotent
	}

	fw.closed = true

	if fw.file != nil {
		return fw.file.Close()
	}

	return nil
}
