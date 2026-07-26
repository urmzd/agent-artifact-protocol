package liveeval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gap "github.com/urmzd/generative-artifact-protocol"
	"github.com/urmzd/saige/eval/harness"
)

const (
	maxSynthesisAttempts = 2
	maxEditAttempts      = 3
)

// gapFlow drives an experiment through the GAP protocol: turn 0 synthesizes
// a marked artifact, later turns request a JSON GAP edit envelope
// (constrained via json_schema), validate it against the supervisor target
// inventory, and apply it, with repair loops on parse, validation, and apply
// failures. Outputs land in <exp.Dir>/outputs/gap.
type gapFlow struct{}

func (gapFlow) Name() string { return gapFlowName }

func (gapFlow) Run(ctx context.Context, client *harness.Client, exp harness.Experiment, _ *harness.FlowContext) (harness.FlowResult, error) {
	outDir := filepath.Join(exp.Dir, "outputs", gapFlowName)
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return harness.FlowResult{}, err
	}
	ext := harness.FormatExt(exp.Format)
	t0, artifact, err := runGAPSynthesis(ctx, client, exp)
	if err != nil {
		return harness.FlowResult{}, err
	}
	if err := harness.WriteText(filepath.Join(outDir, "turn-0"+ext), artifact); err != nil {
		return harness.FlowResult{}, err
	}

	var turns []harness.TurnResult
	version := uint64(1)
	for _, turn := range exp.Turns[1:] {
		tr, nextArtifact, nextVersion, envelope, envelopeText := runGAPEditTurn(ctx, client, exp, artifact, version, turn)
		if envelope != nil {
			_ = harness.WriteJSON(envelopePath(outDir, turn.Index, ext), *envelope)
		} else if envelopeText != "" {
			_ = harness.WriteText(envelopePath(outDir, turn.Index, ext), envelopeText)
		}
		artifact = nextArtifact
		version = nextVersion
		if err := harness.WriteText(filepath.Join(outDir, fmt.Sprintf("turn-%d%s", turn.Index, ext)), artifact); err != nil {
			return harness.FlowResult{Turn0: t0, Turns: turns, Artifact: artifact}, err
		}
		turns = append(turns, tr)
	}
	return harness.FlowResult{Turn0: t0, Turns: turns, Artifact: artifact}, nil
}

func runGAPSynthesis(ctx context.Context, client *harness.Client, exp harness.Experiment) (harness.TurnMetrics, string, error) {
	start := time.Now()
	messages := []harness.Message{
		{Role: "system", Content: exp.Systems[gapInitSystemKey]},
		{Role: "user", Content: exp.Turns[0].Prompt},
	}
	var aggregate harness.ChatResult
	var artifact string
	for attempt := 0; attempt < maxSynthesisAttempts; attempt++ {
		result, err := client.Chat(ctx, messages)
		addChatResult(&aggregate, result)
		if err != nil {
			return harness.TurnMetrics{}, "", err
		}
		artifact = harness.CleanArtifact(result.Text)
		validationErr := validateSynthesisArtifact(artifact, exp.Format)
		if validationErr == nil {
			break
		}
		messages = append(messages,
			harness.Message{Role: "assistant", Content: result.Text},
			harness.Message{Role: "user", Content: synthesisRepairPrompt(artifact, exp.Format, validationErr)},
		)
	}
	return turnMetrics(aggregate, start, artifact), artifact, nil
}

func runGAPEditTurn(ctx context.Context, client *harness.Client, exp harness.Experiment, artifact string, version uint64, turn harness.Turn) (harness.TurnResult, string, uint64, *gap.Envelope, string) {
	start := time.Now()
	messages := []harness.Message{
		{Role: "system", Content: exp.Systems[gapMaintainSystemKey]},
		{Role: "user", Content: editPrompt(artifact, exp.Format, turn.Prompt)},
	}
	var aggregate harness.ChatResult
	var parsed bool
	var applied bool
	var envelopeName string
	var repairAttempts int
	var lastErr error
	var validationErr error
	var lastEnvelope *gap.Envelope
	var lastEnvelopeText string

	for attempt := 0; attempt < maxEditAttempts; attempt++ {
		result, err := client.Chat(ctx, messages, harness.WithJSONSchema("gap_envelope", envelopeSchema()))
		addChatResult(&aggregate, result)
		if err != nil {
			lastErr = err
			break
		}
		envelopeText := harness.CleanArtifact(result.Text)
		lastEnvelopeText = envelopeText
		envelope, parseErr := gap.EnvelopeFromJSON([]byte(harness.ExtractJSONObject(envelopeText)))
		if parseErr != nil {
			lastErr = fmt.Errorf("envelope parse failed: %w", parseErr)
			validationErr = lastErr
			repairAttempts++
			messages = appendRepairMessages(messages, result.Text, artifact, exp.Format, turn.Prompt, lastErr)
			continue
		}
		parsed = true
		envelope = normalizeEnvelope(envelope, exp, version+1)
		envelopeName = string(envelope.Name)
		lastEnvelope = &envelope
		if err := validateEnvelope(artifact, exp.Format, envelope); err != nil {
			lastErr = err
			validationErr = err
			repairAttempts++
			messages = appendRepairMessages(messages, result.Text, artifact, exp.Format, turn.Prompt, err)
			continue
		}
		newArtifact, _, applyErr := gap.Apply(&gap.Artifact{
			ID:      exp.ID,
			Version: version,
			Format:  exp.Format,
			Body:    artifact,
		}, envelope)
		if applyErr != nil {
			lastErr = fmt.Errorf("apply failed: %w", applyErr)
			validationErr = lastErr
			repairAttempts++
			messages = appendRepairMessages(messages, result.Text, artifact, exp.Format, turn.Prompt, lastErr)
			continue
		}
		tr := turnResult(turn, aggregate, start, newArtifact.Body)
		applied = true
		tr.RepairAttempts = repairAttempts
		if validationErr != nil {
			value := validationErr.Error()
			tr.ValidationError = &value
		}
		tr.OutputBytes = len(newArtifact.Body)
		tr.Extra = gapTurnExtra(parsed, applied, envelopeName)
		return tr, newArtifact.Body, newArtifact.Version, &envelope, lastEnvelopeText
	}

	tr := turnResult(turn, aggregate, start, artifact)
	tr.Failed = true
	if lastErr != nil {
		reason := lastErr.Error()
		if !strings.HasPrefix(reason, "apply failed:") && !strings.HasPrefix(reason, "envelope parse failed:") {
			reason = "validation failed: " + reason
		}
		tr.FailureReason = &reason
	}
	tr.RepairAttempts = repairAttempts
	if validationErr != nil {
		value := validationErr.Error()
		tr.ValidationError = &value
	}
	tr.Extra = gapTurnExtra(parsed, applied, envelopeName)
	return tr, artifact, version, lastEnvelope, lastEnvelopeText
}

// gapTurnExtra carries the GAP-specific per-turn keys; the harness flattens
// them into the TurnResult JSON object exactly as before.
func gapTurnExtra(parsed bool, applied bool, envelopeName string) map[string]any {
	return map[string]any{
		"envelope_parsed": parsed,
		"apply_succeeded": applied,
		"envelope_name":   envelopeName,
	}
}

func appendRepairMessages(messages []harness.Message, assistantText string, artifact string, format string, instruction string, err error) []harness.Message {
	return append(messages,
		harness.Message{Role: "assistant", Content: assistantText},
		harness.Message{Role: "user", Content: repairPrompt(artifact, format, instruction, err)},
	)
}

func addChatResult(total *harness.ChatResult, result harness.ChatResult) {
	total.Text = result.Text
	total.InputTokens += result.InputTokens
	total.OutputTokens += result.OutputTokens
	total.CachedInputTokens += result.CachedInputTokens
	total.Retried = total.Retried || result.Retried
}

func normalizeEnvelope(envelope gap.Envelope, exp harness.Experiment, version uint64) gap.Envelope {
	envelope.Protocol = gap.ProtocolVersion
	envelope.ID = exp.ID
	if envelope.Version < version {
		envelope.Version = version
	}
	format := exp.Format
	envelope.Meta.Format = &format
	return envelope
}

func turnMetrics(result harness.ChatResult, start time.Time, artifact string) harness.TurnMetrics {
	return harness.TurnMetrics{
		InputTokens:       result.InputTokens,
		OutputTokens:      result.OutputTokens,
		CachedInputTokens: result.CachedInputTokens,
		LatencyMS:         uint64(time.Since(start).Milliseconds()),
		ArtifactBytes:     len(artifact),
	}
}

func turnResult(turn harness.Turn, result harness.ChatResult, start time.Time, artifact string) harness.TurnResult {
	retried := result.Retried
	return harness.TurnResult{
		Turn:              turn.Index,
		Edit:              harness.Truncate(turn.Prompt, 80),
		InputTokens:       result.InputTokens,
		OutputTokens:      result.OutputTokens,
		CachedInputTokens: result.CachedInputTokens,
		LatencyMS:         uint64(time.Since(start).Milliseconds()),
		OutputBytes:       len(artifact),
		Retried:           &retried,
		Failed:            false,
	}
}

func envelopePath(outDir string, turn int, ext string) string {
	if ext == ".json" {
		return filepath.Join(outDir, fmt.Sprintf("turn-%d.envelope.json", turn))
	}
	return filepath.Join(outDir, fmt.Sprintf("turn-%d.json", turn))
}

// truncate is kept for the supervisor prompt builders.
func truncate(s string, maxLen int) string {
	return harness.Truncate(s, maxLen)
}

// envelopeSchema is the strict JSON schema for GAP edit envelopes, sent via
// response_format json_schema on every edit turn.
func envelopeSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"protocol", "id", "version", "name", "meta", "content"},
		"properties": map[string]any{
			"protocol": map[string]any{"type": "string"},
			"id":       map[string]any{"type": "string"},
			"version":  map[string]any{"type": "integer"},
			"name":     map[string]any{"type": "string", "enum": []string{"edit"}},
			"meta": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"format", "tokens_used", "checksum", "state"},
				"properties": map[string]any{
					"format":      map[string]any{"type": []string{"string", "null"}},
					"tokens_used": map[string]any{"type": []string{"integer", "null"}},
					"checksum":    map[string]any{"type": []string{"string", "null"}},
					"state":       map[string]any{"type": []string{"string", "null"}},
				},
			},
			"content": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"op", "target", "content"},
					"properties": map[string]any{
						"op": map[string]any{
							"type": "string",
							"enum": []string{"replace", "insert_before", "insert_after", "delete"},
						},
						"target": map[string]any{
							"type":                 "object",
							"additionalProperties": false,
							"required":             []string{"type", "value"},
							"properties": map[string]any{
								"type":  map[string]any{"type": "string", "enum": []string{"id", "pointer"}},
								"value": map[string]any{"type": "string"},
							},
						},
						"content": map[string]any{"type": []string{"string", "null"}},
					},
				},
			},
		},
	}
}
