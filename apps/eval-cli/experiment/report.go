package experiment

// BuildTokenTable builds a per-turn token comparison table from metrics.
func BuildTokenTable(metrics map[string]any) map[string]any {
	var turns []map[string]any

	// Turn 0
	bt0 := getMap(metrics, "base_turn0")
	at0 := getMap(metrics, "gap_turn0")
	turns = append(turns, map[string]any{
		"turn":               0,
		"base_input":         geti(bt0, "input_tokens"),
		"base_output":        geti(bt0, "output_tokens"),
		"base_latency_ms":    geti(bt0, "latency_ms"),
		"base_ttft_ms":       bt0["ttft_ms"],
		"base_ttlt_ms":       bt0["ttlt_ms"],
		"base_median_itl_ms": bt0["median_itl_ms"],
		"gap_input":          geti(at0, "input_tokens"),
		"gap_output":         geti(at0, "output_tokens"),
		"gap_latency_ms":     geti(at0, "latency_ms"),
		"gap_ttft_ms":        at0["ttft_ms"],
		"gap_ttlt_ms":        at0["ttlt_ms"],
		"gap_median_itl_ms":  at0["median_itl_ms"],
	})

	// Edit turns
	baseTurns := getSlice(getMap(metrics, "default_flow"), "per_turn")
	gapTurns := getSlice(getMap(metrics, "gap_flow"), "per_turn")
	limit := min(len(baseTurns), len(gapTurns))
	for i := range limit {
		bt := asMap(baseTurns[i])
		at := asMap(gapTurns[i])
		turns = append(turns, map[string]any{
			"turn":               geti(bt, "turn"),
			"base_input":         geti(bt, "input_tokens"),
			"base_output":        geti(bt, "output_tokens"),
			"base_latency_ms":    geti(bt, "latency_ms"),
			"base_ttft_ms":       bt["ttft_ms"],
			"base_ttlt_ms":       bt["ttlt_ms"],
			"base_median_itl_ms": bt["median_itl_ms"],
			"gap_input":          geti(at, "input_tokens"),
			"gap_output":         geti(at, "output_tokens"),
			"gap_latency_ms":     geti(at, "latency_ms"),
			"gap_ttft_ms":        at["ttft_ms"],
			"gap_ttlt_ms":        at["ttlt_ms"],
			"gap_median_itl_ms":  at["median_itl_ms"],
			"envelope_name":      at["envelope_name"],
			"apply_ok":           at["apply_succeeded"],
		})
	}

	// Totals
	var totalBI, totalBO, totalAI, totalAO, totalBMs, totalAMs int
	for _, t := range turns {
		totalBI += geti(t, "base_input")
		totalBO += geti(t, "base_output")
		totalAI += geti(t, "gap_input")
		totalAO += geti(t, "gap_output")
		totalBMs += geti(t, "base_latency_ms")
		totalAMs += geti(t, "gap_latency_ms")
	}

	return map[string]any{
		"turns": turns,
		"totals": map[string]any{
			"base_input":          totalBI,
			"base_output":         totalBO,
			"base_combined":       totalBI + totalBO,
			"gap_input":           totalAI,
			"gap_output":          totalAO,
			"gap_combined":        totalAI + totalAO,
			"base_latency_ms":     totalBMs,
			"gap_latency_ms":      totalAMs,
			"output_savings_pct":  pctSaving(totalBO, totalAO),
			"input_delta_pct":     pctSaving(totalBI, totalAI),
			"combined_savings_pct": pctSaving(totalBI+totalBO, totalAI+totalAO),
			"latency_savings_pct": pctSaving(totalBMs, totalAMs),
		},
	}
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key]; ok {
		if mm, ok := v.(map[string]any); ok {
			return mm
		}
	}
	return map[string]any{}
}

func getSlice(m map[string]any, key string) []any {
	if v, ok := m[key]; ok {
		if s, ok := v.([]any); ok {
			return s
		}
	}
	return nil
}

func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
