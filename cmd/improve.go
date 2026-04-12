package cmd

import (
	"fmt"
	"os"

	"github.com/ai-dev-cli/ai-dev-cli/internal/ai"
	"github.com/spf13/cobra"
)

var improveCmd = &cobra.Command{
	Use:   "improve <file>",
	Short: "Improve code quality using AI",
	Long:  `Analyzes the given Go file and suggests improvements for code quality, performance, and best practices.`,
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
		result, err := prompter.ImproveCode(string(content))
		if err != nil {
			return fmt.Errorf("failed to improve code: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}
