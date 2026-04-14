// Package experiment orchestrates running and reporting on GAP eval experiments.
package experiment

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

// LoadExperimentDirs finds valid experiment directories.
func LoadExperimentDirs(baseDir string) []string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	var dirs []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || e.Name() == "EXPERIMENT.md" {
			continue
		}
		dir := filepath.Join(baseDir, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "README.md")); err == nil {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs
}

// FindTurnFiles returns sorted turn-*.md files from an input directory.
func FindTurnFiles(inputDir string) []string {
	pattern := filepath.Join(inputDir, "turn-*.md")
	matches, _ := filepath.Glob(pattern)
	sort.Strings(matches)
	return matches
}

// LoadTurnPrompts reads turn files and returns TurnPrompts.
// Includes turn-0 as the first element.
func LoadTurnPrompts(inputDir string) (turn0Prompt string, editPrompts []gap.TurnPrompt, err error) {
	files := FindTurnFiles(inputDir)
	if len(files) == 0 {
		return "", nil, nil
	}

	data, err := os.ReadFile(files[0])
	if err != nil {
		return "", nil, err
	}
	turn0Prompt = strings.TrimSpace(string(data))

	for _, f := range files[1:] {
		d, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(f), ".md")
		editPrompts = append(editPrompts, gap.TurnPrompt{
			Name:   name,
			Prompt: strings.TrimSpace(string(d)),
		})
	}

	return turn0Prompt, editPrompts, nil
}
