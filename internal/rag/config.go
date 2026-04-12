package rag

import (
	"context"
	"fmt"

	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
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
	if cfg.Provider == "" {
		cfg.Provider = config.DefaultProvider
	}

	switch cfg.Provider {
	case "openai":
		return newOpenAIEmbedder(cfg)
	case "ollama":
		return newOllamaEmbedder(cfg)
	default:
		return nil, fmt.Errorf("unsupported embedder provider %q", cfg.Provider)
	}
}

func newOpenAIEmbedder(cfg EmbedderConfig) (embeddings.Embedder, error) {
	modelName := cfg.Model
	if modelName == "" {
		modelName = config.DefaultOpenAIEmbedderModel
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = config.DefaultOpenAIBaseURL
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

func newOllamaEmbedder(cfg EmbedderConfig) (embeddings.Embedder, error) {
	modelName := cfg.Model
	if modelName == "" {
		modelName = config.DefaultOllamaEmbedderModel
	}

	ollamaURL := cfg.OllamaURL
	if ollamaURL == "" {
		ollamaURL = config.DefaultOllamaBaseURL
	}

	opts := []ollama.Option{
		ollama.WithModel(modelName),
		ollama.WithServerURL(ollamaURL),
	}

	llm, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}

	return embeddings.NewEmbedder(llm)
}
