package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/ai-dev-cli/ai-dev-cli/internal/output"
	"github.com/ai-dev-cli/ai-dev-cli/internal/rag"
	"github.com/ai-dev-cli/ai-dev-cli/internal/stream"
	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	useRAG        bool
	ragCollection string
	ragProvider   string
	useStream     bool
	outputFormat  string
	outputTheme   string
	outputFile    string
	showDiff      bool
	applyMode     string
)

var improveCmd = &cobra.Command{
	Use:   "improve <file>",
	Short: "Improve code quality using AI",
	Long: `Analyzes the given Go file and suggests improvements for code quality, performance, and best practices.

Use --rag to enable RAG-powered context-aware improvements that use your indexed documentation.
Use --stream to stream improvements in real-time as they are generated.`,
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
		if useStream {
			fmt.Printf("Streaming improvements for %s...\n\n", cleanPath)
			return handleImproveStreaming(cmd, cleanPath, string(content), prompter)
		}

		// Original non-streaming behavior
		var result string
		if useRAG {
			ctx := context.Background()

			ragCfg := rag.RAGConfig(config.GetRAGConfig())
			if cmd.Flags().Changed("provider") && ragProvider != "" {
				ragCfg.Provider = ragProvider
				switch ragProvider {
				case "openai":
					ragCfg.EmbedderModel = config.GetString("openai.embedder_model")
					ragCfg.APIKey = config.GetString("openai.api_key")
					ragCfg.BaseURL = config.GetString("openai.base_url")
				case "ollama":
					ragCfg.EmbedderModel = config.GetString("ollama.embedder_model")
					ragCfg.BaseURL = config.GetString("ollama.base_url")
					ragCfg.OllamaURL = config.GetString("ollama.base_url")
				}
			}
			if cmd.Flags().Changed("collection") {
				ragCfg.CollectionName = ragCollection
			}

			ragService, err := rag.NewRAGService(ctx, ragCfg)
			if err != nil {
				return fmt.Errorf("failed to create RAG service: %w", err)
			}

			contextStr, err := ragService.SearchContext(ctx, fmt.Sprintf("Improve code: %s", filePath))

			if err != nil {
				fmt.Printf("Warning: failed to get RAG context: %v\n", err)
				contextStr = ""
			} else {
				fmt.Println("Using RAG context from vector database")
			}
			result, err = prompter.ImproveCodeWithContext(string(content), contextStr)
			if err != nil {
				return fmt.Errorf("failed to improve code with RAG context: %w", err)
			}
		} else {
			result, err = prompter.ImproveCode(string(content))
		}

		if err != nil {
			return fmt.Errorf("failed to improve code: %w", err)
		}

		originalContent := string(content)

		if showDiff {
			diffOutput := output.FormatDiff(originalContent, result, isTerminal())
			fmt.Println(diffOutput)
		} else {
			format := output.Format(outputFormat)
			if format == "" {
				format = output.FormatMarkdown
			}
			theme := output.Theme(outputTheme)
			if theme == "" {
				theme = output.ThemeDark
			}

			if outputFile != "" {
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

			if outputFile != "" {
				err = os.WriteFile(outputFile, []byte(out), 0o600)
				if err != nil {
					return fmt.Errorf("failed to write output to %s: check directory exists and you have write permissions: %w", outputFile, err)
				}
				fmt.Printf("Saved to %s\n", outputFile)
			} else {
				fmt.Println(out)
			}
		}

		if applyMode != "" {
			shouldApply := false
			if applyMode == "ask" {
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

// spinnerStoppingWriter wraps a StreamWriter and stops the spinner on first chunk.
type spinnerStoppingWriter struct {
	inner    stream.StreamWriter
	progress *stream.ProgressIndicator
	once     *sync.Once
}

func (w *spinnerStoppingWriter) WriteChunk(ctx context.Context, chunk []byte) error {
	// Stop spinner on first chunk
	w.once.Do(func() {
		w.progress.Stop()
	})
	return w.inner.WriteChunk(ctx, chunk)
}

func (w *spinnerStoppingWriter) Close() error {
	w.progress.Stop()
	return w.inner.Close()
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func init() {
	improveCmd.Flags().BoolVar(&useRAG, "rag", false, "Enable RAG-powered context-aware improvements")
	improveCmd.Flags().StringVar(&ragCollection, "collection", "ai-dev-cli-db", "RAG collection name")
	improveCmd.Flags().StringVar(&ragProvider, "provider", "", "RAG provider: openai or ollama")
	improveCmd.Flags().BoolVar(&useStream, "stream", false, "Stream improvements in real-time as they are generated")
	improveCmd.Flags().StringVar(&outputFormat, "format", "markdown", "Output format: markdown, plain, json, yaml")
	improveCmd.Flags().StringVar(&outputTheme, "theme", "dark", "Glamour theme: dark, light, auto")
	improveCmd.Flags().StringVar(&outputFile, "output", "", "Output file path (default: stdout)")
	improveCmd.Flags().BoolVar(&showDiff, "diff", false, "Show diff between original and improved code")
	improveCmd.Flags().StringVar(&applyMode, "apply", "", "Apply changes to source file: use 'yes' for auto-apply, 'ask' for confirmation prompt")
}

// handleImproveStreaming handles the streaming improve command execution.
func handleImproveStreaming(cmd *cobra.Command, filePath, content string, prompter *ai.Prompter) error {
	// Create stream context
	streamCtx := stream.NewStreamContext("improve", filePath)
	startTime := streamCtx.StartedAt
	defer streamCtx.CompleteSuccessfully()

	// Create and start progress indicator (animated spinner)
	progress := stream.NewProgressIndicator("⠋ Generating improvements...")
	progress.Start()

	// Create terminal writer for streaming output
	baseWriter := stream.NewTerminalWriter()
	defer func() {
		_ = baseWriter.Close()
	}()

	// Wrap with tracking writer to record metrics
	trackingWriter := stream.NewTrackingWriter(baseWriter, streamCtx)

	// Setup interrupt handler for Ctrl+C
	handler := stream.NewInterruptHandler(streamCtx)
	handler.RegisterCleanupFunc(baseWriter.Close)
	defer handler.Stop()

	// Create wrapper that stops spinner on first chunk
	wrappedWriter := &spinnerStoppingWriter{
		inner:    trackingWriter,
		progress: progress,
		once:     &sync.Once{},
	}

	// Execute streaming improve
	if useRAG {
		ctx := context.Background()

		ragCfg := rag.RAGConfig(config.GetRAGConfig())
		if cmd.Flags().Changed("provider") && ragProvider != "" {
			ragCfg.Provider = ragProvider
			switch ragProvider {
			case "openai":
				ragCfg.EmbedderModel = config.GetString("openai.embedder_model")
				ragCfg.APIKey = config.GetString("openai.api_key")
				ragCfg.BaseURL = config.GetString("openai.base_url")
			case "ollama":
				ragCfg.EmbedderModel = config.GetString("ollama.embedder_model")
				ragCfg.BaseURL = config.GetString("ollama.base_url")
				ragCfg.OllamaURL = config.GetString("ollama.base_url")
			}
		}
		if cmd.Flags().Changed("collection") {
			ragCfg.CollectionName = ragCollection
		}

		ragService, err := rag.NewRAGService(ctx, ragCfg)
		if err != nil {
			progress.Stop()
			return fmt.Errorf("failed to create RAG service: %w", err)
		}

		contextStr, err := ragService.SearchContext(ctx, fmt.Sprintf("Improve code: %s", filePath))
		if err != nil {
			fmt.Printf("Warning: failed to get RAG context: %v\n", err)
			contextStr = ""
		} else {
			fmt.Println("Using RAG context from vector database")
		}

		err = prompter.ImproveCodeStreamWithContext(streamCtx.Context(), content, contextStr, wrappedWriter)
		if err != nil {
			progress.Stop()
			streamCtx.RecordError(err.Error())
			return fmt.Errorf("failed to stream code improvement with RAG context: %w", err)
		}
	} else {
		err := prompter.ImproveCodeStream(streamCtx.Context(), content, wrappedWriter)
		if err != nil {
			progress.Stop()
			streamCtx.RecordError(err.Error())
			return fmt.Errorf("failed to stream code improvement: %w", err)
		}
	}

	// Ensure spinner is stopped
	progress.Stop()

	// Skip metrics logging if interrupted
	if streamCtx.Status != stream.StatusInterrupted {
		// Log streaming metrics only for successful completions
		metrics := trackingWriter.GetMetrics()
		if metrics.HasFirstChunkTime {
			timeSinceStart := metrics.FirstChunkTime.Sub(startTime).Seconds()
			fmt.Printf("\nStreaming complete: %d chunks, %d bytes, first chunk in %.2fs\n",
				metrics.ChunkCount, metrics.BytesReceived, timeSinceStart)
		}
	}

	return nil
}
