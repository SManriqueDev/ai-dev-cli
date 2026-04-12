package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	DefaultProvider            = "openai"
	DefaultOpenAIBaseURL       = "https://api.openai.com/v1"
	DefaultOpenAIModel         = "gpt-4o"
	DefaultOpenAIEmbedderModel = "text-embedding-3-small"
	DefaultOllamaBaseURL       = "http://localhost:11434"
	DefaultOllamaModel         = "llama3.2"
	DefaultOllamaEmbedderModel = "nomic-embed-text"
	DefaultChromaURL           = "http://localhost:8000"
	DefaultCollectionName      = "ai-dev-cli-db"
	DefaultChunkSize           = 1000
	DefaultChunkOverlap        = 200
)

var cfg *viper.Viper

type AppConfig struct {
	Provider string
	OpenAI   ProviderConfig
	Ollama   ProviderConfig
	RAG      RAGConfig
	Chroma   ChromaConfig
}

type ProviderConfig struct {
	BaseURL       string
	Model         string
	APIKey        string
	EmbedderModel string
}

type ChromaConfig struct {
	URL            string
	CollectionName string
}

type RAGConfig struct {
	ChromaURL      string
	CollectionName string
	Provider       string
	EmbedderModel  string
	APIKey         string
	BaseURL        string
	OllamaURL      string
	ChunkSize      int
	ChunkOverlap   int
	NumResults     int
}

type AIConfig struct {
	Provider string
	BaseURL  string
	Model    string
	APIKey   string
}

func Load() error {
	_ = godotenv.Load()

	cfg = viper.New()
	cfg.AutomaticEnv()

	if err := bindEnv("provider", "AI_PROVIDER"); err != nil {
		return err
	}
	if err := bindEnv("openai.base_url", "OPENAI_BASE_URL"); err != nil {
		return err
	}
	if err := bindEnv("openai.model", "OPENAI_MODEL"); err != nil {
		return err
	}
	if err := bindEnv("openai.api_key", "OPENAI_API_KEY"); err != nil {
		return err
	}
	if err := bindEnv("openai.embedder_model", "OPENAI_EMBEDDER_MODEL"); err != nil {
		return err
	}
	if err := bindEnv("ollama.base_url", "OLLAMA_BASE_URL"); err != nil {
		return err
	}
	if err := bindEnv("ollama.model", "OLLAMA_MODEL"); err != nil {
		return err
	}
	if err := bindEnv("ollama.embedder_model", "OLLAMA_EMBEDDER_MODEL"); err != nil {
		return err
	}
	if err := bindEnv("chroma.url", "CHROMA_URL"); err != nil {
		return err
	}
	if err := bindEnv("chroma.collection_name", "CHROMA_COLLECTION_NAME"); err != nil {
		return err
	}
	if err := bindEnv("rag.chunk_size", "RAG_CHUNK_SIZE"); err != nil {
		return err
	}
	if err := bindEnv("rag.chunk_overlap", "RAG_CHUNK_OVERLAP"); err != nil {
		return err
	}

	setDefaults(cfg)

	if provider := GetProvider(); provider != DefaultProvider && provider != "ollama" {
		return fmt.Errorf("unsupported provider %q", provider)
	}

	return nil
}

func Current() AppConfig {
	provider := GetProvider()

	return AppConfig{
		Provider: provider,
		OpenAI: ProviderConfig{
			BaseURL:       getString("openai.base_url", DefaultOpenAIBaseURL),
			Model:         getString("openai.model", DefaultOpenAIModel),
			APIKey:        getString("openai.api_key", ""),
			EmbedderModel: getString("openai.embedder_model", DefaultOpenAIEmbedderModel),
		},
		Ollama: ProviderConfig{
			BaseURL:       getString("ollama.base_url", DefaultOllamaBaseURL),
			Model:         getString("ollama.model", DefaultOllamaModel),
			EmbedderModel: getString("ollama.embedder_model", DefaultOllamaEmbedderModel),
		},
		Chroma: ChromaConfig{
			URL:            getString("chroma.url", DefaultChromaURL),
			CollectionName: getString("chroma.collection_name", DefaultCollectionName),
		},
		RAG: RAGConfig{
			ChromaURL:      getString("chroma.url", DefaultChromaURL),
			CollectionName: getString("chroma.collection_name", DefaultCollectionName),
			Provider:       provider,
			EmbedderModel:  getActiveEmbedderModel(provider),
			APIKey:         getActiveAPIKey(provider),
			BaseURL:        getActiveBaseURL(provider),
			OllamaURL:      getActiveOllamaURL(provider),
			ChunkSize:      getInt("rag.chunk_size", DefaultChunkSize),
			ChunkOverlap:   getInt("rag.chunk_overlap", DefaultChunkOverlap),
			NumResults:     0,
		},
	}
}

func Get(key string) any {
	if cfg == nil {
		return nil
	}

	return cfg.Get(key)
}

func GetString(key string) string {
	return getString(key, "")
}

func GetProvider() string {
	return normalizeProvider(getString("provider", DefaultProvider))
}

func GetAIConfig() AIConfig {
	current := Current()
	switch current.Provider {
	case "ollama":
		return AIConfig{
			Provider: current.Provider,
			BaseURL:  current.Ollama.BaseURL,
			Model:    current.Ollama.Model,
		}
	default:
		return AIConfig{
			Provider: current.Provider,
			BaseURL:  current.OpenAI.BaseURL,
			Model:    current.OpenAI.Model,
			APIKey:   current.OpenAI.APIKey,
		}
	}
}

func GetRAGConfig() RAGConfig {
	return Current().RAG
}

func GetChromaConfig() ChromaConfig {
	return Current().Chroma
}

func getActiveEmbedderModel(provider string) string {
	switch provider {
	case "ollama":
		return getString("ollama.embedder_model", DefaultOllamaEmbedderModel)
	default:
		return getString("openai.embedder_model", DefaultOpenAIEmbedderModel)
	}
}

func getActiveAPIKey(provider string) string {
	if provider == "openai" {
		return getString("openai.api_key", "")
	}

	return ""
}

func getActiveBaseURL(provider string) string {
	switch provider {
	case "ollama":
		return getString("ollama.base_url", DefaultOllamaBaseURL)
	default:
		return getString("openai.base_url", DefaultOpenAIBaseURL)
	}
}

func getActiveOllamaURL(provider string) string {
	if provider == "ollama" {
		return getString("ollama.base_url", DefaultOllamaBaseURL)
	}

	return ""
}

func getString(key, fallback string) string {
	if cfg == nil {
		return fallback
	}

	value := strings.TrimSpace(cfg.GetString(key))
	if value == "" {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) int {
	if cfg == nil {
		return fallback
	}

	if !cfg.IsSet(key) {
		return fallback
	}

	value := cfg.GetInt(key)
	if value == 0 {
		return fallback
	}

	return value
}

func bindEnv(key, env string) error {
	if err := cfg.BindEnv(key, env); err != nil {
		return err
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("provider", DefaultProvider)
	v.SetDefault("openai.base_url", DefaultOpenAIBaseURL)
	v.SetDefault("openai.model", DefaultOpenAIModel)
	v.SetDefault("openai.embedder_model", DefaultOpenAIEmbedderModel)
	v.SetDefault("ollama.base_url", DefaultOllamaBaseURL)
	v.SetDefault("ollama.model", DefaultOllamaModel)
	v.SetDefault("ollama.embedder_model", DefaultOllamaEmbedderModel)
	v.SetDefault("chroma.url", DefaultChromaURL)
	v.SetDefault("chroma.collection_name", DefaultCollectionName)
	v.SetDefault("rag.chunk_size", DefaultChunkSize)
	v.SetDefault("rag.chunk_overlap", DefaultChunkOverlap)
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}
