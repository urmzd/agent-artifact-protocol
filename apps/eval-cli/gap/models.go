package gap

// TurnResult holds token usage and outcome for a single turn.
type TurnResult struct {
	Turn          int      `json:"turn"`
	Edit          string   `json:"edit,omitempty"`
	InputTokens   int      `json:"input_tokens"`
	OutputTokens  int      `json:"output_tokens"`
	LatencyMs     int      `json:"latency_ms"`
	OutputBytes   int      `json:"output_bytes"`
	TTFTMs        *int64   `json:"ttft_ms"`
	TTLTMs        *int64   `json:"ttlt_ms"`
	MedianITLMs   *float64 `json:"median_itl_ms"`
	Failed        bool     `json:"failed"`
	FailureReason string   `json:"failure_reason,omitempty"`
}

// BaseTurnResult is the result for one turn in the base flow.
type BaseTurnResult struct {
	TurnResult
}

// GAPTurnResult is the result for one turn in the GAP flow.
type GAPTurnResult struct {
	TurnResult
	EnvelopeParsed bool   `json:"envelope_parsed"`
	ApplySucceeded bool   `json:"apply_succeeded"`
	EnvelopeName   string `json:"envelope_name,omitempty"`
}

// ContentQualityScore holds per-turn content quality comparison metrics.
type ContentQualityScore struct {
	Turn               int      `json:"turn"`
	SequenceSimilarity float64  `json:"sequence_similarity"`
	TokenF1            float64  `json:"token_f1"`
	BaseCharCount      int      `json:"base_char_count"`
	GAPCharCount       int      `json:"gap_char_count"`
	CharDeltaPct       float64  `json:"char_delta_pct"`
	LinesAdded         int      `json:"lines_added"`
	LinesRemoved       int      `json:"lines_removed"`
	RougeL             *float64 `json:"rouge_l"`
	BLEU               *float64 `json:"bleu"`
}

// JudgeScore holds an LLM-as-judge score for a single turn.
type JudgeScore struct {
	Turn      int     `json:"turn"`
	Flow      string  `json:"flow"` // "base" or "gap"
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning,omitempty"`
}

// TurnJudgeComparison holds side-by-side judge scores for one turn.
type TurnJudgeComparison struct {
	Turn             int     `json:"turn"`
	EditInstruction  string  `json:"edit_instruction,omitempty"`
	BaseScore        float64 `json:"base_score"`
	GAPScore         float64 `json:"gap_score"`
	BaseReasoning    string  `json:"base_reasoning,omitempty"`
	GAPReasoning     string  `json:"gap_reasoning,omitempty"`
}

// ExperimentQuality holds quality scores across all turns for one experiment.
type ExperimentQuality struct {
	PerTurn                 []ContentQualityScore  `json:"per_turn"`
	MeanSequenceSimilarity  float64                `json:"mean_sequence_similarity"`
	MeanTokenF1             float64                `json:"mean_token_f1"`
	MeanRougeL              *float64               `json:"mean_rouge_l"`
	MeanBLEU                *float64               `json:"mean_bleu"`
	JudgeComparisons        []TurnJudgeComparison  `json:"judge_comparisons,omitempty"`
	MeanBaseJudge           *float64               `json:"mean_base_judge"`
	MeanGAPJudge            *float64               `json:"mean_gap_judge"`
}

// TurnPrompt pairs a turn name (e.g. "turn-1") with its prompt text.
type TurnPrompt struct {
	Name   string
	Prompt string
}
