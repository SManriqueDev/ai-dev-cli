package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/ai-dev-cli/ai-dev-cli/internal/rag"
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

			provider := ragProvider
			if provider == "" {
				provider = os.Getenv("AI_PROVIDER")
				if provider == "" {
					provider = "openai"
				}
			}

			collection := ragCollection
			if collection == "" {
				collection = "ai-dev-cli-db"
			}

			ragCfg := rag.RAGConfig{
				ChromaURL:      os.Getenv("CHROMA_URL"),
				CollectionName: collection,
				Provider:       provider,
				EmbedderModel:  os.Getenv("EMBEDDER_MODEL"),
				APIKey:         os.Getenv("OPENAI_API_KEY"),
				OllamaURL:      os.Getenv("OLLAMA_BASE_URL"),
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
