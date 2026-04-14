package experiment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urmzd/saige/agent/types"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/runner"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/scorer"
)

// RunSingleExperiment runs base vs GAP flows for one experiment.
func RunSingleExperiment(ctx context.Context, provider types.Provider, providerName, modelName string, expDir string, flow string, skipEval bool, gapInitSpec, gapMaintainSpec string, log func(string)) error {
	expName := filepath.Base(expDir)
	format, ext := gap.ParseExperimentFormat(filepath.Join(expDir, "README.md"))

	if _, err := os.Stat(filepath.Join(expDir, "metrics.json")); err == nil {
		log(fmt.Sprintf("%s — already done, skipping", expName))
		return nil
	}

	log(fmt.Sprintf("%s (%s) via %s", expName, format, providerName))

	baseInput := filepath.Join(expDir, "inputs", "base")
	baseOutput := filepath.Join(expDir, "outputs", "base")
	gapOutput := filepath.Join(expDir, "outputs", "gap")
	_ = os.MkdirAll(baseOutput, 0o755)
	_ = os.MkdirAll(gapOutput, 0o755)

	systemData, err := os.ReadFile(filepath.Join(baseInput, "system.md"))
	if err != nil {
		return fmt.Errorf("read system.md: %w", err)
	}
	baseSystem := strings.TrimSpace(string(systemData))
	initSystem := baseSystem + "\n\n" + gapInitSpec
	maintainSystem := baseSystem + "\n\n" + gapMaintainSpec

	turn0Prompt, editPrompts, err := LoadTurnPrompts(baseInput)
	if err != nil || turn0Prompt == "" {
		log(fmt.Sprintf("  %s: no turn files, skipping", expName))
		return nil
	}

	metrics := map[string]any{
		"experiment_id": expName,
		"model":         modelName,
		"provider":      providerName,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"format":        format,
	}

	// Turn 0
	if flow == "base" || flow == "both" {
		artifact, ag, bt0, err := runner.RunBaseTurn0(ctx, provider, baseSystem, turn0Prompt, baseOutput, ext)
		if err != nil {
			return fmt.Errorf("base turn-0: %w", err)
		}
		metrics["base_turn0"] = bt0
		log(fmt.Sprintf("  [%s] %s base turn-0: %d tokens out, %d bytes, %dms", providerName, expName, bt0.OutputTokens, bt0.ArtifactBytes, bt0.LatencyMs))

		if len(editPrompts) > 0 {
			_ = artifact // used implicitly via agent tree
			baseResults, _, err := runner.RunBaseFlow(ctx, ag, editPrompts, baseOutput, ext)
			if err != nil {
				return fmt.Errorf("base flow: %w", err)
			}
			totalOut, totalIn, totalMs := sumTurnResults(baseResults)
			metrics["default_flow"] = map[string]any{
				"per_turn":            baseResults,
				"total_input_tokens":  totalIn,
				"total_output_tokens": totalOut,
				"total_latency_ms":    totalMs,
			}
			for _, r := range baseResults {
				status := "ok"
				if r.Failed {
					status = fmt.Sprintf("FAIL: %s", r.FailureReason)
				}
				log(fmt.Sprintf("  [%s] %s base turn-%d: %d tokens out, %d in, %d bytes, %dms [%s]", providerName, expName, r.Turn, r.OutputTokens, r.InputTokens, r.OutputBytes, r.LatencyMs, status))
			}
		}
	}

	if flow == "gap" || flow == "both" {
		gapArt, at0, err := runner.RunGAPTurn0(ctx, provider, initSystem, turn0Prompt, gapOutput, ext)
		if err != nil {
			return fmt.Errorf("gap turn-0: %w", err)
		}
		metrics["gap_turn0"] = at0
		log(fmt.Sprintf("  [%s] %s gap  turn-0: %d tokens out, %d bytes, %dms", providerName, expName, at0.OutputTokens, at0.ArtifactBytes, at0.LatencyMs))

		if len(editPrompts) > 0 {
			gapResults, _, err := runner.RunGAPFlow(ctx, provider, maintainSystem, gapArt, editPrompts, format, gapOutput, ext)
			if err != nil {
				return fmt.Errorf("gap flow: %w", err)
			}
			totalOut, totalIn, totalMs := sumGAPTurnResults(gapResults)
			parseOK := 0
			applyOK := 0
			for _, r := range gapResults {
				if r.EnvelopeParsed {
					parseOK++
				}
				if r.ApplySucceeded {
					applyOK++
				}
			}
			numEdits := len(gapResults)
			parseRate := 0.0
			applyRate := 0.0
			if numEdits > 0 {
				parseRate = float64(parseOK) / float64(numEdits)
				applyRate = float64(applyOK) / float64(numEdits)
			}
			metrics["gap_flow"] = map[string]any{
				"per_turn":            gapResults,
				"total_input_tokens":  totalIn,
				"total_output_tokens": totalOut,
				"total_latency_ms":    totalMs,
				"envelope_parse_rate": parseRate,
				"apply_success_rate":  applyRate,
			}
			for _, r := range gapResults {
				status := "ok"
				if r.Failed {
					status = fmt.Sprintf("FAIL: %s", r.FailureReason)
				} else if !r.ApplySucceeded {
					status = "apply-fail"
				}
				log(fmt.Sprintf("  [%s] %s gap  turn-%d: %d tokens out, %d in, %d bytes, %dms, %s [%s]", providerName, expName, r.Turn, r.OutputTokens, r.InputTokens, r.OutputBytes, r.LatencyMs, r.EnvelopeName, status))
			}
		}
	}

	// Comparison
	if df, ok := metrics["default_flow"]; ok {
		if gf, ok2 := metrics["gap_flow"]; ok2 {
			dfm := df.(map[string]any)
			gfm := gf.(map[string]any)
			bo := geti(dfm, "total_output_tokens")
			ao := geti(gfm, "total_output_tokens")
			bi := geti(dfm, "total_input_tokens")
			ai := geti(gfm, "total_input_tokens")
			bms := geti(dfm, "total_latency_ms")
			ams := geti(gfm, "total_latency_ms")

			metrics["comparison"] = map[string]any{
				"output_token_savings_pct": pctSaving(bo, ao),
				"input_token_savings_pct":  pctSaving(bi, ai),
				"latency_savings_pct":      pctSaving(bms, ams),
			}
			metrics["token_table"] = BuildTokenTable(metrics)
		}
	}

	// Quality eval
	if !skipEval {
		if _, err := os.Stat(baseOutput); err == nil {
			if _, err := os.Stat(gapOutput); err == nil {
				quality := scorer.ScoreExperiment(baseOutput, gapOutput, ext)
				if len(quality.PerTurn) > 0 {
					metrics["quality"] = quality
				}
			}
		}
	}

	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	return os.WriteFile(filepath.Join(expDir, "metrics.json"), append(metricsJSON, '\n'), 0o644)
}

// RunMultiProvider runs experiments across multiple providers with round-robin distribution.
func RunMultiProvider(ctx context.Context, providers []ProviderSlot, expDirs []string, flow string, skipEval bool, gapInitSpec, gapMaintainSpec string, log func(string)) (succeeded, failed int) {
	if len(providers) == 1 {
		p := providers[0]
		for _, dir := range expDirs {
			err := RunSingleExperiment(ctx, p.Provider, p.Name, p.Model, dir, flow, skipEval, gapInitSpec, gapMaintainSpec, log)
			if err != nil {
				log(fmt.Sprintf("  [%s] %s FAILED: %s", p.Name, filepath.Base(dir), err))
				failed++
			} else {
				succeeded++
			}
		}
		return
	}

	// Round-robin assignment.
	queues := make(map[string][]string)
	for i, dir := range expDirs {
		p := providers[i%len(providers)]
		queues[p.Name] = append(queues[p.Name], dir)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, p := range providers {
		dirs := queues[p.Name]
		if len(dirs) == 0 {
			continue
		}
		wg.Add(1)
		go func(p ProviderSlot, dirs []string) {
			defer wg.Done()
			for _, dir := range dirs {
				err := RunSingleExperiment(ctx, p.Provider, p.Name, p.Model, dir, flow, skipEval, gapInitSpec, gapMaintainSpec, log)
				mu.Lock()
				if err != nil {
					log(fmt.Sprintf("  [%s] %s FAILED: %s", p.Name, filepath.Base(dir), err))
					failed++
				} else {
					succeeded++
				}
				mu.Unlock()
			}
		}(p, dirs)
	}
	wg.Wait()
	return
}

// ProviderSlot pairs a provider with its name and model.
type ProviderSlot struct {
	Provider types.Provider
	Name     string
	Model    string
}

func sumTurnResults(results []gap.BaseTurnResult) (outToks, inToks, latMs int) {
	for _, r := range results {
		outToks += r.OutputTokens
		inToks += r.InputTokens
		latMs += r.LatencyMs
	}
	return
}

func sumGAPTurnResults(results []gap.GAPTurnResult) (outToks, inToks, latMs int) {
	for _, r := range results {
		outToks += r.OutputTokens
		inToks += r.InputTokens
		latMs += r.LatencyMs
	}
	return
}

func geti(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

func pctSaving(base, alt int) float64 {
	if base == 0 {
		return 0
	}
	return float64(int(1000*float64(base-alt)/float64(base))) / 10.0
}
