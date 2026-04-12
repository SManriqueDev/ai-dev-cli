package rag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestDocumentProcessor_SplitText(t *testing.T) {
	processor := NewDocumentProcessor(100, 20)

	text := "This is a long text that needs to be split into smaller chunks for vector storage. " +
		"It should respect the chunk size and overlap parameters. " +
		"Each chunk should be meaningful and preserve context."

	chunks, err := processor.SplitText(text)

	require.NoError(t, err)
	assert.NotEmpty(t, chunks)
	assert.Greater(t, len(chunks), 1)

	for _, chunk := range chunks {
		assert.LessOrEqual(t, len(chunk), 120)
	}
}

func TestDocumentProcessor_ProcessText(t *testing.T) {
	processor := NewDocumentProcessor(50, 10)

	text := "This is sample text for testing. It should be split into chunks. " +
		"Each chunk will have metadata attached."

	docs, err := processor.ProcessText(text, map[string]any{
		"source": "test",
		"type":   "test document",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, docs)

	for _, doc := range docs {
		assert.NotEmpty(t, doc.PageContent)
		assert.Equal(t, "test", doc.Metadata["source"])
		assert.Equal(t, "test document", doc.Metadata["type"])
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"ftp://example.com", false},
		{"example.com", false},
		{"/path/to/file", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidURL(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestFormatContext(t *testing.T) {
	docs := []schema.Document{
		{
			PageContent: "This is the first chunk of content.",
			Metadata:    map[string]any{"source": "test1.go"},
		},
		{
			PageContent: "This is the second chunk of content.",
			Metadata:    map[string]any{"source": "test2.go"},
		},
		{
			PageContent: "This is the third chunk of content.",
			Metadata:    map[string]any{"source": "test3.go"},
		},
	}

	result := FormatContext(docs)

	assert.Contains(t, result, "Contexto Recuperado")
	assert.Contains(t, result, "test1.go")
	assert.Contains(t, result, "test2.go")
	assert.Contains(t, result, "test3.go")
}
