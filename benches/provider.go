package main

import (
	"fmt"

	"github.com/urmzd/saige/agent/provider/ollama"
	"github.com/urmzd/saige/agent/types"
)

func createProvider(provider, model, host string) (types.Provider, error) {
	switch provider {
	case "ollama":
		if model == "" {
			model = "qwen3.5:4b"
		}
		client := ollama.NewClient(host, model, "")
		return ollama.NewAdapter(client), nil
	default:
		return nil, fmt.Errorf("provider %q not yet supported (add openai/anthropic/google as needed)", provider)
	}
}
