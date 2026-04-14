package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/experiment"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/provider"
)

func runCmd(ctx context.Context) *cobra.Command {
	var (
		experimentsDir string
		count          int
		experimentID   string
		flow           string
		skipEval       bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run conversation benchmark experiments (base vs GAP flows)",
		Long:  "Use --providers for parallel execution across multiple providers:\n  gap-eval run --providers google,groq,github",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve experiments dir relative to data dir.
			if experimentsDir == "" {
				// Default: sibling path to this binary's source tree.
				exe, _ := os.Executable()
				experimentsDir = filepath.Join(filepath.Dir(exe), "..", "..", "libs", "evals", "data", "experiments")
			}

			// Load GAP spec files.
			dataDir := filepath.Join(filepath.Dir(experimentsDir), "..")
			gapInitSpec := readFileOr(filepath.Join(dataDir, "gap-spec-init.md"), "")
			gapMaintainSpec := readFileOr(filepath.Join(dataDir, "gap-spec-maintain.md"), "")

			// Build provider list.
			var providerNames []string
			if flagProviders != "" {
				for _, p := range strings.Split(flagProviders, ",") {
					if p = strings.TrimSpace(p); p != "" {
						providerNames = append(providerNames, p)
					}
				}
			} else {
				providerNames = []string{flagProvider}
			}

			var slots []experiment.ProviderSlot
			for _, pName := range providerNames {
				m := flagModel
				if len(providerNames) > 1 {
					m = "" // use defaults for multi-provider
				}
				p, err := provider.Build(ctx, pName, m, flagHost, flagFallback)
				if err != nil {
					return fmt.Errorf("build provider %q: %w", pName, err)
				}
				modelName := m
				if modelName == "" {
					modelName = provider.Defaults[pName]
				}
				slots = append(slots, experiment.ProviderSlot{
					Provider: p,
					Name:     pName,
					Model:    modelName,
				})
			}

			// Collect experiment dirs.
			expDirs := experiment.LoadExperimentDirs(experimentsDir)
			if experimentID != "" {
				var filtered []string
				for _, d := range expDirs {
					if strings.HasPrefix(filepath.Base(d), experimentID) {
						filtered = append(filtered, d)
					}
				}
				expDirs = filtered
			}
			if count > 0 && count < len(expDirs) {
				expDirs = expDirs[:count]
			}
			if len(expDirs) == 0 {
				return fmt.Errorf("no experiments found")
			}

			labels := make([]string, len(providerNames))
			for i, p := range providerNames {
				labels[i] = fmt.Sprintf("%s (%s)", p, provider.Defaults[p])
			}
			fmt.Printf("Running %d experiment(s) across %d provider(s): %s\n\n", len(expDirs), len(slots), strings.Join(labels, ", "))

			succeeded, failed := experiment.RunMultiProvider(ctx, slots, expDirs, flow, skipEval, gapInitSpec, gapMaintainSpec, func(msg string) {
				fmt.Println(msg)
			})

			fmt.Printf("\nDone. %d succeeded, %d failed across %d providers.\n", succeeded, failed, len(slots))
			return nil
		},
	}

	cmd.Flags().StringVar(&experimentsDir, "experiments-dir", "", "Experiments directory (default: libs/evals/data/experiments)")
	cmd.Flags().IntVar(&count, "count", 0, "Max experiments (0=all)")
	cmd.Flags().StringVar(&experimentID, "id", "", "Run single experiment by ID prefix")
	cmd.Flags().StringVar(&flow, "flow", "both", "Which flow: base, gap, both")
	cmd.Flags().BoolVar(&skipEval, "skip-eval", false, "Skip quality eval")

	return cmd
}

func readFileOr(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	return strings.TrimSpace(string(data))
}
