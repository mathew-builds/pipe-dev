package adapter

import "github.com/mathew-builds/pipe-dev/internal/pipeline"

// Adapter parses a pipeline definition from some source into the common Pipeline model.
type Adapter interface {
	Name() string
	Parse(input string) (*pipeline.Pipeline, error)
}
