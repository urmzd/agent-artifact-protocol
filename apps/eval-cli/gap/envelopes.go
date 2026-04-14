package gap

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var intRe = regexp.MustCompile(`\d+`)

// MakeEnvelope creates a raw envelope map.
func MakeEnvelope(artifactID string, version int, name, format string, content any) map[string]any {
	return map[string]any{
		"protocol": "gap/0.1",
		"id":       artifactID,
		"version":  version,
		"name":     name,
		"meta":     map[string]any{"format": format},
		"content":  content,
	}
}

// mutateText creates a deterministic replacement for target content.
func mutateText(target string) string {
	mutated := intRe.ReplaceAllStringFunc(target, func(m string) string {
		n, _ := strconv.Atoi(m)
		return strconv.Itoa(n + 42)
	})
	if mutated != target {
		return mutated
	}
	h := md5.Sum([]byte(target))
	return fmt.Sprintf("UPD%x_%s", h[:4], target)
}

func editReplace(content, aid, format string, sectionIDs []string) []map[string]any {
	var envs []map[string]any
	limit := min(len(sectionIDs), 4)
	for i := range limit {
		sc, ok := ExtractTargetContent(content, sectionIDs[i], format)
		if !ok {
			continue
		}
		envs = append(envs, MakeEnvelope(aid, 2+i, "edit", format, []map[string]any{
			{"op": "replace", "target": map[string]string{"type": "id", "value": sectionIDs[i]}, "content": mutateText(strings.TrimSpace(sc))},
		}))
	}
	return envs
}

func editDelete(content, aid, format string, sectionIDs []string) []map[string]any {
	var envs []map[string]any
	start := max(len(sectionIDs)-2, 0)
	for i, sid := range sectionIDs[start:] {
		_, ok := ExtractTargetContent(content, sid, format)
		if !ok {
			continue
		}
		envs = append(envs, MakeEnvelope(aid, 10+i, "edit", format, []map[string]any{
			{"op": "delete", "target": map[string]string{"type": "id", "value": sid}},
		}))
	}
	return envs
}

func editMulti(content, aid, format string, sectionIDs []string) []map[string]any {
	type pair struct {
		sid string
		sc  string
	}
	var valid []pair
	for _, sid := range sectionIDs {
		sc, ok := ExtractTargetContent(content, sid, format)
		if ok {
			valid = append(valid, pair{sid, sc})
		}
	}
	if len(valid) < 3 {
		return nil
	}
	var ops []map[string]any
	for _, p := range valid[:3] {
		text := strings.TrimSpace(p.sc)
		if len(text) > 100 {
			text = text[:100]
		}
		ops = append(ops, map[string]any{
			"op": "replace", "target": map[string]string{"type": "id", "value": p.sid}, "content": mutateText(text),
		})
	}
	return []map[string]any{MakeEnvelope(aid, 20, "edit", format, ops)}
}

type pointerEntry struct {
	path string
	val  any
}

// extractPointers recursively extracts JSON Pointer paths and values.
func extractPointers(value any, prefix string) []pointerEntry {
	var results []pointerEntry

	switch v := value.(type) {
	case map[string]any:
		for k, child := range v {
			escaped := strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
			path := prefix + "/" + escaped
			results = append(results, pointerEntry{path, child})
			results = append(results, extractPointers(child, path)...)
		}
	case []any:
		for i, child := range v {
			path := fmt.Sprintf("%s/%d", prefix, i)
			results = append(results, pointerEntry{path, child})
			results = append(results, extractPointers(child, path)...)
		}
	}
	return results
}

func editPointer(content, aid, format string) []map[string]any {
	var parsed any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil
	}

	all := extractPointers(parsed, "")
	// Filter to leaves only.
	var leaves []pointerEntry
	for _, item := range all {
		switch item.val.(type) {
		case map[string]any, []any:
			continue
		default:
			leaves = append(leaves, item)
		}
	}
	if len(leaves) == 0 {
		return nil
	}

	var envs []map[string]any
	limit := min(len(leaves), 4)
	for i := range limit {
		ptr := leaves[i].path
		val := leaves[i].val
		var newVal string
		switch v := val.(type) {
		case string:
			b, _ := json.Marshal(v + "_updated")
			newVal = string(b)
		case float64:
			b, _ := json.Marshal(v + 42)
			newVal = string(b)
		case bool:
			b, _ := json.Marshal(!v)
			newVal = string(b)
		default:
			newVal = `"null_replaced"`
		}
		envs = append(envs, MakeEnvelope(aid, 2+i, "edit", format, []map[string]any{
			{"op": "replace", "target": map[string]string{"type": "pointer", "value": ptr}, "content": newVal},
		}))
	}
	return envs
}

func synthesize(content, aid, format string) []map[string]any {
	return []map[string]any{MakeEnvelope(aid, 1, "synthesize", format, []map[string]any{
		{"body": content},
	})}
}

// GenerateAllEnvelopes generates all envelope types for an artifact.
// Returns a map of filename to list of envelope maps.
func GenerateAllEnvelopes(content, artifactID, format string, sectionIDs []string) map[string][]map[string]any {
	result := make(map[string][]map[string]any)

	result["synthesize.jsonl"] = synthesize(content, artifactID, format)

	if format == "application/json" {
		if envs := editPointer(content, artifactID, format); len(envs) > 0 {
			result["edit-pointer.jsonl"] = envs
		}
	} else {
		var validSIDs []string
		for _, s := range sectionIDs {
			if _, ok := ExtractTargetContent(content, s, format); ok {
				validSIDs = append(validSIDs, s)
			}
		}
		if len(validSIDs) > 0 {
			if envs := editReplace(content, artifactID, format, validSIDs); len(envs) > 0 {
				result["edit-replace.jsonl"] = envs
			}
			if envs := editDelete(content, artifactID, format, validSIDs); len(envs) > 0 {
				result["edit-delete.jsonl"] = envs
			}
			if envs := editMulti(content, artifactID, format, validSIDs); len(envs) > 0 {
				result["edit-multi.jsonl"] = envs
			}
		}
	}

	return result
}
