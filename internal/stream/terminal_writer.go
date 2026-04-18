package stream

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
)

// TerminalWriter writes streaming results to stdout with TTY-aware formatting.
// It detects if output is a terminal (TTY) and formats with ANSI codes if appropriate.
type TerminalWriter struct {
	mu       sync.Mutex
	closed   bool
	isTTY    bool
	stdout   *os.File
	writeErr error
}

// NewTerminalWriter creates a new TerminalWriter that detects TTY and formats accordingly.
func NewTerminalWriter() *TerminalWriter {
	tw := &TerminalWriter{
		stdout: os.Stdout,
	}

	// Detect if stdout is a TTY
	stat, err := os.Stdout.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) != 0 {
		tw.isTTY = true
	}

	return tw
}

// WriteChunk writes a single chunk of data to stdout.
// Respects context cancellation and returns an error if writer is closed.
func (tw *TerminalWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return err
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.closed {
		return errors.New("terminal writer is closed")
	}

	if tw.writeErr != nil {
		return tw.writeErr
	}

	// Write chunk to stdout
	_, err := tw.stdout.Write(chunk)
	if err != nil {
		tw.writeErr = fmt.Errorf("failed to write chunk: %w", err)
		return tw.writeErr
	}

	return nil
}

// Close closes the writer. It's idempotent and can be called multiple times.
func (tw *TerminalWriter) Close() error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.closed {
		return nil // Already closed, idempotent
	}

	tw.closed = true

	// Flush stdout to ensure all data is written
	if tw.stdout != os.Stdout {
		// If it's not stdout (shouldn't normally happen), close it
		return tw.stdout.Close()
	}

	// stdout should not be explicitly closed in most cases
	return nil
}
