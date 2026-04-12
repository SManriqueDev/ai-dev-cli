package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/ai-dev-cli/ai-dev-cli/internal/rag"
	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
	"github.com/spf13/cobra"
)

var (
	useRAG        bool
	ragCollection string
	ragProvider   string
)

var improveCmd = &cobra.Command{
	Use:   "improve <file>",
	Short: "Improve code quality using AI",
	Long: `Analyzes the given Go file and suggests improvements for code quality, performance, and best practices.

Use --rag to enable RAG-powered context-aware improvements that use your indexed documentation.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client, err := ai.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create AI client: %w", err)
		}

		prompter := ai.NewPrompter(client)

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
		} else {
			result, err = prompter.ImproveCode(string(content))
		}

		if err != nil {
			return fmt.Errorf("failed to improve code: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}

func init() {
	improveCmd.Flags().BoolVar(&useRAG, "rag", false, "Enable RAG-powered context-aware improvements")
	improveCmd.Flags().StringVar(&ragCollection, "collection", "ai-dev-cli-db", "RAG collection name")
	improveCmd.Flags().StringVar(&ragProvider, "provider", "", "RAG provider: openai or ollama")
}
