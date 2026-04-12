package ai

import (
	"context"
)

type CodeChain struct {
	prompter *Prompter
}

func NewCodeChain(prompter *Prompter) *CodeChain {
	return &CodeChain{prompter: prompter}
}

func (c *CodeChain) ImproveCode(ctx context.Context, code string) (string, error) {
	return c.prompter.ImproveCode(code)
}

func (c *CodeChain) GenerateTests(ctx context.Context, code string) (string, error) {
	return c.prompter.GenerateTests(code)
}
