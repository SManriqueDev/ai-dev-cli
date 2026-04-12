package ai

import (
	"context"
	"errors"
	"os"

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
	provider := os.Getenv("AI_PROVIDER")

	switch provider {
	case "openai":
		return newOpenAIClient()
	case "ollama":
		return newOllamaClient()
	default:
		return newOpenAIClient()
	}
}

func newOpenAIClient() (AIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY not set")
	}

	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithModel("gpt-4o"),
	)
	if err != nil {
		return nil, err
	}

	return &client{llm: llm}, nil
}

func newOllamaClient() (AIClient, error) {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}

	llm, err := ollama.New(
		ollama.WithModel(model),
	)
	if err != nil {
		return nil, err
	}

	return &client{llm: llm}, nil
}

func (c *client) Generate(ctx context.Context, prompt string) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, c.llm, prompt)
}
