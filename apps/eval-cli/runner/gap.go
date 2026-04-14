package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	agent "github.com/urmzd/saige/agent"
	agenteval "github.com/urmzd/saige/agent/eval"
	"github.com/urmzd/saige/agent/types"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/apply"
	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

// RunGAPTurn0 runs turn-0 for the GAP flow (with target markers in the system prompt).
func RunGAPTurn0(ctx context.Context, provider types.Provider, initSystem, turn0Prompt, outputDir, ext string) (string, Turn0Metrics, error) {
	ag := agent.NewAgent(agent.AgentConfig{
		SystemPrompt: initSystem,
		Provider:     provider,
		MaxIter:      1,
	})

	t0 := time.Now()
	stream := ag.Invoke(ctx, []types.Message{types.NewUserMessage(turn0Prompt)})
	timing, rawText, deltas := agenteval.CollectStreamTiming(stream.Deltas())
	if err := stream.Wait(); err != nil {
		return "", Turn0Metrics{}, fmt.Errorf("gap turn-0: %w", err)
	}
	if streamErr := extractStreamError(deltas); streamErr != nil {
		return "", Turn0Metrics{}, fmt.Errorf("gap turn-0: %w", streamErr)
	}
	ms := int(time.Since(t0).Milliseconds())

	artifact := gap.CleanArtifact(rawText)
	_ = os.WriteFile(filepath.Join(outputDir, "turn-0"+ext), []byte(artifact), 0o644)

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

	return artifact, metrics, nil
}

// RunGAPFlow runs all edit turns for the GAP flow.
// Each turn is stateless: the LLM receives only the current artifact + edit instruction
// and returns a structured LLMEnvelope, which is applied via the Rust engine.
func RunGAPFlow(ctx context.Context, provider types.Provider, maintainSystem, artifact string, editPrompts []gap.TurnPrompt, format, outputDir, ext string) ([]gap.GAPTurnResult, string, error) {
	schema := types.SchemaFrom[gap.LLMEnvelopeSchema]()

	var results []gap.GAPTurnResult
	version := 1

	for _, tp := range editPrompts {
		turnNum := parseTurnNum(tp.Name)
		userMsg := fmt.Sprintf("## Current Artifact\n\n```\n%s\n```\n\n## Edit Instruction\n\n%s", artifact, tp.Prompt)

		// Fresh agent per turn (stateless).
		ag := agent.NewAgent(agent.AgentConfig{
			SystemPrompt:   maintainSystem,
			Provider:       provider,
			MaxIter:        1,
			ResponseSchema: &schema,
		})

		t0 := time.Now()
		stream := ag.Invoke(ctx, []types.Message{types.NewUserMessage(userMsg)})
		timing, rawText, deltas := agenteval.CollectStreamTiming(stream.Deltas())
		err := stream.Wait()
		if err == nil {
			err = extractStreamError(deltas)
		}
		ms := int(time.Since(t0).Milliseconds())

		parsed := false
		succeeded := false
		envName := ""
		envelopeJSON := ""

		if err == nil {
			// Parse the structured output as an Envelope.
			var envelope gap.Envelope
			if jsonErr := json.Unmarshal([]byte(rawText), &envelope); jsonErr == nil {
				parsed = true
				envName = envelope.Name
				prettyJSON, _ := json.MarshalIndent(envelope, "", "  ")
				envelopeJSON = string(prettyJSON)

				// Apply the envelope to the artifact.
				newArtifact, applyErr := apply.ApplyEnvelope(artifact, []byte(rawText), format, envelope.ID, envelope.Version)
				if applyErr == nil {
					succeeded = true
					artifact = newArtifact
					version++
				}
			}
		}

		if envelopeJSON != "" {
			_ = os.WriteFile(filepath.Join(outputDir, tp.Name+".json"), []byte(envelopeJSON), 0o644)
		}
		_ = os.WriteFile(filepath.Join(outputDir, tp.Name+ext), []byte(artifact), 0o644)

		tr := gap.TurnResult{
			Turn:         turnNum,
			Edit:         truncate(tp.Prompt, 80),
			InputTokens:  timing.InputTokens,
			OutputTokens: timing.OutputTokens,
			LatencyMs:    ms,
			OutputBytes:  len(artifact),
			Failed:       !succeeded,
		}
		if !succeeded {
			tr.FailureReason = "parse or apply failed"
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

		results = append(results, gap.GAPTurnResult{
			TurnResult:     tr,
			EnvelopeParsed: parsed,
			ApplySucceeded: succeeded,
			EnvelopeName:   envName,
		})
	}

	return results, artifact, nil
}
