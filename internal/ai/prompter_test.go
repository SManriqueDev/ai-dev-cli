package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockAIClient struct {
	Response string
	Err      error
}

func (m *MockAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.Response, nil
}

func TestNewPrompter(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "test response",
	}

	prompter := NewPrompter(mockClient)
	require.NotNil(t, prompter)
	require.Same(t, mockClient, prompter.client)
}

func TestPrompter_ImproveCode(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "improved code",
	}

	prompter := NewPrompter(mockClient)
	result, err := prompter.ImproveCode("func example() {}")

	require.NoError(t, err)
	require.Equal(t, "improved code", result)
}

func TestPrompter_GenerateTests(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "func TestExample(t *testing.T) {}",
	}

	prompter := NewPrompter(mockClient)
	result, err := prompter.GenerateTests("func example() {}")

	require.NoError(t, err)
	require.Contains(t, result, "TestExample")
}

func TestPrompter_ImproveCode_Error(t *testing.T) {
	mockClient := &MockAIClient{
		Err: context.DeadlineExceeded,
	}

	prompter := NewPrompter(mockClient)
	_, err := prompter.ImproveCode("func example() {}")

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
