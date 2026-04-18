package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestTrackingWriter_New tests TrackingWriter creation.
func TestTrackingWriter_New(t *testing.T) {
	t.Run("NewTrackingWriter returns valid instance", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")

		writer := NewTrackingWriter(underlying, streamCtx)

		require.NotNil(t, writer)
		require.Equal(t, underlying, writer.underlying)
		require.Equal(t, streamCtx, writer.streamCtx)
	})
}

// TestTrackingWriter_WriteChunk tests chunk writing with tracking.
func TestTrackingWriter_WriteChunk(t *testing.T) {
	t.Run("WriteChunk records metrics", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		testChunks := [][]byte{
			[]byte("chunk1"),
			[]byte("chunk2"),
			[]byte("chunk3"),
		}

		for _, chunk := range testChunks {
			err := writer.WriteChunk(context.Background(), chunk)
			require.NoError(t, err)
		}

		metrics := writer.GetMetrics()
		require.Equal(t, 3, metrics.ChunkCount)
		require.Equal(t, int64(18), metrics.BytesReceived)
		require.True(t, metrics.HasFirstChunkTime)
	})

	t.Run("WriteChunk records first chunk time", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		startTime := time.Now()

		err := writer.WriteChunk(context.Background(), []byte("first"))
		require.NoError(t, err)

		metrics := writer.GetMetrics()
		require.True(t, metrics.HasFirstChunkTime)
		require.True(t, metrics.FirstChunkTime.After(startTime) || metrics.FirstChunkTime.Equal(startTime))
	})

	t.Run("WriteChunk respects context cancellation", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := writer.WriteChunk(ctx, []byte("chunk"))
		require.Error(t, err)
	})

	t.Run("WriteChunk delegates to underlying writer", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		testChunk := []byte("test chunk")
		err := writer.WriteChunk(context.Background(), testChunk)
		require.NoError(t, err)

		mockChunks := underlying.GetChunks()
		require.Equal(t, 1, len(mockChunks))
		require.Equal(t, testChunk, mockChunks[0])
	})
}

// TestTrackingWriter_GetMetrics tests metrics retrieval.
func TestTrackingWriter_GetMetrics(t *testing.T) {
	t.Run("GetMetrics returns empty metrics when no chunks written", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		metrics := writer.GetMetrics()
		require.Equal(t, 0, metrics.ChunkCount)
		require.Equal(t, int64(0), metrics.BytesReceived)
		require.False(t, metrics.HasFirstChunkTime)
	})

	t.Run("GetMetrics returns correct metrics after writes", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		chunks := [][]byte{[]byte("a"), []byte("bb"), []byte("ccc")}
		for _, chunk := range chunks {
			_ = writer.WriteChunk(context.Background(), chunk)
		}

		metrics := writer.GetMetrics()
		require.Equal(t, 3, metrics.ChunkCount)
		require.Equal(t, int64(6), metrics.BytesReceived) // 1 + 2 + 3
	})
}

// TestTrackingWriter_Close tests closing behavior.
func TestTrackingWriter_Close(t *testing.T) {
	t.Run("Close delegates to underlying writer", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		err := writer.Close()
		require.NoError(t, err)

		// Verify underlying writer is closed
		err = underlying.WriteChunk(context.Background(), []byte("should fail"))
		require.Error(t, err)
	})
}

// TestTrackingWriter_StreamContextIntegration tests integration with StreamContext.
func TestTrackingWriter_StreamContextIntegration(t *testing.T) {
	t.Run("WriteChunk records chunks in StreamContext", func(t *testing.T) {
		underlying := NewMockWriter()
		streamCtx := NewStreamContext("test", "file.go")
		writer := NewTrackingWriter(underlying, streamCtx)

		err := writer.WriteChunk(context.Background(), []byte("chunk1"))
		require.NoError(t, err)

		err = writer.WriteChunk(context.Background(), []byte("chunk2"))
		require.NoError(t, err)

		snapshot := streamCtx.GetSnapshot()
		require.Equal(t, 2, snapshot.ChunkCount)
		require.Equal(t, int64(12), snapshot.BytesReceived)
	})
}
