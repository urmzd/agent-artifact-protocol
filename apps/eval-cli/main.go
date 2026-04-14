package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Root(ctx).Execute(); err != nil {
		os.Exit(1)
	}
}
