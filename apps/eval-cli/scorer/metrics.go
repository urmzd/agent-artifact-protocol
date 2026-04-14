// Package scorer provides text quality scoring and LLM-as-judge evaluation.
package scorer

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

// tokenF1 computes word-token F1 between two texts.
func tokenF1(a, b string) float64 {
	ta := wordCounts(strings.ToLower(a))
	tb := wordCounts(strings.ToLower(b))
	common := 0
	for k, va := range ta {
		if vb, ok := tb[k]; ok {
			common += min(va, vb)
		}
	}
	if common == 0 {
		return 0
	}
	sumA := sumValues(ta)
	sumB := sumValues(tb)
	precision := float64(common) / float64(max(sumB, 1))
	recall := float64(common) / float64(max(sumA, 1))
	return 2 * precision * recall / (precision + recall)
}

func wordCounts(text string) map[string]int {
	m := make(map[string]int)
	for _, w := range strings.Fields(text) {
		m[w]++
	}
	return m
}

func sumValues(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}

// sequenceSimilarity computes character-level similarity ratio (like difflib.SequenceMatcher).
func sequenceSimilarity(a, b string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	lcsLen := lcs(a, b)
	return 2.0 * float64(lcsLen) / float64(len(a)+len(b))
}

// lcs computes the length of the longest common subsequence.
func lcs(a, b string) int {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return 0
	}
	// Space-optimized LCS.
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = max(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
		clear(curr)
	}
	return prev[n]
}

// diffLineCounts counts lines added and removed between two texts.
func diffLineCounts(a, b string) (added, removed int) {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	setA := make(map[string]int)
	for _, l := range linesA {
		setA[l]++
	}
	setB := make(map[string]int)
	for _, l := range linesB {
		setB[l]++
	}

	for l, count := range setB {
		if countA, ok := setA[l]; ok {
			added += max(count-countA, 0)
		} else {
			added += count
		}
	}
	for l, count := range setA {
		if countB, ok := setB[l]; ok {
			removed += max(count-countB, 0)
		} else {
			removed += count
		}
	}
	return
}

// rougeL computes ROUGE-L F1 between reference and hypothesis.
func rougeL(reference, hypothesis string) float64 {
	refWords := strings.Fields(reference)
	hypWords := strings.Fields(hypothesis)
	if len(refWords) == 0 || len(hypWords) == 0 {
		return 0
	}
	lcsLen := lcsWords(refWords, hypWords)
	precision := float64(lcsLen) / float64(len(hypWords))
	recall := float64(lcsLen) / float64(len(refWords))
	if precision+recall == 0 {
		return 0
	}
	return 2 * precision * recall / (precision + recall)
}

func lcsWords(a, b []string) int {
	m, n := len(a), len(b)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = max(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
		clear(curr)
	}
	return prev[n]
}

// ScoreTurn computes quality metrics for one turn pair.
func ScoreTurn(baseText, gapText string, turn int) gap.ContentQualityScore {
	gapClean := gap.StripGAPMarkers(gapText)
	seqSim := sequenceSimilarity(baseText, gapClean)
	f1 := tokenF1(baseText, gapClean)
	added, removed := diffLineCounts(baseText, gapClean)
	baseChars := len(baseText)
	gapChars := len(gapClean)
	deltaPct := 0.0
	if baseChars > 0 {
		deltaPct = float64(gapChars-baseChars) / float64(baseChars) * 100
	}
	rl := rougeL(baseText, gapClean)

	return gap.ContentQualityScore{
		Turn:               turn,
		SequenceSimilarity: round4(seqSim),
		TokenF1:            round4(f1),
		BaseCharCount:      baseChars,
		GAPCharCount:       gapChars,
		CharDeltaPct:       math.Round(deltaPct*10) / 10,
		LinesAdded:         added,
		LinesRemoved:       removed,
		RougeL:             ptr(round4(rl)),
	}
}

// ScoreExperiment scores all turns for one experiment by reading artifacts from disk.
func ScoreExperiment(baseOutputDir, gapOutputDir, ext string) gap.ExperimentQuality {
	pattern := filepath.Join(baseOutputDir, "turn-*"+ext)
	baseFiles, _ := filepath.Glob(pattern)
	sort.Strings(baseFiles)

	var scores []gap.ContentQualityScore
	for _, bf := range baseFiles {
		name := filepath.Base(bf)
		af := filepath.Join(gapOutputDir, name)
		if _, err := os.Stat(af); err != nil {
			continue
		}
		baseData, _ := os.ReadFile(bf)
		gapData, _ := os.ReadFile(af)
		turn := parseTurnFromFilename(name)
		scores = append(scores, ScoreTurn(string(baseData), string(gapData), turn))
	}

	if len(scores) == 0 {
		return gap.ExperimentQuality{PerTurn: []gap.ContentQualityScore{}}
	}

	var sumSim, sumF1, sumRL float64
	rlCount := 0
	for _, s := range scores {
		sumSim += s.SequenceSimilarity
		sumF1 += s.TokenF1
		if s.RougeL != nil {
			sumRL += *s.RougeL
			rlCount++
		}
	}
	n := float64(len(scores))
	quality := gap.ExperimentQuality{
		PerTurn:                scores,
		MeanSequenceSimilarity: round4(sumSim / n),
		MeanTokenF1:            round4(sumF1 / n),
	}
	if rlCount > 0 {
		rl := round4(sumRL / float64(rlCount))
		quality.MeanRougeL = &rl
	}
	return quality
}

func parseTurnFromFilename(name string) int {
	// name like "turn-0.html"
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.SplitN(base, "-", 2)
	if len(parts) < 2 {
		return 0
	}
	n := 0
	for _, c := range parts[1] {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func ptr[T any](v T) *T {
	return &v
}
