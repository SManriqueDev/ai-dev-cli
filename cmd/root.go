package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ai-dev-cli",
	Short: "AI-powered CLI for code improvement and test generation",
	Long: `AI-Dev-CLI helps developers validate prompts, generate unit tests,
and improve code quality using AI with RAG techniques.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println("AI-Dev-CLI v1.0.0")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
	rootCmd.AddCommand(improveCmd)
	rootCmd.AddCommand(testCmd)
}
