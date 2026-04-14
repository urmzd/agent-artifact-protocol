// Package cmd implements the cobra CLI commands for gap-eval.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var (
	flagProvider      string
	flagProviders     string
	flagModel         string
	flagHost          string
	flagFallback      string
	flagExperimentsDir string
	flagFormat        string
)

// Root creates the root cobra command.
func Root(ctx context.Context) *cobra.Command {
	root := &cobra.Command{
		Use:   "gap-eval",
		Short: "GAP benchmarks and evaluations",
	}

	root.PersistentFlags().StringVar(&flagProvider, "provider", "google", "LLM provider (google|openai|github|groq|ollama)")
	root.PersistentFlags().StringVar(&flagProviders, "providers", "", "Comma-separated providers for parallel execution")
	root.PersistentFlags().StringVar(&flagModel, "model", "", "Model name (applies to single --provider only)")
	root.PersistentFlags().StringVar(&flagHost, "host", "http://localhost:11434", "Ollama host")
	root.PersistentFlags().StringVar(&flagFallback, "fallback", "", "Fallback provider (comma-separated)")
	root.PersistentFlags().StringVar(&flagFormat, "format", "human", "Output format: human or json")

	root.AddCommand(
		versionCmd(),
		runCmd(ctx),
		scoreCmd(),
		reportCmd(),
		evalCmd(ctx),
	)

	return root
}
