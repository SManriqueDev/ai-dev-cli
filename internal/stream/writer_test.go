package stream

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============ TerminalWriter Tests ============

// TestTerminalWriter_New tests TerminalWriter creation
func TestTerminalWriter_New(t *testing.T) {
	t.Run("NewTerminalWriter returns TerminalWriter", func(t *testing.T) {
		writer := NewTerminalWriter()
		require.NotNil(t, writer)
	})

	t.Run("NewTerminalWriter implements StreamWriter interface", func(t *testing.T) {
		writer := NewTerminalWriter()
		var _ StreamWriter = writer
		require.NotNil(t, writer)
	})
}

// TestTerminalWriter_WriteChunk tests writing chunks to terminal
func TestTerminalWriter_WriteChunk(t *testing.T) {
	t.Run("WriteChunk writes chunk without error", func(t *testing.T) {
		writer := NewTerminalWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunk := []byte("test improvement")
		err := writer.WriteChunk(ctx, chunk)
		require.NoError(t, err)
	})

	t.Run("WriteChunk handles empty chunk", func(t *testing.T) {
		writer := NewTerminalWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunk := []byte{}
		err := writer.WriteChunk(ctx, chunk)
		require.NoError(t, err)
	})

	t.Run("WriteChunk respects context cancellation", func(t *testing.T) {
		writer := NewTerminalWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		chunk := []byte("test")
		err := writer.WriteChunk(ctx, chunk)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("WriteChunk returns error after Close", func(t *testing.T) {
		writer := NewTerminalWriter()
		err := writer.Close()
		require.NoError(t, err)

		ctx := context.Background()
		chunk := []byte("test")
		err = writer.WriteChunk(ctx, chunk)
		require.Error(t, err)
	})

	t.Run("WriteChunk handles large chunk", func(t *testing.T) {
		writer := NewTerminalWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunk := make([]byte, 1024*1024) // 1MB
		for i := range chunk {
			chunk[i] = byte('a')
		}
		err := writer.WriteChunk(ctx, chunk)
		require.NoError(t, err)
	})
}

// TestTerminalWriter_Close tests closing the writer
func TestTerminalWriter_Close(t *testing.T) {
	t.Run("Close succeeds on open writer", func(t *testing.T) {
		writer := NewTerminalWriter()
		err := writer.Close()
		require.NoError(t, err)
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		writer := NewTerminalWriter()
		err1 := writer.Close()
		err2 := writer.Close()
		require.NoError(t, err1)
		require.NoError(t, err2)
	})

	t.Run("WriteChunk fails after Close", func(t *testing.T) {
		writer := NewTerminalWriter()
		err := writer.Close()
		require.NoError(t, err)

		err = writer.WriteChunk(context.Background(), []byte("test"))
		require.Error(t, err)
	})
}

// TestTerminalWriter_MultipleWrites tests multiple sequential writes
func TestTerminalWriter_MultipleWrites(t *testing.T) {
	t.Run("Multiple WriteChunk calls succeed in sequence", func(t *testing.T) {
		writer := NewTerminalWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		for i := 0; i < 10; i++ {
			chunk := []byte(fmt.Sprintf("chunk %d\n", i))
			err := writer.WriteChunk(ctx, chunk)
			require.NoError(t, err)
		}
	})
}

// ============ FileWriter Tests ============

// TestFileWriter_New tests FileWriter creation
func TestFileWriter_New(t *testing.T) {
	t.Run("NewFileWriter returns FileWriter", func(t *testing.T) {
		// Use a temporary file path
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		require.NotNil(t, writer)
		_ = writer.Close()
	})

	t.Run("NewFileWriter implements StreamWriter interface", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		var _ StreamWriter = writer
		_ = writer.Close()
	})

	t.Run("NewFileWriter returns error for invalid path", func(t *testing.T) {
		// Try to create in a non-existent directory without proper path
		writer, err := NewFileWriter("/nonexistent/directory/that/does/not/exist/file.txt")
		require.Error(t, err)
		require.Nil(t, writer)
	})
}

// TestFileWriter_WriteChunk tests writing chunks to file
func TestFileWriter_WriteChunk(t *testing.T) {
	t.Run("WriteChunk writes chunk to file without error", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunk := []byte("test content")
		err = writer.WriteChunk(ctx, chunk)
		require.NoError(t, err)

		// Verify content was written to file
		// #nosec G304 - path is controlled in test
		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)
		require.Equal(t, chunk, content)
	})

	t.Run("WriteChunk handles multiple writes", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunks := [][]byte{
			[]byte("chunk1"),
			[]byte("chunk2"),
			[]byte("chunk3"),
		}

		for _, chunk := range chunks {
			err = writer.WriteChunk(ctx, chunk)
			require.NoError(t, err)
		}

		// Verify all content was written to file
		// #nosec G304 - path is controlled in test
		content, err := os.ReadFile(tmpfile)
		require.NoError(t, err)
		require.Equal(t, []byte("chunk1chunk2chunk3"), content)
	})

	t.Run("WriteChunk respects context cancellation", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		defer func() {
			_ = writer.Close()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		chunk := []byte("test")
		err = writer.WriteChunk(ctx, chunk)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("WriteChunk returns error after Close", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		err = writer.WriteChunk(context.Background(), []byte("test"))
		require.Error(t, err)
	})
}

// TestFileWriter_Close tests closing file writer
func TestFileWriter_Close(t *testing.T) {
	t.Run("Close closes file without error", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		tmpfile := filepath.Join(t.TempDir(), "test_output.txt")
		writer, err := NewFileWriter(tmpfile)
		require.NoError(t, err)

		err1 := writer.Close()
		err2 := writer.Close()
		require.NoError(t, err1)
		require.NoError(t, err2)
	})
}

// ============ MockWriter Tests ============

// TestMockWriter_New tests MockWriter creation
func TestMockWriter_New(t *testing.T) {
	t.Run("NewMockWriter returns MockWriter", func(t *testing.T) {
		writer := NewMockWriter()
		require.NotNil(t, writer)
	})

	t.Run("NewMockWriter implements StreamWriter interface", func(t *testing.T) {
		writer := NewMockWriter()
		var _ StreamWriter = writer
		require.NotNil(t, writer)
	})
}

// TestMockWriter_WriteChunk tests writing chunks to mock
func TestMockWriter_WriteChunk(t *testing.T) {
	t.Run("WriteChunk records chunk", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunk := []byte("test chunk")
		err := writer.WriteChunk(ctx, chunk)
		require.NoError(t, err)
	})

	t.Run("WriteChunk respects context cancellation", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		chunk := []byte("test")
		err := writer.WriteChunk(ctx, chunk)
		require.Equal(t, context.Canceled, err)
	})

	t.Run("WriteChunk returns error after Close", func(t *testing.T) {
		writer := NewMockWriter()
		err := writer.Close()
		require.NoError(t, err)

		err = writer.WriteChunk(context.Background(), []byte("test"))
		require.Error(t, err)
	})
}

// TestMockWriter_GetChunks tests retrieving recorded chunks
func TestMockWriter_GetChunks(t *testing.T) {
	t.Run("GetChunks returns all recorded chunks", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		ctx := context.Background()
		chunks := [][]byte{
			[]byte("chunk1"),
			[]byte("chunk2"),
			[]byte("chunk3"),
		}

		for _, chunk := range chunks {
			err := writer.WriteChunk(ctx, chunk)
			require.NoError(t, err)
		}

		recorded := writer.GetChunks()
		require.Len(t, recorded, 3)
		require.Equal(t, chunks[0], recorded[0])
		require.Equal(t, chunks[1], recorded[1])
		require.Equal(t, chunks[2], recorded[2])
	})

	t.Run("GetChunks returns empty list for no writes", func(t *testing.T) {
		writer := NewMockWriter()
		defer func() {
			_ = writer.Close()
		}()

		recorded := writer.GetChunks()
		require.Len(t, recorded, 0)
	})
}

// TestMockWriter_Close tests closing mock writer
func TestMockWriter_Close(t *testing.T) {
	t.Run("Close succeeds on open writer", func(t *testing.T) {
		writer := NewMockWriter()
		err := writer.Close()
		require.NoError(t, err)
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		writer := NewMockWriter()
		err1 := writer.Close()
		err2 := writer.Close()
		require.NoError(t, err1)
		require.NoError(t, err2)
	})
}
