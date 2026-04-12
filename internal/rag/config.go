package rag

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type EmbedderConfig struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	OllamaURL   string
	OllamaModel string
}

func NewEmbedder(ctx context.Context, cfg EmbedderConfig) (embeddings.Embedder, error) {
	switch cfg.Provider {
	case "openai":
		return newOpenAIEmbedder(ctx, cfg)
	case "ollama":
		return newOllamaEmbedder(ctx, cfg)
	default:
		return newOpenAIEmbedder(ctx, cfg)
	}
}

func newOpenAIEmbedder(ctx context.Context, cfg EmbedderConfig) (embeddings.Embedder, error) {
	modelName := cfg.Model
	if modelName == "" {
		modelName = getEnvWithDefault("OPENAI_EMBEDDER_MODEL", "text-embedding-3-small")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("OPENAI_BASE_URL")
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
	}

	opts := []openai.Option{
		openai.WithModel(modelName),
		openai.WithBaseURL(baseURL),
	}
	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return embeddings.NewEmbedder(llm)
}

func newOllamaEmbedder(ctx context.Context, cfg EmbedderConfig) (embeddings.Embedder, error) {
	modelName := cfg.Model
	if modelName == "" {
		modelName = getEnvWithDefault("OLLAMA_EMBEDDER_MODEL", "nomic-embed-text")
	}

	ollamaURL := cfg.OllamaURL
	if ollamaURL == "" {
		ollamaURL = os.Getenv("OLLAMA_BASE_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
	}

	fmt.Printf("DEBUG: Creating Ollama embedder with model=%s, url=%s\n", modelName, ollamaURL)

	opts := []ollama.Option{
		ollama.WithModel(modelName),
		ollama.WithServerURL(ollamaURL),
	}

	llm, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}

	emb, err := embeddings.NewEmbedder(llm)
	if err != nil {
		return nil, err
	}

	fmt.Printf("DEBUG: Ollama embedder created successfully\n")
	return emb, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
