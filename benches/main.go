// Conversation benchmark for the Agent-Artifact Protocol.
//
// Two subcommands:
//   - generate: create experiment input directories from the prompt catalog (no LLM)
//   - run: execute experiments against an LLM and collect metrics
//
// See data/experiments/EXPERIMENT.md for methodology.
//
// Usage:
//
//	go run . generate                              # generate all 88 experiments
//	go run . generate --count 10                   # generate first 10
//	go run . run --single 1                        # run experiment #1
//	go run . run --count 10 --model qwen3.5:9b     # run first 10 with specific model
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: bench <generate|run> [flags]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...) // shift for flag parsing

	switch cmd {
	case "generate":
		cmdGenerate()
	case "run":
		cmdRun()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func cmdGenerate() {
	var (
		promptsFile = flag.String("prompts", "data/prompts.json", "path to prompts JSON")
		outputDir   = flag.String("output", "data/experiments", "output directory")
		count       = flag.Int("count", 0, "number of experiments (0 = all prompts)")
	)
	flag.Parse()

	prompts, err := loadPrompts(*promptsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading prompts: %v\n", err)
		os.Exit(1)
	}

	n := len(prompts)
	if *count > 0 {
		n = *count
	}

	if err := generateAllInputs(prompts, *outputDir, n); err != nil {
		fmt.Fprintf(os.Stderr, "error generating: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %d experiment input directories in %s\n", n, *outputDir)
}

func cmdRun() {
	var (
		expDir   = flag.String("experiments", "data/experiments", "experiments directory")
		prompts  = flag.String("prompts", "data/prompts.json", "path to prompts JSON")
		provider = flag.String("provider", "ollama", "LLM provider")
		model    = flag.String("model", "", "model name")
		host     = flag.String("host", "http://localhost:11434", "Ollama host")
		single   = flag.Int("single", 0, "run single experiment by number")
		count    = flag.Int("count", 0, "number of experiments to run")
		verbose  = flag.Bool("verbose", false, "verbose output")
	)
	flag.Parse()

	_ = prompts  // prompts used for metadata lookup
	_ = expDir
	_ = provider
	_ = model
	_ = host
	_ = single
	_ = count
	_ = verbose

	// TODO: implement run command — reads inputs from experiment dirs,
	// sends to LLM, collects metrics, writes outputs + metrics.json
	fmt.Println("run command not yet implemented — use 'generate' first to create inputs")
}

func loadPrompts(path string) ([]Prompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var prompts []Prompt
	if err := json.Unmarshal(data, &prompts); err != nil {
		return nil, err
	}
	return prompts, nil
}

func writeJSON(path string, v any) {
	os.MkdirAll(filepath.Dir(path), 0o755)
	data, _ := json.MarshalIndent(v, "", "  ")
	os.WriteFile(path, append(data, '\n'), 0o644)
}

func writeFile(path, content string) {
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(content), 0o644)
}
