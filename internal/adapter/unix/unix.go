package unix

import (
	"fmt"
	"strings"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// Adapter parses a Unix pipe chain string into a Pipeline.
type Adapter struct{}

func (a *Adapter) Name() string { return "unix" }

func (a *Adapter) Parse(input string) (*pipeline.Pipeline, error) {
	parts := strings.Split(input, "|")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty pipeline")
	}

	p := pipeline.NewPipeline("Unix Pipeline")
	for i, part := range parts {
		cmd := strings.TrimSpace(part)
		if cmd == "" {
			continue
		}

		// Split into command and args
		tokens := strings.Fields(cmd)
		stage := &pipeline.Stage{
			ID:      fmt.Sprintf("stage-%d", i),
			Name:    tokens[0],
			Command: tokens[0],
			Status:  pipeline.StatusPending,
			Stats:   &pipeline.StageStats{},
		}
		if len(tokens) > 1 {
			stage.Args = tokens[1:]
		}

		p.AddStage(stage)
	}

	if len(p.Stages) == 0 {
		return nil, fmt.Errorf("no valid stages found in pipeline")
	}

	return p, nil
}
