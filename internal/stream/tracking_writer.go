package stream

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TrackingWriter wraps a StreamWriter and tracks metrics like chunk count, bytes, and first chunk timing.
type TrackingWriter struct {
	underlying     StreamWriter
	streamCtx      *StreamContext
	mu             sync.Mutex
	firstChunkTime *time.Time
	chunkCount     int
	bytesReceived  int64
}

// NewTrackingWriter creates a new TrackingWriter that wraps the given StreamWriter.
func NewTrackingWriter(underlying StreamWriter, streamCtx *StreamContext) *TrackingWriter {
	return &TrackingWriter{
		underlying:    underlying,
		streamCtx:     streamCtx,
		chunkCount:    0,
		bytesReceived: 0,
	}
}

// WriteChunk writes a chunk to the underlying writer and records metrics.
func (tw *TrackingWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	tw.mu.Lock()

	// Record first chunk time
	if tw.firstChunkTime == nil {
		now := time.Now()
		tw.firstChunkTime = &now
	}

	tw.chunkCount++
	tw.bytesReceived += int64(len(chunk))

	tw.mu.Unlock()

	// Record in stream context
	tw.streamCtx.RecordChunk(len(chunk))

	// Write to underlying writer
	if err := tw.underlying.WriteChunk(ctx, chunk); err != nil {
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	return nil
}

// Close closes the underlying writer.
func (tw *TrackingWriter) Close() error {
	return tw.underlying.Close()
}

// GetMetrics returns the current tracking metrics.
func (tw *TrackingWriter) GetMetrics() TrackingMetrics {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	metrics := TrackingMetrics{
		ChunkCount:    tw.chunkCount,
		BytesReceived: tw.bytesReceived,
	}

	if tw.firstChunkTime != nil {
		metrics.FirstChunkTime = *tw.firstChunkTime
		metrics.HasFirstChunkTime = true
	}

	return metrics
}

// TrackingMetrics represents tracking statistics for a streaming operation.
type TrackingMetrics struct {
	ChunkCount        int
	BytesReceived     int64
	FirstChunkTime    time.Time
	HasFirstChunkTime bool
}
