// Package runner implements base and GAP flow runners using the saige SDK.
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	agent "github.com/urmzd/saige/agent"
	agenteval "github.com/urmzd/saige/agent/eval"
	"github.com/urmzd/saige/agent/types"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

// Turn0Metrics holds metrics for a turn-0 run.
type Turn0Metrics struct {
	InputTokens  int      `json:"input_tokens"`
	OutputTokens int      `json:"output_tokens"`
	LatencyMs    int      `json:"latency_ms"`
	ArtifactBytes int     `json:"artifact_bytes"`
	TTFTMs        *int64   `json:"ttft_ms"`
	TTLTMs        *int64   `json:"ttlt_ms"`
	MedianITLMs  *float64 `json:"median_itl_ms"`
}

// RunBaseTurn0 runs turn-0 for the base flow.
// Returns (artifact_text, agent_for_continuing, metrics).
func RunBaseTurn0(ctx context.Context, provider types.Provider, systemPrompt, turn0Prompt, outputDir, ext string) (string, *agent.Agent, Turn0Metrics, error) {
	ag := agent.NewAgent(agent.AgentConfig{
		SystemPrompt: systemPrompt,
		Provider:     provider,
		MaxIter:      1,
	})

	t0 := time.Now()
	stream := ag.Invoke(ctx, []types.Message{types.NewUserMessage(turn0Prompt)})
	timing, rawText, deltas := agenteval.CollectStreamTiming(stream.Deltas())
	if err := stream.Wait(); err != nil {
		return "", nil, Turn0Metrics{}, fmt.Errorf("base turn-0: %w", err)
	}
	if streamErr := extractStreamError(deltas); streamErr != nil {
		return "", nil, Turn0Metrics{}, fmt.Errorf("base turn-0: %w", streamErr)
	}
	ms := int(time.Since(t0).Milliseconds())

	artifact := gap.CleanArtifact(rawText)
	if err := os.WriteFile(filepath.Join(outputDir, "turn-0"+ext), []byte(artifact), 0o644); err != nil {
		return "", nil, Turn0Metrics{}, fmt.Errorf("write turn-0: %w", err)
	}

	metrics := Turn0Metrics{
		InputTokens:   timing.InputTokens,
		OutputTokens:  timing.OutputTokens,
		LatencyMs:     ms,
		ArtifactBytes: len(artifact),
	}
	if timing.TTFTMs > 0 {
		metrics.TTFTMs = &timing.TTFTMs
	}
	if timing.TTLTMs > 0 {
		metrics.TTLTMs = &timing.TTLTMs
	}
	if timing.MedianITL > 0 {
		metrics.MedianITLMs = &timing.MedianITL
	}

	return artifact, ag, metrics, nil
}

// RunBaseFlow runs all edit turns for the base flow using a continuing agent.
// The agent's tree accumulates history automatically.
func RunBaseFlow(ctx context.Context, ag *agent.Agent, editPrompts []gap.TurnPrompt, outputDir, ext string) ([]gap.BaseTurnResult, string, error) {
	var results []gap.BaseTurnResult
	var artifact string

	for _, tp := range editPrompts {
		turnNum := parseTurnNum(tp.Name)
		t0 := time.Now()

		stream := ag.Invoke(ctx, []types.Message{types.NewUserMessage(tp.Prompt)})
		timing, rawText, deltas := agenteval.CollectStreamTiming(stream.Deltas())
		err := stream.Wait()
		if err == nil {
			err = extractStreamError(deltas)
		}
		ms := int(time.Since(t0).Milliseconds())

		if err != nil {
			results = append(results, gap.BaseTurnResult{TurnResult: gap.TurnResult{
				Turn:          turnNum,
				Edit:          truncate(tp.Prompt, 80),
				LatencyMs:     ms,
				Failed:        true,
				FailureReason: err.Error(),
			}})
			continue
		}

		artifact = gap.CleanArtifact(rawText)
		_ = os.WriteFile(filepath.Join(outputDir, tp.Name+ext), []byte(artifact), 0o644)

		tr := gap.TurnResult{
			Turn:         turnNum,
			Edit:         truncate(tp.Prompt, 80),
			InputTokens:  timing.InputTokens,
			OutputTokens: timing.OutputTokens,
			LatencyMs:    ms,
			OutputBytes:  len(artifact),
		}
		if timing.TTFTMs > 0 {
			tr.TTFTMs = &timing.TTFTMs
		}
		if timing.TTLTMs > 0 {
			tr.TTLTMs = &timing.TTLTMs
		}
		if timing.MedianITL > 0 {
			tr.MedianITLMs = &timing.MedianITL
		}
		results = append(results, gap.BaseTurnResult{TurnResult: tr})
	}

	return results, artifact, nil
}
