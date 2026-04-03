package main

import "fmt"

// Prompt is a test case from the exported prompt catalog.
type Prompt struct {
	ID               string   `json:"id"`
	Format           string   `json:"format"`
	Extension        string   `json:"extension"`
	Filename         string   `json:"filename"`
	SizeHint         string   `json:"size_hint"`
	ExpectedSections []string `json:"expected_sections"`
	Prompt           string   `json:"prompt"`
	Turns            []string `json:"turns"`
}

// Experiment is the full metrics output for one experiment.
type Experiment struct {
	ExperimentID string   `json:"experiment_id"`
	PromptID     string   `json:"prompt_id"`
	Model        string   `json:"model"`
	Provider     string   `json:"provider"`
	Timestamp    string   `json:"timestamp"`
	DefaultFlow  FlowData `json:"default_flow"`
	AAPFlow      AAPData  `json:"aap_flow"`
	Comparison   *Comparison `json:"comparison,omitempty"`
}

// FlowData captures default flow metrics.
type FlowData struct {
	SystemPromptTokens int           `json:"system_prompt_tokens"`
	PerTurn            []TurnMetrics `json:"per_turn"`
	TotalInputTokens   int           `json:"total_input_tokens"`
	TotalOutputTokens  int           `json:"total_output_tokens"`
	TotalLatencyMs     int64         `json:"total_latency_ms"`
}

func (f *FlowData) summarize() {
	for _, m := range f.PerTurn {
		f.TotalInputTokens += m.InputTokens
		f.TotalOutputTokens += m.OutputTokens
		f.TotalLatencyMs += m.LatencyMs
	}
}

// AAPData captures AAP flow metrics with envelope-specific fields.
type AAPData struct {
	FlowData
	MaintainSystemPromptTokens int     `json:"maintain_system_prompt_tokens"`
	EnvelopeParseRate          float64 `json:"envelope_parse_rate"`
	ApplySuccessRate           float64 `json:"apply_success_rate"`
}

func (a *AAPData) summarize() {
	a.FlowData.summarize()
	editTurns := 0
	parsed := 0
	applied := 0
	for _, m := range a.PerTurn {
		if m.Turn > 0 {
			editTurns++
			if m.EnvelopeParsed {
				parsed++
			}
			if m.ApplySucceeded {
				applied++
			}
		}
	}
	if editTurns > 0 {
		a.EnvelopeParseRate = float64(parsed) / float64(editTurns)
		a.ApplySuccessRate = float64(applied) / float64(editTurns)
	}
}

// TurnMetrics captures per-turn measurements.
type TurnMetrics struct {
	Turn             int    `json:"turn"`
	Edit             string `json:"edit"`
	InputTokens      int    `json:"input_tokens"`
	OutputTokens     int    `json:"output_tokens"`
	LatencyMs        int64  `json:"latency_ms"`
	OutputBytes      int    `json:"output_bytes"`
	EnvelopeParsed   bool   `json:"envelope_parsed,omitempty"`
	ApplySucceeded   bool   `json:"apply_succeeded,omitempty"`
	ApplyLatencyUs   int64  `json:"apply_latency_us,omitempty"`
	EnvelopeName     string `json:"envelope_name,omitempty"`
	EnvelopeOpsCount int    `json:"envelope_ops_count,omitempty"`
}

// Comparison captures the delta between flows.
type Comparison struct {
	OutputTokenSavingsPct       float64 `json:"output_token_savings_pct"`
	InputTokenSavingsPct        float64 `json:"input_token_savings_pct"`
	LatencySavingsPct           float64 `json:"latency_savings_pct"`
	BreakEvenTurn               int     `json:"break_even_turn"`
	ProtocolOverheadTokens      int     `json:"protocol_overhead_tokens"`
}

func (e *Experiment) writeComparison() {
	d := &e.DefaultFlow
	a := &e.AAPFlow
	if len(d.PerTurn) == 0 || len(a.PerTurn) == 0 {
		return
	}
	c := &Comparison{
		ProtocolOverheadTokens: a.SystemPromptTokens - d.SystemPromptTokens,
	}
	if d.TotalOutputTokens > 0 {
		c.OutputTokenSavingsPct = 100 * float64(d.TotalOutputTokens-a.TotalOutputTokens) / float64(d.TotalOutputTokens)
	}
	if d.TotalInputTokens > 0 {
		c.InputTokenSavingsPct = 100 * float64(d.TotalInputTokens-a.TotalInputTokens) / float64(d.TotalInputTokens)
	}
	if d.TotalLatencyMs > 0 {
		c.LatencySavingsPct = 100 * float64(d.TotalLatencyMs-a.TotalLatencyMs) / float64(d.TotalLatencyMs)
	}

	// Find break-even turn (cumulative output tokens)
	var defCum, aapCum int
	minTurns := len(d.PerTurn)
	if len(a.PerTurn) < minTurns {
		minTurns = len(a.PerTurn)
	}
	for t := 0; t < minTurns; t++ {
		defCum += d.PerTurn[t].OutputTokens
		aapCum += a.PerTurn[t].OutputTokens
		if aapCum < defCum && c.BreakEvenTurn == 0 {
			c.BreakEvenTurn = t
		}
	}

	e.Comparison = c
}

// defaultSystemPromptFor returns the system prompt for the default flow.
func defaultSystemPromptFor(format string) string {
	return fmt.Sprintf("You produce %s artifacts. Output raw code only. No markdown fences, no explanation.", format)
}

// aapInitSystemPromptFor returns the init-agent system prompt.
func aapInitSystemPromptFor(format string) string {
	marker := sectionMarkerInstruction(format)
	return fmt.Sprintf("You produce %s artifacts with AAP section markers for incremental updates.\n\n%s\n\nOutput raw code only. No markdown fences, no explanation.", format, marker)
}

// aapMaintainSystemPromptFor returns the maintain-agent system prompt.
func aapMaintainSystemPromptFor(format string) string {
	return fmt.Sprintf(`You are an AAP maintain-agent. You read artifacts and produce minimal AAP envelopes.

The artifact format is %s. Given an artifact and an edit instruction, produce a JSON object with these fields:

- "name": either "diff" (for small targeted changes) or "section" (for rewriting a whole section)
- "content": an array of operation objects

For name "diff", each content item has:
  {"op": "replace", "target": {"search": "exact old text"}, "content": "new text"}
  The search target MUST be an exact substring that exists in the artifact.

For name "section", each content item has:
  {"id": "section-id", "content": "new section content"}

Choose "diff" for small value changes (updating a number, changing a color).
Choose "section" for rewriting a significant block of content.

Output ONLY the JSON object. No explanation, no markdown fences.`, format)
}
