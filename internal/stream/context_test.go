package stream

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============ StreamContext Tests ============

// TestStreamContext_New tests StreamContext creation
func TestStreamContext_New(t *testing.T) {
	t.Run("NewStreamContext returns valid context", func(t *testing.T) {
		ctx := NewStreamContext("improve", "test.go")
		require.NotNil(t, ctx)
		require.Equal(t, "improve", ctx.Command)
		require.Equal(t, "test.go", ctx.FilePath)
		require.Equal(t, StatusRunning, ctx.Status)
		require.Equal(t, 0, ctx.ChunkCount)
		require.Equal(t, int64(0), ctx.BytesReceived)
	})

	t.Run("NewStreamContext generates unique OperationID", func(t *testing.T) {
		ctx1 := NewStreamContext("improve", "test.go")
		time.Sleep(10 * time.Millisecond) // Ensure different timestamp
		ctx2 := NewStreamContext("improve", "test.go")
		require.NotEqual(t, ctx1.OperationID, ctx2.OperationID)
	})
}

// TestStreamContext_Context tests context retrieval
func TestStreamContext_Context(t *testing.T) {
	t.Run("Context returns valid context", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		ctx := sc.Context()
		require.NotNil(t, ctx)
		require.NoError(t, ctx.Err())
	})

	t.Run("Context can be cancelled", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		sc.Cancel()
		ctx := sc.Context()
		require.Error(t, ctx.Err())
	})
}

// TestStreamContext_RecordChunk tests recording chunks
func TestStreamContext_RecordChunk(t *testing.T) {
	t.Run("RecordChunk increments count", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		require.Equal(t, 0, sc.ChunkCount)

		sc.RecordChunk(100)
		require.Equal(t, 1, sc.ChunkCount)
		require.Equal(t, int64(100), sc.BytesReceived)

		sc.RecordChunk(50)
		require.Equal(t, 2, sc.ChunkCount)
		require.Equal(t, int64(150), sc.BytesReceived)
	})

	t.Run("RecordChunk stops after status change", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		sc.RecordChunk(100)
		sc.CompleteSuccessfully()
		sc.RecordChunk(50)

		// Second chunk should not be recorded
		require.Equal(t, 1, sc.ChunkCount)
		require.Equal(t, int64(100), sc.BytesReceived)
	})
}

// TestStreamContext_CompleteSuccessfully tests successful completion
func TestStreamContext_CompleteSuccessfully(t *testing.T) {
	t.Run("CompleteSuccessfully sets status", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		require.Equal(t, StatusRunning, sc.Status)

		sc.CompleteSuccessfully()
		require.Equal(t, StatusCompleted, sc.Status)
	})

	t.Run("CompleteSuccessfully is idempotent", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		sc.CompleteSuccessfully()
		sc.CompleteSuccessfully()
		require.Equal(t, StatusCompleted, sc.Status)
	})
}

// TestStreamContext_Interrupt tests interruption
func TestStreamContext_Interrupt(t *testing.T) {
	t.Run("Interrupt sets status", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		require.Equal(t, StatusRunning, sc.Status)

		sc.Interrupt()
		require.Equal(t, StatusInterrupted, sc.Status)
		require.NotNil(t, sc.InterruptedAt)
	})

	t.Run("Interrupt cancels context", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		ctx := sc.Context()
		require.NoError(t, ctx.Err())

		sc.Interrupt()
		ctx = sc.Context()
		require.Error(t, ctx.Err())
	})
}

// TestStreamContext_RecordError tests error recording
func TestStreamContext_RecordError(t *testing.T) {
	t.Run("RecordError sets status and message", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		require.Equal(t, StatusRunning, sc.Status)

		sc.RecordError("network timeout")
		require.Equal(t, StatusFailed, sc.Status)
		require.Equal(t, "network timeout", sc.ErrorMessage)
	})

	t.Run("RecordError does not change non-running status", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		sc.CompleteSuccessfully()
		sc.RecordError("error after completion")

		require.Equal(t, StatusCompleted, sc.Status)
		require.Empty(t, sc.ErrorMessage)
	})
}

// TestStreamContext_GetSnapshot tests snapshot creation
func TestStreamContext_GetSnapshot(t *testing.T) {
	t.Run("GetSnapshot returns accurate data", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		sc.RecordChunk(100)

		snapshot := sc.GetSnapshot()
		require.Equal(t, "improve", snapshot.Command)
		require.Equal(t, "test.go", snapshot.FilePath)
		require.Equal(t, StatusRunning, snapshot.Status)
		require.Equal(t, 1, snapshot.ChunkCount)
		require.Equal(t, int64(100), snapshot.BytesReceived)
	})
}

// ============ InterruptHandler Tests ============

// TestInterruptHandler_New tests InterruptHandler creation
func TestInterruptHandler_New(t *testing.T) {
	t.Run("NewInterruptHandler returns valid handler", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)
		require.NotNil(t, handler)
	})
}

// TestInterruptHandler_RegisterCancelFunc tests registering cancel functions
func TestInterruptHandler_RegisterCancelFunc(t *testing.T) {
	t.Run("RegisterCancelFunc accepts cancel functions", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)

		ctx, cancel := context.WithCancel(context.Background())
		require.NoError(t, ctx.Err())

		handler.RegisterCancelFunc(cancel)
		defer handler.Stop()
	})

	t.Run("RegisterCancelFunc handles nil gracefully", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)
		defer handler.Stop()

		handler.RegisterCancelFunc(nil) // Should not panic
	})
}

// TestInterruptHandler_RegisterCleanupFunc tests registering cleanup functions
func TestInterruptHandler_RegisterCleanupFunc(t *testing.T) {
	t.Run("RegisterCleanupFunc accepts cleanup functions", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)
		defer handler.Stop()

		_ = false // Using a dummy variable since we're just testing registration
		handler.RegisterCleanupFunc(func() error {
			_ = true // Cleanup would be called on interrupt
			return nil
		})
	})

	t.Run("RegisterCleanupFunc handles nil gracefully", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)
		defer handler.Stop()

		handler.RegisterCleanupFunc(nil) // Should not panic
	})
}

// TestInterruptHandler_Stop tests stopping the handler
func TestInterruptHandler_Stop(t *testing.T) {
	t.Run("Stop is safe to call multiple times", func(t *testing.T) {
		sc := NewStreamContext("improve", "test.go")
		handler := NewInterruptHandler(sc)

		handler.Stop()
		handler.Stop() // Should not panic
	})
}

// ============ StreamResult Tests ============

// TestStreamResult_New tests StreamResult creation
func TestStreamResult_New(t *testing.T) {
	t.Run("NewStreamResult returns valid result", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte("test content"), ContentTypeText)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, result.SequenceNumber)
		require.Equal(t, []byte("test content"), result.Content)
		require.Equal(t, ContentTypeText, result.ContentType)
		require.Equal(t, 12, result.Size)
	})

	t.Run("NewStreamResult validates sequence number", func(t *testing.T) {
		_, err := NewStreamResult(0, []byte("test"), ContentTypeText)
		require.Error(t, err)
		require.Contains(t, err.Error(), "sequence number")
	})

	t.Run("NewStreamResult validates content", func(t *testing.T) {
		_, err := NewStreamResult(1, []byte{}, ContentTypeText)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty")
	})

	t.Run("NewStreamResult validates content type", func(t *testing.T) {
		_, err := NewStreamResult(1, []byte("test"), ContentType("invalid"))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid content type")
	})

	t.Run("NewStreamResult defaults to text content type", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte("test"), "")
		require.NoError(t, err)
		require.Equal(t, ContentTypeText, result.ContentType)
	})
}

// TestStreamResult_String tests string representation
func TestStreamResult_String(t *testing.T) {
	t.Run("String returns formatted representation", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte("test"), ContentTypeText)
		require.NoError(t, err)

		str := result.String()
		require.Contains(t, str, "seq=1")
		require.Contains(t, str, "type=text")
		require.Contains(t, str, "size=4")
	})
}

// TestStreamResult_ContentTypes tests different content types
func TestStreamResult_ContentTypes(t *testing.T) {
	t.Run("Creates result with JSON content type", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte(`{"key":"value"}`), ContentTypeJSON)
		require.NoError(t, err)
		require.Equal(t, ContentTypeJSON, result.ContentType)
	})

	t.Run("Creates result with error content type", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte("error message"), ContentTypeError)
		require.NoError(t, err)
		require.Equal(t, ContentTypeError, result.ContentType)
	})

	t.Run("Creates result with progress content type", func(t *testing.T) {
		result, err := NewStreamResult(1, []byte("50% done"), ContentTypeProgress)
		require.NoError(t, err)
		require.Equal(t, ContentTypeProgress, result.ContentType)
	})
}

// TestStreamResult_ImmutableContent tests immutability
func TestStreamResult_ImmutableContent(t *testing.T) {
	t.Run("StreamResult maintains content integrity", func(t *testing.T) {
		original := []byte("test content")
		result, err := NewStreamResult(1, original, ContentTypeText)
		require.NoError(t, err)

		// Content should be stored as-is
		require.Equal(t, original, result.Content)
		require.Equal(t, len(original), result.Size)
	})
}

// ============ MockWriter Helper Functions Tests ============

// TestMockWriter_GetChunkCount tests getting chunk count
func TestMockWriter_GetChunkCount(t *testing.T) {
	t.Run("GetChunkCount returns correct count", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		require.Equal(t, 0, writer.GetChunkCount())

		ctx := context.Background()
		_ = writer.WriteChunk(ctx, []byte("chunk1"))
		require.Equal(t, 1, writer.GetChunkCount())

		_ = writer.WriteChunk(ctx, []byte("chunk2"))
		require.Equal(t, 2, writer.GetChunkCount())
	})
}

// TestMockWriter_CombinedContent tests combined content retrieval
func TestMockWriter_CombinedContent(t *testing.T) {
	t.Run("CombinedContent returns concatenated chunks", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		_ = writer.WriteChunk(ctx, []byte("hello"))
		_ = writer.WriteChunk(ctx, []byte(" "))
		_ = writer.WriteChunk(ctx, []byte("world"))

		combined := writer.CombinedContent()
		require.Equal(t, []byte("hello world"), combined)
	})

	t.Run("CombinedContent returns empty for no writes", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		combined := writer.CombinedContent()
		require.Empty(t, combined)
	})
}
