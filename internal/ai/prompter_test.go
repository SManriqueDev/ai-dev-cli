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

func (m *MockAIClient) GenerateStream(ctx context.Context, prompt string, onChunk func(chunk string) error) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	// Simulate streaming by calling onChunk with the full response
	if err := onChunk(m.Response); err != nil {
		return "", err
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

func TestPrompter_ImproveCodeWithContext(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "improved code",
	}
	ragContext := "This is some relevant context about the project."

	prompter := NewPrompter(mockClient)
	result, err := prompter.ImproveCodeWithContext("func example() {}", ragContext)

	require.NoError(t, err)
	require.Equal(t, "improved code", result)
}

// ============ Streaming Tests ============

func TestPrompter_ImproveCodeStream(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "improved code response",
	}

	prompter := NewPrompter(mockClient)
	mockWriter := NewMockStreamWriter()

	ctx := context.Background()
	err := prompter.ImproveCodeStream(ctx, "func example() {}", mockWriter)

	require.NoError(t, err)
	require.Greater(t, len(mockWriter.Chunks), 0)

	// Verify content was written
	combined := combineChunks(mockWriter.Chunks)
	require.Equal(t, "improved code response", combined)
}

func TestPrompter_ImproveCodeStream_Error(t *testing.T) {
	mockClient := &MockAIClient{
		Err: context.DeadlineExceeded,
	}

	prompter := NewPrompter(mockClient)
	mockWriter := NewMockStreamWriter()

	ctx := context.Background()
	err := prompter.ImproveCodeStream(ctx, "func example() {}", mockWriter)

	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPrompter_ImproveCodeStream_NilWriter(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "improved code",
	}

	prompter := NewPrompter(mockClient)
	ctx := context.Background()
	err := prompter.ImproveCodeStream(ctx, "func example() {}", nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "stream writer cannot be nil")
}

func TestPrompter_ImproveCodeStreamWithContext(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "improved code with context",
	}

	prompter := NewPrompter(mockClient)
	mockWriter := NewMockStreamWriter()
	ragContext := "project context"

	ctx := context.Background()
	err := prompter.ImproveCodeStreamWithContext(ctx, "func example() {}", ragContext, mockWriter)

	require.NoError(t, err)
	require.Greater(t, len(mockWriter.Chunks), 0)

	combined := combineChunks(mockWriter.Chunks)
	require.Equal(t, "improved code with context", combined)
}

func TestPrompter_GenerateTestsStream(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "func TestExample(t *testing.T) {}",
	}

	prompter := NewPrompter(mockClient)
	mockWriter := NewMockStreamWriter()

	ctx := context.Background()
	err := prompter.GenerateTestsStream(ctx, "func example() {}", mockWriter)

	require.NoError(t, err)
	require.Greater(t, len(mockWriter.Chunks), 0)

	combined := combineChunks(mockWriter.Chunks)
	require.Equal(t, "func TestExample(t *testing.T) {}", combined)
}

func TestPrompter_GenerateTestsStream_NilWriter(t *testing.T) {
	mockClient := &MockAIClient{
		Response: "test code",
	}

	prompter := NewPrompter(mockClient)
	ctx := context.Background()
	err := prompter.GenerateTestsStream(ctx, "func example() {}", nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "stream writer cannot be nil")
}

// ============ Mock Stream Writer for Testing ============

type MockStreamWriter struct {
	Chunks [][]byte
	Closed bool
}

func NewMockStreamWriter() *MockStreamWriter {
	return &MockStreamWriter{
		Chunks: make([][]byte, 0),
	}
}

func (m *MockStreamWriter) WriteChunk(_ context.Context, chunk []byte) error {
	if m.Closed {
		return context.Canceled
	}
	// Make a copy of the chunk
	chunkCopy := make([]byte, len(chunk))
	copy(chunkCopy, chunk)
	m.Chunks = append(m.Chunks, chunkCopy)
	return nil
}

func (m *MockStreamWriter) Close() error {
	m.Closed = true
	return nil
}

func combineChunks(chunks [][]byte) string {
	var result string
	for _, chunk := range chunks {
		result += string(chunk)
	}
	return result
}
