package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/provider"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/scorer"
)

func evalCmd(ctx context.Context) *cobra.Command {
	var (
		experimentsDir string
		count          int
		experimentID   string
	)

	cmd := &cobra.Command{
		Use:   "eval",
		Short: "Run LLM-as-judge evaluation on completed experiments",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := provider.Build(ctx, flagProvider, flagModel, flagHost, flagFallback)
			if err != nil {
				return fmt.Errorf("build provider: %w", err)
			}

			entries, err := os.ReadDir(experimentsDir)
			if err != nil {
				return err
			}

			processed := 0
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				if experimentID != "" && e.Name() != experimentID {
					continue
				}
				expDir := filepath.Join(experimentsDir, e.Name())
				mf := filepath.Join(expDir, "metrics.json")
				data, err := os.ReadFile(mf)
				if err != nil {
					continue
				}
				var metrics map[string]any
				if json.Unmarshal(data, &metrics) != nil {
					continue
				}

				format, _ := metrics["format"].(string)
				ext := gap.FormatToExt[format]
				if ext == "" {
					ext = ".txt"
				}

				// Text quality metrics.
				baseOut := filepath.Join(expDir, "outputs", "base")
				gapOut := filepath.Join(expDir, "outputs", "gap")
				quality := scorer.ScoreExperiment(baseOut, gapOut, ext)

				// LLM-as-judge.
				comparisons, judgeErr := scorer.JudgeExperiment(ctx, p, expDir, ext)
				if judgeErr == nil && len(comparisons) > 0 {
					quality.JudgeComparisons = comparisons
					var sumBase, sumGAP float64
					for _, c := range comparisons {
						sumBase += c.BaseScore
						sumGAP += c.GAPScore
					}
					n := float64(len(comparisons))
					mb := sumBase / n
					mg := sumGAP / n
					quality.MeanBaseJudge = &mb
					quality.MeanGAPJudge = &mg
				}

				// Write eval.json.
				evalJSON, _ := json.MarshalIndent(quality, "", "  ")
				_ = os.WriteFile(filepath.Join(expDir, "eval.json"), append(evalJSON, '\n'), 0o644)

				fmt.Printf("%s: seq_sim=%.3f f1=%.3f", e.Name(), quality.MeanSequenceSimilarity, quality.MeanTokenF1)
				if quality.MeanBaseJudge != nil && quality.MeanGAPJudge != nil {
					fmt.Printf(" base_judge=%.3f gap_judge=%.3f", *quality.MeanBaseJudge, *quality.MeanGAPJudge)
				}
				fmt.Println()

				processed++
				if count > 0 && processed >= count {
					break
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&experimentsDir, "experiments-dir", "", "Experiments directory")
	cmd.Flags().IntVar(&count, "count", 0, "Max experiments (0=all)")
	cmd.Flags().StringVar(&experimentID, "id", "", "Evaluate single experiment")
	return cmd
}
