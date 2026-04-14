package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/scorer"
)

func scoreCmd() *cobra.Command {
	var experimentsDir string

	cmd := &cobra.Command{
		Use:   "score",
		Short: "Score content quality for completed experiments (retroactive)",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(experimentsDir)
			if err != nil {
				return err
			}

			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				mf := filepath.Join(experimentsDir, e.Name(), "metrics.json")
				data, err := os.ReadFile(mf)
				if err != nil {
					continue
				}
				var metrics map[string]any
				if err := json.Unmarshal(data, &metrics); err != nil {
					continue
				}

				format, _ := metrics["format"].(string)
				if format == "" {
					format = "text/html"
				}
				ext := gap.FormatToExt[format]
				if ext == "" {
					ext = ".txt"
				}

				baseOut := filepath.Join(experimentsDir, e.Name(), "outputs", "base")
				gapOut := filepath.Join(experimentsDir, e.Name(), "outputs", "gap")

				if _, err := os.Stat(baseOut); err != nil {
					continue
				}
				if _, err := os.Stat(gapOut); err != nil {
					continue
				}

				quality := scorer.ScoreExperiment(baseOut, gapOut, ext)
				if len(quality.PerTurn) > 0 {
					metrics["quality"] = quality
					out, _ := json.MarshalIndent(metrics, "", "  ")
					_ = os.WriteFile(mf, append(out, '\n'), 0o644)
					fmt.Printf("%s: seq_sim=%.3f f1=%.3f\n",
						metrics["experiment_id"], quality.MeanSequenceSimilarity, quality.MeanTokenF1)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&experimentsDir, "experiments-dir", "", "Experiments directory")
	return cmd
}
