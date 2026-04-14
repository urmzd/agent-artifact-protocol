package scorer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	agent "github.com/urmzd/saige/agent"
	agenteval "github.com/urmzd/saige/agent/eval"
	"github.com/urmzd/saige/agent/types"

	"github.com/urmzd/generative-artifact-protocol/eval-cli/gap"
)

const judgeSystem = `You are evaluating whether a text artifact correctly implements an edit instruction.

Given the edit instruction and the resulting artifact, score how well the edit was fulfilled.

- Score 0.0 = the edit was not applied at all
- Score 0.5 = partially applied (some elements present, others missing)
- Score 1.0 = perfectly fulfilled

Focus ONLY on whether the specific changes requested are present in the output.
Do NOT penalize for differences in style, formatting, or structure.
Do NOT penalize for the presence of XML-like markers or annotations.

Respond with a JSON object: {"score": 0.0-1.0, "reasoning": "brief explanation"}`

// JudgeOutput is the structured output from the judge LLM.
type JudgeOutput struct {
	Score     float64 `json:"score" description:"Edit fulfillment score from 0.0 to 1.0"`
	Reasoning string  `json:"reasoning" description:"Brief explanation of the score"`
}

// JudgeTurn scores a single turn's output against an edit instruction.
func JudgeTurn(ctx context.Context, provider types.Provider, editInstruction, artifactText, flow string, turn int) (gap.JudgeScore, error) {
	schema := types.SchemaFrom[JudgeOutput]()
	ag := agent.NewAgent(agent.AgentConfig{
		SystemPrompt:   judgeSystem,
		Provider:       provider,
		MaxIter:        1,
		ResponseSchema: &schema,
	})

	userMsg := fmt.Sprintf("## Edit Instruction\n\n%s\n\n## Resulting Artifact\n\n```\n%s\n```", editInstruction, artifactText)
	stream := ag.Invoke(ctx, []types.Message{types.NewUserMessage(userMsg)})
	_, rawText, _ := agenteval.CollectStreamTiming(stream.Deltas())
	if err := stream.Wait(); err != nil {
		return gap.JudgeScore{}, fmt.Errorf("judge %s turn-%d: %w", flow, turn, err)
	}

	var out JudgeOutput
	if err := json.Unmarshal([]byte(rawText), &out); err != nil {
		return gap.JudgeScore{}, fmt.Errorf("parse judge output: %w", err)
	}

	// Clamp score to [0, 1].
	if out.Score < 0 {
		out.Score = 0
	}
	if out.Score > 1 {
		out.Score = 1
	}

	return gap.JudgeScore{
		Turn:      turn,
		Flow:      flow,
		Score:     round4(out.Score),
		Reasoning: out.Reasoning,
	}, nil
}

// JudgeExperiment judges all edit turns for one experiment, scoring both flows concurrently.
func JudgeExperiment(ctx context.Context, provider types.Provider, expDir, ext string) ([]gap.TurnJudgeComparison, error) {
	baseInput := filepath.Join(expDir, "inputs", "base")
	baseOutput := filepath.Join(expDir, "outputs", "base")
	gapOutput := filepath.Join(expDir, "outputs", "gap")

	pattern := filepath.Join(baseInput, "turn-*.md")
	turnFiles, _ := filepath.Glob(pattern)
	sort.Strings(turnFiles)

	var comparisons []gap.TurnJudgeComparison
	for _, tf := range turnFiles {
		turnNum := parseTurnFromFilename(filepath.Base(tf))
		if turnNum < 1 {
			continue // skip turn-0 (creation, not edit)
		}

		turnName := fmt.Sprintf("turn-%d", turnNum)
		editInstruction, _ := os.ReadFile(tf)
		baseFile := filepath.Join(baseOutput, turnName+ext)
		gapFile := filepath.Join(gapOutput, turnName+ext)

		if _, err := os.Stat(baseFile); err != nil {
			continue
		}
		if _, err := os.Stat(gapFile); err != nil {
			continue
		}

		baseText, _ := os.ReadFile(baseFile)
		gapText := gap.StripGAPMarkers(string(must(os.ReadFile(gapFile))))

		// Judge base and GAP concurrently.
		var (
			baseScore, gapScore gap.JudgeScore
			baseErr, gapErr     error
			wg                  sync.WaitGroup
		)
		wg.Add(2)
		go func() {
			defer wg.Done()
			baseScore, baseErr = JudgeTurn(ctx, provider, string(editInstruction), string(baseText), "base", turnNum)
		}()
		go func() {
			defer wg.Done()
			gapScore, gapErr = JudgeTurn(ctx, provider, strings.TrimSpace(string(editInstruction)), gapText, "gap", turnNum)
		}()
		wg.Wait()

		if baseErr != nil || gapErr != nil {
			continue
		}

		comparisons = append(comparisons, gap.TurnJudgeComparison{
			Turn:            turnNum,
			EditInstruction: truncateStr(string(editInstruction), 120),
			BaseScore:       baseScore.Score,
			GAPScore:        gapScore.Score,
			BaseReasoning:   baseScore.Reasoning,
			GAPReasoning:    gapScore.Reasoning,
		})
	}

	return comparisons, nil
}

func must(data []byte, err error) []byte {
	if err != nil {
		return nil
	}
	return data
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
