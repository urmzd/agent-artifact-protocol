// Package liveeval runs GAP's live eval experiments on top of the generic
// saige eval harness. The harness owns the chat client, the base and
// stateless flows, corpus skipping, and output plumbing; this package keeps
// only the protocol-specific parts: the GAP flow (synthesis plus envelope
// edit turns with repair loops), the supervisor target inventory, and the
// GAP metrics document schema.
package liveeval

import (
	"context"
	"fmt"

	"github.com/urmzd/generative-artifact-protocol/evalset"
	"github.com/urmzd/saige/eval/harness"
)

// Flow names as spelled in metrics assembly and output directories.
const (
	baseFlowName      = "base"
	statelessFlowName = "stateless"
	gapFlowName       = "gap"
)

// System prompt keys used when mapping evalset experiments onto
// harness.Experiment.Systems.
const (
	baseSystemKey        = baseFlowName
	gapInitSystemKey     = "gap.init"
	gapMaintainSystemKey = "gap.maintain"
)

type Config struct {
	ExperimentsDir string
	Count          int
	IDFilter       string
	Flow           string
	Model          string
	APIBase        string
	APIKey         string
	Force          bool
}

func Run(ctx context.Context, cfg Config) error {
	flows, err := flowsFor(cfg.Flow)
	if err != nil {
		return err
	}
	inputs, err := evalset.LoadExperiments(cfg.ExperimentsDir)
	if err != nil {
		return err
	}
	experiments := harness.FilterExperiments(toHarnessExperiments(inputs), cfg.IDFilter, cfg.Count)
	client := harness.NewClient(cfg.APIBase, cfg.APIKey, cfg.Model)
	runner := &harness.Runner{
		Client: client,
		Flows:  flows,
		Force:  cfg.Force,
		Assemble: func(exp harness.Experiment, results map[string]harness.FlowResult) (any, error) {
			return assembleMetrics(client.Model, exp, results), nil
		},
	}
	return runner.Run(ctx, experiments)
}

func flowsFor(flow string) ([]harness.Flow, error) {
	switch flow {
	case baseFlowName:
		return []harness.Flow{harness.BaseFlow{}}, nil
	case statelessFlowName:
		return []harness.Flow{harness.StatelessFlow{}}, nil
	case gapFlowName:
		return []harness.Flow{gapFlow{}}, nil
	case "both":
		return []harness.Flow{harness.BaseFlow{}, gapFlow{}}, nil
	case "abc", "all":
		return []harness.Flow{harness.BaseFlow{}, harness.StatelessFlow{}, gapFlow{}}, nil
	default:
		return nil, fmt.Errorf("unknown flow %q", flow)
	}
}

func toHarnessExperiments(inputs []evalset.ExperimentInput) []harness.Experiment {
	experiments := make([]harness.Experiment, 0, len(inputs))
	for _, input := range inputs {
		turns := make([]harness.Turn, 0, len(input.Turns))
		for _, turn := range input.Turns {
			turns = append(turns, harness.Turn{Index: turn.Turn, Prompt: turn.Prompt})
		}
		experiments = append(experiments, harness.Experiment{
			ID:     input.ExperimentID,
			Format: input.Format,
			Dir:    input.Paths.ExperimentDir,
			Systems: map[string]string{
				baseSystemKey:        input.BaseSystem,
				gapInitSystemKey:     input.GAPInitSystem,
				gapMaintainSystemKey: input.GAPMaintain,
			},
			Turns: turns,
		})
	}
	return experiments
}
