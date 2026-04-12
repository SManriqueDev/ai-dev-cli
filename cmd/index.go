package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ai-dev-cli/ai-dev-cli/internal/rag"
	"github.com/spf13/cobra"
)

var (
	indexPath         string
	indexURL          string
	indexCollection   string
	indexChunkSize    int
	indexChunkOverlap int
	indexProvider     string
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index documents into the vector database",
	Long: `Index documents from local files, directories, or URLs into the 
ChromaDB vector database for RAG-powered code improvements.

Examples:
  # Index a directory
  ai-dev index --path ./docs

  # Index a single file
  ai-dev index --path ./README.md

  # Index a URL
  ai-dev index --url https://docs.example.com/api

  # Index with custom collection
  ai-dev index --path ./docs --collection my-docs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		provider := indexProvider
		if provider == "" {
			provider = os.Getenv("AI_PROVIDER")
			if provider == "" {
				provider = "openai"
			}
		}

		cfg := rag.RAGConfig{
			ChromaURL:      os.Getenv("CHROMA_URL"),
			CollectionName: indexCollection,
			Provider:       provider,
			EmbedderModel:  os.Getenv("EMBEDDER_MODEL"),
			APIKey:         os.Getenv("OPENAI_API_KEY"),
			OllamaURL:      os.Getenv("OLLAMA_BASE_URL"),
			ChunkSize:      indexChunkSize,
			ChunkOverlap:   indexChunkOverlap,
		}

		if cfg.CollectionName == "" {
			cfg.CollectionName = "ai-dev-cli-db"
		}

		service, err := rag.NewRAGService(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to create RAG service: %w", err)
		}

		if indexPath != "" {
			fmt.Printf("Indexing path: %s\n", indexPath)
			if err := service.IndexPath(ctx, indexPath); err != nil {
				return fmt.Errorf("failed to index path: %w", err)
			}
			fmt.Println("Indexing completed successfully")
		}

		if indexURL != "" {
			fmt.Printf("Indexing URL: %s\n", indexURL)
			if err := service.IndexURL(ctx, indexURL); err != nil {
				return fmt.Errorf("failed to index URL: %w", err)
			}
			fmt.Println("URL indexing completed successfully")
		}

		if indexPath == "" && indexURL == "" {
			return fmt.Errorf("either --path or --url must be specified")
		}

		return nil
	},
}

func init() {
	indexCmd.Flags().StringVar(&indexPath, "path", "", "Path to file or directory to index")
	indexCmd.Flags().StringVar(&indexURL, "url", "", "URL to index")
	indexCmd.Flags().StringVar(&indexCollection, "collection", "ai-dev-cli-db", "Collection name")
	indexCmd.Flags().IntVar(&indexChunkSize, "chunk-size", 1000, "Chunk size for text splitting")
	indexCmd.Flags().IntVar(&indexChunkOverlap, "chunk-overlap", 200, "Chunk overlap for text splitting")
	indexCmd.Flags().StringVar(&indexProvider, "provider", "", "Provider: openai or ollama (default: from AI_PROVIDER env)")

	rootCmd.AddCommand(indexCmd)
}
