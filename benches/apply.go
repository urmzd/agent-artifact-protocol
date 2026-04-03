package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// applyEnvelope resolves an AAP envelope against artifact content.
func applyEnvelope(content, name string, items []json.RawMessage, format string) (string, error) {
	switch name {
	case "diff":
		return applyDiff(content, items)
	case "section":
		return applySectionUpdate(content, items, format)
	default:
		return "", fmt.Errorf("unsupported operation name: %s", name)
	}
}

// ── Diff apply ───────────────────────────────────────────────────────────────

type diffOp struct {
	Op      string `json:"op"`
	Target  target `json:"target"`
	Content string `json:"content"`
}

type target struct {
	Search  string `json:"search,omitempty"`
	Section string `json:"section,omitempty"`
}

func applyDiff(content string, items []json.RawMessage) (string, error) {
	result := content
	for _, raw := range items {
		var op diffOp
		if err := json.Unmarshal(raw, &op); err != nil {
			return "", fmt.Errorf("parse diff op: %w", err)
		}

		if op.Target.Search == "" {
			return "", fmt.Errorf("diff op missing search target")
		}

		switch op.Op {
		case "replace":
			if !strings.Contains(result, op.Target.Search) {
				return "", fmt.Errorf("search target not found: %q", truncate(op.Target.Search, 80))
			}
			result = strings.Replace(result, op.Target.Search, op.Content, 1)

		case "delete":
			if !strings.Contains(result, op.Target.Search) {
				return "", fmt.Errorf("delete target not found: %q", truncate(op.Target.Search, 80))
			}
			result = strings.Replace(result, op.Target.Search, "", 1)

		case "insert_before":
			idx := strings.Index(result, op.Target.Search)
			if idx < 0 {
				return "", fmt.Errorf("insert_before target not found: %q", truncate(op.Target.Search, 80))
			}
			result = result[:idx] + op.Content + result[idx:]

		case "insert_after":
			idx := strings.Index(result, op.Target.Search)
			if idx < 0 {
				return "", fmt.Errorf("insert_after target not found: %q", truncate(op.Target.Search, 80))
			}
			end := idx + len(op.Target.Search)
			result = result[:end] + op.Content + result[end:]

		default:
			return "", fmt.Errorf("unknown diff op: %s", op.Op)
		}
	}
	return result, nil
}

// ── Section update apply ─────────────────────────────────────────────────────

type sectionUpdate struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func applySectionUpdate(content string, items []json.RawMessage, format string) (string, error) {
	result := content
	for _, raw := range items {
		var su sectionUpdate
		if err := json.Unmarshal(raw, &su); err != nil {
			return "", fmt.Errorf("parse section update: %w", err)
		}

		start, end := sectionMarkers(su.ID, format)
		re := regexp.MustCompile(
			regexp.QuoteMeta(start) + `[\s\S]*?` + regexp.QuoteMeta(end),
		)

		replaced := re.ReplaceAllString(result, start+"\n"+su.Content+"\n"+end)
		if replaced == result {
			return "", fmt.Errorf("section %q not found in content", su.ID)
		}
		result = replaced
	}
	return result, nil
}

// sectionMarkers returns start/end markers for a section ID based on MIME type.
func sectionMarkers(id, format string) (string, string) {
	switch {
	case strings.HasPrefix(format, "text/html"),
		strings.HasPrefix(format, "text/markdown"),
		strings.HasPrefix(format, "image/svg"),
		strings.Contains(format, "+xml"):
		return fmt.Sprintf("<!-- section:%s -->", id), fmt.Sprintf("<!-- /section:%s -->", id)

	case strings.Contains(format, "javascript"),
		strings.Contains(format, "typescript"),
		strings.Contains(format, "x-rust"),
		strings.Contains(format, "x-go"),
		strings.Contains(format, "x-java"),
		strings.Contains(format, "x-c"),
		strings.Contains(format, "css"):
		return fmt.Sprintf("// #region %s", id), fmt.Sprintf("// #endregion %s", id)

	case strings.Contains(format, "x-python"),
		strings.Contains(format, "x-ruby"),
		strings.Contains(format, "x-sh"),
		strings.Contains(format, "yaml"):
		return fmt.Sprintf("# region %s", id), fmt.Sprintf("# endregion %s", id)

	default:
		return fmt.Sprintf("<!-- section:%s -->", id), fmt.Sprintf("<!-- /section:%s -->", id)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
