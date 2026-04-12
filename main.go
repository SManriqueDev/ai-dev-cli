package main

import (
	"log"

	"github.com/ai-dev-cli/ai-dev-cli/cmd"
	"github.com/ai-dev-cli/ai-dev-cli/platform/config"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
