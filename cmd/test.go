package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/ai-dev-cli/ai-dev-cli/internal/output"
	"github.com/ai-dev-cli/ai-dev-cli/internal/stream"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	useTestStream    bool
	testOutputFormat string
	testOutputTheme  string
	testOutputFile   string
	testShowDiff     bool
	testApplyMode    string
)

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
			return fmt.Errorf("failed to read file %s: check that the file exists and you have read permissions: %w", cleanPath, err)
		}

		client, err := ai.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create AI client: check your OPENAI_API_KEY or OLLAMA_BASE_URL environment variables: %w", err)
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

		originalContent := string(content)

		if testShowDiff {
			diffOutput := output.FormatDiff(originalContent, result, testIsTerminal())
			fmt.Println(diffOutput)
		} else {
			format := output.Format(testOutputFormat)
			if format == "" {
				format = output.FormatMarkdown
			}
			theme := output.Theme(testOutputTheme)
			if theme == "" {
				theme = output.ThemeDark
			}

			if testOutputFile != "" {
				format = output.FormatPlain
			}

			formatter := output.NewFormatter(format, theme)
			out, err := formatter.Format(output.OutputData{
				Content:  result,
				Original: originalContent,
				Format:   string(format),
			})
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			if testOutputFile != "" {
				err = os.WriteFile(testOutputFile, []byte(out), 0o600)
				if err != nil {
					return fmt.Errorf("failed to write output to %s: check directory exists and you have write permissions: %w", testOutputFile, err)
				}
				fmt.Printf("Saved to %s\n", testOutputFile)
			} else {
				fmt.Println(out)
			}
		}

		if testApplyMode != "" {
			shouldApply := false
			if testApplyMode == "ask" {
				fmt.Printf("Apply changes to %s? [y/N]: ", cleanPath)
				var response string
				_, _ = fmt.Scanln(&response)
				shouldApply = response == "y" || response == "Y"
			} else {
				shouldApply = true
			}

			if !shouldApply {
				fmt.Println("Apply cancelled.")
				return nil
			}

			codeToApply := output.ExtractCodeFromMarkdown(result)
			backupPath := cleanPath + ".bak"
			err = os.WriteFile(backupPath, []byte(originalContent), 0o600)
			if err != nil {
				return fmt.Errorf("failed to create backup at %s: check write permissions in directory: %w", backupPath, err)
			}
			err = os.WriteFile(cleanPath, []byte(codeToApply), 0o600)
			if err != nil {
				return fmt.Errorf("failed to apply changes to %s: check write permissions: %w", cleanPath, err)
			}

			summary := output.ComputeChangeSummary(originalContent, codeToApply)
			fmt.Printf("Applied changes to %s (backup: %s)\n", cleanPath, backupPath)
			fmt.Printf("Changes: +%d lines, -%d lines\n", summary.LinesAdded, summary.LinesRemoved)
		}

		return nil
	},
}

func init() {
	testCmd.Flags().BoolVar(&useTestStream, "stream", false, "Stream test generation in real-time as tests are generated")
	testCmd.Flags().StringVar(&testOutputFormat, "format", "markdown", "Output format: markdown, plain, json, yaml")
	testCmd.Flags().StringVar(&testOutputTheme, "theme", "dark", "Glamour theme: dark, light, auto")
	testCmd.Flags().StringVar(&testOutputFile, "output", "", "Output file path (default: stdout)")
	testCmd.Flags().BoolVar(&testShowDiff, "diff", false, "Show diff between original and generated tests")
	testCmd.Flags().StringVar(&testApplyMode, "apply", "", "Apply changes to source file: use 'yes' for auto-apply, 'ask' for confirmation prompt")
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

func testIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
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
