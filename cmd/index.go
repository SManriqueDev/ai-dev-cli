package cmd

import (
	"context"
	"fmt"

	"github.com/ai-dev-cli/ai-dev-cli/internal/rag"
	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
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

const (
	providerOpenAI = "openai"
	providerOllama = "ollama"
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

		cfg := config.GetRAGConfig()
		if cmd.Flags().Changed("provider") {
			cfg.Provider = indexProvider
			switch indexProvider {
			case providerOpenAI:
				cfg.EmbedderModel = config.GetString("openai.embedder_model")
				cfg.APIKey = config.GetString("openai.api_key")
				cfg.BaseURL = config.GetString("openai.base_url")
			case providerOllama:
				cfg.EmbedderModel = config.GetString("ollama.embedder_model")
				cfg.BaseURL = config.GetString("ollama.base_url")
				cfg.OllamaURL = config.GetString("ollama.base_url")
			}
		}
		if cmd.Flags().Changed("collection") {
			cfg.CollectionName = indexCollection
		}
		cfg.ChunkSize = indexChunkSize
		cfg.ChunkOverlap = indexChunkOverlap

		service, err := rag.NewRAGService(ctx, rag.RAGConfig(cfg))
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
