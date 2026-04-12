package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var cfg *viper.Viper

func Load() error {
	_ = godotenv.Load()

	cfg = viper.New()

	cfg.SetEnvPrefix("AI")
	cfg.AutomaticEnv()

	cfg.SetDefault("provider", "openai")
	cfg.SetDefault("openai.model", "gpt-4o")
	cfg.SetDefault("ollama.model", "llama3.2")
	cfg.SetDefault("ollama.url", "http://localhost:11434")

	_ = cfg.ReadInConfig()

	fmt.Println("Configuration loaded:")
	fmt.Printf("  Provider: %s\n", cfg.GetString("provider"))
	fmt.Printf("  OpenAI Model: %s\n", cfg.GetString("openai.model"))
	fmt.Printf("  Ollama URL: %s\n", cfg.GetString("ollama.url"))
	fmt.Printf("  Ollama Model: %s\n", cfg.GetString("ollama.model"))

	return nil
}

func Get(key string) any {
	return cfg.Get(key)
}

func GetString(key string) string {
	return cfg.GetString(key)
}

func GetProvider() string {
	return GetString("provider")
}

func GetOpenAIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

func GetOllamaURL() string {
	return GetString("ollama.url")
}
