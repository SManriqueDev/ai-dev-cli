package ai

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

var ErrNoProvider = errors.New("no AI provider configured")

type AIClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, onChunk func(chunk string) error) (string, error)
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

// GenerateStream generates a response using streaming callbacks.
// onChunk is called for each chunk of bytes received from the LLM.
func (c *client) GenerateStream(ctx context.Context, prompt string, onChunk func(chunk string) error) (string, error) {
	var fullResponse string
	var mu sync.Mutex

	// Create a callback that captures chunks
	streamCallback := func(ctx context.Context, chunk []byte) error {
		if len(chunk) == 0 {
			return nil
		}

		text := string(chunk)

		// Call user's chunk handler
		if err := onChunk(text); err != nil {
			return err
		}

		// Also accumulate the full response
		mu.Lock()
		fullResponse += text
		mu.Unlock()

		return nil
	}

	// Use LangChain's streaming capability
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	// Try to stream the response
	_, err := c.llm.GenerateContent(ctx, messages, llms.WithStreamingFunc(streamCallback))
	if err != nil {
		return "", fmt.Errorf("streaming generation failed: %w", err)
	}

	return fullResponse, nil
}
