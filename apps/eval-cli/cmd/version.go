package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set via ldflags at build time.
var Version = "dev"

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the gap-eval version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
