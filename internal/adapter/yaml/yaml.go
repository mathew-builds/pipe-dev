package yaml

import (
	"fmt"
	"os"
	"strings"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
	"gopkg.in/yaml.v3"
)

// pipelineFile represents the YAML structure of a pipeline definition.
type pipelineFile struct {
	Name   string      `yaml:"name"`
	Stages []stageSpec `yaml:"stages"`
}

// stageSpec represents a single stage in the YAML file.
type stageSpec struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// Adapter parses a YAML pipeline file into a Pipeline.
type Adapter struct{}

func (a *Adapter) Name() string { return "yaml" }

func (a *Adapter) Parse(input string) (*pipeline.Pipeline, error) {
	data, err := os.ReadFile(input)
	if err != nil {
		return nil, fmt.Errorf("reading pipeline file: %w", err)
	}

	var spec pipelineFile
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing pipeline YAML: %w", err)
	}

	if len(spec.Stages) == 0 {
		return nil, fmt.Errorf("pipeline has no stages")
	}

	name := spec.Name
	if name == "" {
		name = "YAML Pipeline"
	}

	p := pipeline.NewPipeline(name)
	for i, s := range spec.Stages {
		tokens := strings.Fields(s.Command)
		if len(tokens) == 0 {
			return nil, fmt.Errorf("stage %d has empty command", i)
		}

		stageName := s.Name
		if stageName == "" {
			stageName = tokens[0]
		}

		stage := &pipeline.Stage{
			ID:      fmt.Sprintf("stage-%d", i),
			Name:    stageName,
			Command: tokens[0],
			Status:  pipeline.StatusPending,
			Stats:   &pipeline.StageStats{},
		}
		if len(tokens) > 1 {
			stage.Args = tokens[1:]
		}

		p.AddStage(stage)
	}

	return p, nil
}
