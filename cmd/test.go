package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"charm.land/glamour/v2"
	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/ai-dev-cli/ai-dev-cli/internal/stream"
	"github.com/spf13/cobra"
)

var useTestStream bool

var testCmd = &cobra.Command{
	Use:   "test <file>",
	Short: "Generate unit tests for the given file",
	Long: `Analyzes the given Go file and generates comprehensive unit tests using AI.

Use --stream to stream test generation in real-time as tests are generated.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		cleanPath := filepath.Clean(filePath)
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client, err := ai.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create AI client: %w", err)
		}

		prompter := ai.NewPrompter(client)

		// Use streaming if --stream flag is enabled
		if useTestStream {
			return handleTestStreaming(cmd, cleanPath, string(content), prompter)
		}

		// Original non-streaming behavior
		result, err := prompter.GenerateTests(string(content))
		if err != nil {
			return fmt.Errorf("failed to generate tests: %w", err)
		}

		out, err := glamour.Render(result, "dark")
		if err != nil {
			return fmt.Errorf("failed to render output: %w", err)
		}
		fmt.Println(out)

		return nil
	},
}

func init() {
	testCmd.Flags().BoolVar(&useTestStream, "stream", false, "Stream test generation in real-time as tests are generated")
}

// testSpinnerStoppingWriter wraps a StreamWriter and stops the spinner on first chunk.
type testSpinnerStoppingWriter struct {
	inner    stream.StreamWriter
	progress *stream.ProgressIndicator
	once     *sync.Once
}

func (w *testSpinnerStoppingWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	// Stop spinner on first chunk
	w.once.Do(func() {
		w.progress.Stop()
	})
	return w.inner.WriteChunk(ctx, chunk)
}

func (w *testSpinnerStoppingWriter) Close() error {
	w.progress.Stop()
	return w.inner.Close()
}

// handleTestStreaming handles the streaming test generation command execution.
func handleTestStreaming(_ *cobra.Command, filePath, content string, prompter *ai.Prompter) error {
	// Create stream context
	streamCtx := stream.NewStreamContext("test", filePath)
	startTime := streamCtx.StartedAt
	defer streamCtx.CompleteSuccessfully()

	// Create and start progress indicator (animated spinner)
	progress := stream.NewProgressIndicator("⠋ Generating tests...")
	progress.Start()

	// Create terminal writer for streaming output
	baseWriter := stream.NewTerminalWriter()
	defer func() {
		_ = baseWriter.Close()
	}()

	// Wrap with tracking writer to record metrics
	trackingWriter := stream.NewTrackingWriter(baseWriter, streamCtx)

	// Create wrapper that stops spinner on first chunk
	wrappedWriter := &testSpinnerStoppingWriter{
		inner:    trackingWriter,
		progress: progress,
		once:     &sync.Once{},
	}

	// Setup interrupt handler for Ctrl+C
	handler := stream.NewInterruptHandler(streamCtx)
	handler.RegisterCleanupFunc(baseWriter.Close)
	defer handler.Stop()

	// Execute streaming test generation
	err := prompter.GenerateTestsStream(streamCtx.Context(), content, wrappedWriter)
	if err != nil {
		progress.Stop()
		streamCtx.RecordError(err.Error())
		return fmt.Errorf("failed to stream test generation: %w", err)
	}

	// Ensure spinner is stopped
	progress.Stop()

	// Skip metrics logging if interrupted
	if streamCtx.Status != stream.StatusInterrupted {
		// Log streaming metrics only for successful completions
		metrics := trackingWriter.GetMetrics()
		if metrics.HasFirstChunkTime {
			timeSinceStart := metrics.FirstChunkTime.Sub(startTime).Seconds()
			fmt.Printf("\nTest generation complete: %d chunks, %d bytes, first chunk in %.2fs\n",
				metrics.ChunkCount, metrics.BytesReceived, timeSinceStart)
		}
	}

	return nil
}
