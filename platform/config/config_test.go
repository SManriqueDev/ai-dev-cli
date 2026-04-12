package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultsToOpenAI(t *testing.T) {
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("OPENAI_EMBEDDER_MODEL", "")
	t.Setenv("OLLAMA_BASE_URL", "")
	t.Setenv("OLLAMA_MODEL", "")
	t.Setenv("OLLAMA_EMBEDDER_MODEL", "")
	t.Setenv("CHROMA_URL", "")
	t.Setenv("CHROMA_COLLECTION_NAME", "")
	t.Setenv("RAG_CHUNK_SIZE", "")
	t.Setenv("RAG_CHUNK_OVERLAP", "")

	cfg = nil
	t.Cleanup(func() {
		cfg = nil
	})

	require.NoError(t, Load())

	require.Equal(t, DefaultProvider, GetProvider())
	require.Equal(t, DefaultOpenAIModel, GetAIConfig().Model)
	require.Equal(t, DefaultOpenAIBaseURL, GetAIConfig().BaseURL)
	require.Equal(t, DefaultOpenAIEmbedderModel, GetRAGConfig().EmbedderModel)
	require.Equal(t, DefaultOpenAIBaseURL, GetRAGConfig().BaseURL)
	require.Equal(t, DefaultChromaURL, GetRAGConfig().ChromaURL)
	require.Equal(t, DefaultCollectionName, GetRAGConfig().CollectionName)
	require.Equal(t, DefaultChunkSize, GetRAGConfig().ChunkSize)
	require.Equal(t, DefaultChunkOverlap, GetRAGConfig().ChunkOverlap)
}

func TestLoad_UsesOllamaConfiguration(t *testing.T) {
	t.Setenv("AI_PROVIDER", "ollama")
	t.Setenv("OLLAMA_BASE_URL", "http://ollama:11434")
	t.Setenv("OLLAMA_MODEL", "llama3.1")
	t.Setenv("OLLAMA_EMBEDDER_MODEL", "nomic-embed-custom")
	t.Setenv("CHROMA_URL", "http://chroma:8000")
	t.Setenv("CHROMA_COLLECTION_NAME", "docs")
	t.Setenv("RAG_CHUNK_SIZE", "750")
	t.Setenv("RAG_CHUNK_OVERLAP", "100")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_MODEL", "")
	t.Setenv("OPENAI_EMBEDDER_MODEL", "")

	cfg = nil
	t.Cleanup(func() {
		cfg = nil
	})

	require.NoError(t, Load())

	aiCfg := GetAIConfig()
	ragCfg := GetRAGConfig()

	require.Equal(t, "ollama", aiCfg.Provider)
	require.Equal(t, "llama3.1", aiCfg.Model)
	require.Equal(t, "http://ollama:11434", aiCfg.BaseURL)
	require.Equal(t, "nomic-embed-custom", ragCfg.EmbedderModel)
	require.Equal(t, "http://ollama:11434", ragCfg.BaseURL)
	require.Equal(t, "http://ollama:11434", ragCfg.OllamaURL)
	require.Equal(t, "http://chroma:8000", ragCfg.ChromaURL)
	require.Equal(t, "docs", ragCfg.CollectionName)
	require.Equal(t, 750, ragCfg.ChunkSize)
	require.Equal(t, 100, ragCfg.ChunkOverlap)
}

func TestLoad_RejectsUnsupportedProvider(t *testing.T) {
	t.Setenv("AI_PROVIDER", "unsupported")

	cfg = nil
	t.Cleanup(func() {
		cfg = nil
	})

	require.Error(t, Load())
}
