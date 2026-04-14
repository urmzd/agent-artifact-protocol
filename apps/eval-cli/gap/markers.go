package gap

import (
	"fmt"
	"regexp"
	"strings"
)

var gapMarkerRe = regexp.MustCompile(`</?gap:target[^>]*>`)

// MarkersFor returns the (start, end) marker pair for a target ID.
// Returns ("", "", false) for JSON format (which uses pointer addressing).
func MarkersFor(targetID, format string) (start, end string, ok bool) {
	if format == "application/json" {
		return "", "", false
	}
	return fmt.Sprintf(`<gap:target id="%s">`, targetID), "</gap:target>", true
}

// MarkerExample returns a human-readable marker example for prompts.
func MarkerExample(format string) string {
	if format == "application/json" {
		return ""
	}
	return `<gap:target id="ID"> ... </gap:target>`
}

// StripGAPMarkers removes all <gap:target ...> and </gap:target> tags.
func StripGAPMarkers(text string) string {
	return gapMarkerRe.ReplaceAllString(text, "")
}

// findMatchingClose finds the position of the matching </gap:target> with depth counting.
func findMatchingClose(content string, contentStart int) int {
	const openPrefix = "<gap:target "
	const closeTag = "</gap:target>"
	depth := 1
	cursor := contentStart

	for cursor < len(content) && depth > 0 {
		nextOpen := strings.Index(content[cursor:], openPrefix)
		nextClose := strings.Index(content[cursor:], closeTag)

		if nextClose == -1 {
			return -1
		}

		// Adjust to absolute positions.
		if nextOpen != -1 {
			nextOpen += cursor
		}
		nextClose += cursor

		if nextOpen != -1 && nextOpen < nextClose {
			depth++
			cursor = nextOpen + len(openPrefix)
		} else {
			depth--
			if depth == 0 {
				return nextClose
			}
			cursor = nextClose + len(closeTag)
		}
	}

	return -1
}

// ExtractTargetContent finds the content between markers for a given target ID.
// Returns ("", false) if not found or if format is JSON.
func ExtractTargetContent(content, targetID, format string) (string, bool) {
	start, _, ok := MarkersFor(targetID, format)
	if !ok {
		return "", false
	}

	si := strings.Index(content, start)
	if si == -1 {
		return "", false
	}

	contentStart := si + len(start)
	ei := findMatchingClose(content, contentStart)
	if ei == -1 {
		return "", false
	}

	return content[contentStart:ei], true
}
