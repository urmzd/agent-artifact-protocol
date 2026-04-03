package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateExperimentInputs creates the input directory structure for one experiment.
// No LLM needed — just writes system prompts and turn messages from the prompt catalog.
//
// Structure:
//
//	inputs/base/     — shared foundation (same prompts for both flows)
//	inputs/aap/      — protocol extension (ONLY the extra system prompts)
//	outputs/         — empty, written by `run` command
func generateExperimentInputs(p Prompt, expDir string) error {
	// ── inputs/base/ — shared by both flows ──────────────────────────────

	baseSys := defaultSystemPromptFor(p.Format)
	writeFile(filepath.Join(expDir, "inputs", "base", "system.md"), baseSys)

	// Turn 0: creation prompt (same text for both flows)
	writeFile(filepath.Join(expDir, "inputs", "base", "turn-0.md"), p.Prompt)

	// Turns 1+: edit instructions (same text for both flows)
	for i, turn := range p.Turns {
		writeFile(filepath.Join(expDir, "inputs", "base", fmt.Sprintf("turn-%d.md", i+1)), turn)
	}

	// ── inputs/aap/ — only the protocol delta ────────────────────────────
	// These are the ONLY differences from base. Two system prompts that
	// implement the protocol: one for the init-agent (adds marker instructions),
	// one for the maintain-agent (adds envelope spec).

	initSys := aapInitSystemPromptFor(p.Format)
	maintainSys := aapMaintainSystemPromptFor(p.Format)

	writeFile(filepath.Join(expDir, "inputs", "aap", "init-system.md"), initSys)
	writeFile(filepath.Join(expDir, "inputs", "aap", "maintain-system.md"), maintainSys)

	// ── outputs/ — empty dirs for run command ────────────────────────────

	os.MkdirAll(filepath.Join(expDir, "outputs", "base"), 0o755)
	os.MkdirAll(filepath.Join(expDir, "outputs", "aap"), 0o755)

	// ── README.md — experiment metadata ──────────────────────────────────

	meta := fmt.Sprintf(`# Experiment: %s

**Format:** %s | **Size:** %s | **Edits:** %d

**Expected sections:** %s

## Protocol cost (the only difference from base)

| Prompt | Chars | ~Tokens |
|---|---|---|
| Base system | %d | %d |
| AAP init system | %d | %d |
| AAP maintain system | %d | %d |
| **Protocol overhead** | | **~%d tokens** |

## Turns

| Turn | Edit |
|---|---|
| 0 | (creation) |
%s`,
		p.ID, p.Format, p.SizeHint, len(p.Turns),
		strings.Join(p.ExpectedSections, ", "),
		len(baseSys), len(baseSys)/4,
		len(initSys), len(initSys)/4,
		len(maintainSys), len(maintainSys)/4,
		(len(initSys)+len(maintainSys)-len(baseSys))/4,
		formatTurnsTable(p.Turns),
	)
	writeFile(filepath.Join(expDir, "README.md"), meta)

	return nil
}

func formatTurnsTable(turns []string) string {
	var sb strings.Builder
	for i, t := range turns {
		// Truncate long turn descriptions for the table
		desc := t
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s |\n", i+1, desc))
	}
	return sb.String()
}

// generateAllInputs creates experiment directories for all prompts.
func generateAllInputs(prompts []Prompt, baseDir string, count int) error {
	os.MkdirAll(baseDir, 0o755)

	for i := range count {
		p := prompts[i%len(prompts)]
		num := fmt.Sprintf("%03d", i+1)
		expDir := filepath.Join(baseDir, fmt.Sprintf("%s-%s", num, p.ID))

		if err := generateExperimentInputs(p, expDir); err != nil {
			return fmt.Errorf("experiment %s: %w", num, err)
		}
	}

	return nil
}
