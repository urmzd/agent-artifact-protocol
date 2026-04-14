// Package provider builds saige LLM providers from CLI flags.
package provider

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urmzd/saige/agent/provider/fallback"
	"github.com/urmzd/saige/agent/provider/google"
	"github.com/urmzd/saige/agent/provider/openai"
	"github.com/urmzd/saige/agent/types"
)

// Defaults maps provider names to default model IDs.
var Defaults = map[string]string{
	"google": "gemini-2.5-flash",
	"openai": "gpt-4o-mini",
	"ollama": "gemma4",
	"github": "openai/gpt-4o-mini",
	"groq":   "qwen3-32b",
}

// Build creates a types.Provider from a provider name and optional model override.
func Build(ctx context.Context, name, model, host, fallbackProviders string) (types.Provider, error) {
	primary, err := buildSingle(ctx, name, model, host)
	if err != nil {
		return nil, err
	}
	if fallbackProviders == "" {
		return primary, nil
	}
	var fallbacks []types.Provider
	for _, fb := range strings.Split(fallbackProviders, ",") {
		fb = strings.TrimSpace(fb)
		if fb == "" {
			continue
		}
		p, err := buildSingle(ctx, fb, "", host)
		if err != nil {
			return nil, fmt.Errorf("fallback provider %q: %w", fb, err)
		}
		fallbacks = append(fallbacks, p)
	}
	all := append([]types.Provider{primary}, fallbacks...)
	return fallback.New(all...), nil
}

func buildSingle(ctx context.Context, name, model, host string) (types.Provider, error) {
	if model == "" {
		model = Defaults[name]
	}
	switch name {
	case "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		return google.NewAdapter(ctx, apiKey, model)
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		return openai.NewAdapter(apiKey, model), nil
	case "github":
		apiKey := os.Getenv("GITHUB_TOKEN")
		if apiKey == "" {
			out, err := exec.Command("gh", "auth", "token").Output()
			if err == nil {
				apiKey = strings.TrimSpace(string(out))
			}
		}
		return openai.NewAdapter(apiKey, model, openai.WithBaseURL("https://models.inference.ai.azure.com")), nil
	case "groq":
		apiKey := os.Getenv("GROQ_API_KEY")
		return openai.NewAdapter(apiKey, model, openai.WithBaseURL("https://api.groq.com/openai/v1")), nil
	case "ollama":
		apiKey := os.Getenv("OLLAMA_API_KEY")
		base := strings.TrimRight(host, "/")
		if !strings.HasSuffix(base, "/v1") {
			base += "/v1"
		}
		return openai.NewAdapter(apiKey, model, openai.WithBaseURL(base)), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
