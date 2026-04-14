package gap

import (
	"regexp"
	"strings"
)

var thinkRe = regexp.MustCompile(`(?s)<think>.*?</think>`)

// CleanArtifact strips markdown code fences and <think> blocks from LLM output.
func CleanArtifact(text string) string {
	text = strings.TrimSpace(text)

	// Strip leading ```lang\n
	if strings.HasPrefix(text, "```") {
		if nl := strings.Index(text, "\n"); nl != -1 {
			text = text[nl+1:]
		}
	}

	// Strip trailing ```
	if strings.HasSuffix(text, "```") {
		text = strings.TrimSpace(text[:len(text)-3])
	}

	// Remove <think>...</think> blocks.
	text = thinkRe.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}
