package rag

import (
	"context"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

type EmbedderConfig struct {
	Provider  string
	Model     string
	APIKey    string
	BaseURL   string
	OllamaURL string
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
		modelName = "text-embedding-3-small"
	}

	opts := []openai.Option{
		openai.WithModel(modelName),
	}
	if cfg.APIKey != "" {
		opts = append(opts, openai.WithToken(cfg.APIKey))
	}
	if cfg.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
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
		modelName = "nomic-embed-text"
	}

	ollamaURL := cfg.OllamaURL
	if ollamaURL == "" {
		ollamaURL = os.Getenv("OLLAMA_BASE_URL")
		if ollamaURL == "" {
			ollamaURL = "http://localhost:11434"
		}
	}

	opts := []openai.Option{
		openai.WithModel(modelName),
		openai.WithBaseURL(ollamaURL + "/v1"),
		openai.WithToken("ollama"),
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}

	return embeddings.NewEmbedder(llm)
}
