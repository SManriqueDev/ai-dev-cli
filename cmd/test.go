package cmd

import (
	"fmt"
	"os"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <file>",
	Short: "Generate unit tests for the given file",
	Long:  `Analyzes the given Go file and generates comprehensive unit tests using AI.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client, err := ai.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create AI client: %w", err)
		}

		prompter := ai.NewPrompter(client)
		result, err := prompter.GenerateTests(string(content))
		if err != nil {
			return fmt.Errorf("failed to generate tests: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}
