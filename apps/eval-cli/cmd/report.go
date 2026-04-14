package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func reportCmd() *cobra.Command {
	var (
		experimentsDir string
		output         string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate report from experiment metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := filepath.Join(experimentsDir, "*/metrics.json")
			files, _ := filepath.Glob(pattern)
			if len(files) == 0 {
				return fmt.Errorf("no metrics found")
			}

			var results []map[string]any
			for _, f := range files {
				data, err := os.ReadFile(f)
				if err != nil {
					continue
				}
				var m map[string]any
				if json.Unmarshal(data, &m) == nil {
					results = append(results, m)
				}
			}

			if flagFormat == "json" {
				out, _ := json.MarshalIndent(results, "", "  ")
				if output != "" {
					return os.WriteFile(output, append(out, '\n'), 0o644)
				}
				fmt.Println(string(out))
				return nil
			}

			// Human-readable summary table.
			fmt.Printf("%-40s %-12s %-8s %-8s %-10s %-8s %-8s\n",
				"Experiment", "Provider", "Out Sav", "In Sav", "Lat Sav", "SeqSim", "F1")
			fmt.Println(dash(94))

			for _, r := range results {
				comp := asMap2(r["comparison"])
				qual := asMap2(r["quality"])
				fmt.Printf("%-40s %-12s %7.1f%% %7.1f%% %9.1f%% %7.3f %7.3f\n",
					r["experiment_id"],
					r["provider"],
					getf(comp, "output_token_savings_pct"),
					getf(comp, "input_token_savings_pct"),
					getf(comp, "latency_savings_pct"),
					getf(qual, "mean_sequence_similarity"),
					getf(qual, "mean_token_f1"),
				)
			}

			if output != "" {
				// Also write markdown.
				out, _ := json.MarshalIndent(results, "", "  ")
				return os.WriteFile(output, append(out, '\n'), 0o644)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&experimentsDir, "experiments-dir", "", "Experiments directory")
	cmd.Flags().StringVar(&output, "output", "", "Output file")
	return cmd
}

func asMap2(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func getf(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func dash(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '-'
	}
	return string(b)
}
