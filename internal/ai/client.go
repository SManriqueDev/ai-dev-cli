package ai

import (
	"context"
	"errors"

	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

var ErrNoProvider = errors.New("no AI provider configured")

type AIClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type client struct {
	llm llms.Model
}

func NewClient() (AIClient, error) {
	cfg := config.GetAIConfig()

	switch cfg.Provider {
	case "openai":
		return newOpenAIClient(cfg)
	case "ollama":
		return newOllamaClient(cfg)
	default:
		return nil, errors.New("unsupported AI provider")
	}
}

func newOpenAIClient(cfg config.AIConfig) (AIClient, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("OPENAI_API_KEY not set")
	}

	llm, err := openai.New(
		openai.WithToken(cfg.APIKey),
		openai.WithModel(cfg.Model),
		openai.WithBaseURL(cfg.BaseURL),
	)
	if err != nil {
		return nil, err
	}

	return &client{llm: llm}, nil
}

func newOllamaClient(cfg config.AIConfig) (AIClient, error) {
	if cfg.Model == "" {
		cfg.Model = config.DefaultOllamaModel
	}

	llm, err := ollama.New(
		ollama.WithModel(cfg.Model),
		ollama.WithServerURL(cfg.BaseURL),
	)
	if err != nil {
		return nil, err
	}

	return &client{llm: llm}, nil
}

func (c *client) Generate(ctx context.Context, prompt string) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, c.llm, prompt)
}
