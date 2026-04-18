package stream

import (
	"context"
)

// StreamWriter defines the interface for writing streaming results.
// It abstracts different output targets (terminal, file, pipe, mock).
// Implementations must be thread-safe if concurrent writes are expected.
type StreamWriter interface {
	// WriteChunk writes a single chunk of data to the output stream.
	// It respects context cancellation and should return context.Canceled if the context is cancelled.
	// Implementations should not buffer the entire response - stream chunks as they arrive.
	//
	// Parameters:
	//   - ctx: context.Context for cancellation and deadlines
	//   - chunk: []byte containing the data to write
	//
	// Returns:
	//   - error: nil if successful, wrapped error if write fails
	//           fmt.Errorf("%w", underlying_error) for proper error chaining
	WriteChunk(ctx context.Context, chunk []byte) error

	// Close closes the writer and releases any associated resources.
	// It should be idempotent - calling Close multiple times should not cause errors.
	// After Close, subsequent WriteChunk calls should return an error indicating the writer is closed.
	//
	// Returns:
	//   - error: nil if successful or already closed, error only if Close encounters a failure
	Close() error
}
