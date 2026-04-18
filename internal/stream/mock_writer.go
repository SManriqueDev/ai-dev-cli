package stream

import (
	"context"
	"errors"
	"sync"
)

// MockWriter records streaming results in memory for testing purposes.
type MockWriter struct {
	mu       sync.Mutex
	closed   bool
	chunks   [][]byte
	writeErr error
}

// NewMockWriter creates a new MockWriter that records chunks in memory.
func NewMockWriter() *MockWriter {
	return &MockWriter{
		chunks: make([][]byte, 0),
	}
}

// WriteChunk writes a single chunk of data to the mock writer's buffer.
// Respects context cancellation and returns an error if writer is closed.
func (mw *MockWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	// Check context cancellation first
	if err := ctx.Err(); err != nil {
		return err
	}

	mw.mu.Lock()
	defer mw.mu.Unlock()

	if mw.closed {
		return errors.New("mock writer is closed")
	}

	if mw.writeErr != nil {
		return mw.writeErr
	}

	// Make a copy of the chunk so it's not mutated by caller
	chunkCopy := make([]byte, len(chunk))
	copy(chunkCopy, chunk)

	mw.chunks = append(mw.chunks, chunkCopy)

	return nil
}

// Close closes the writer. It's idempotent and can be called multiple times.
func (mw *MockWriter) Close() error {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	if mw.closed {
		return nil // Already closed, idempotent
	}

	mw.closed = true
	return nil
}

// GetChunks returns all chunks that were written to this mock writer.
// This is for assertion purposes in tests.
func (mw *MockWriter) GetChunks() [][]byte {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	// Return a copy so caller can't mutate our internal state
	result := make([][]byte, len(mw.chunks))
	copy(result, mw.chunks)
	return result
}

// GetChunkCount returns the number of chunks written.
// This is for assertion purposes in tests.
func (mw *MockWriter) GetChunkCount() int {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	return len(mw.chunks)
}

// CombinedContent returns all chunks concatenated into a single byte slice.
// This is for assertion purposes in tests.
func (mw *MockWriter) CombinedContent() []byte {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	totalLen := 0
	for _, chunk := range mw.chunks {
		totalLen += len(chunk)
	}

	result := make([]byte, 0, totalLen)
	for _, chunk := range mw.chunks {
		result = append(result, chunk...)
	}

	return result
}
