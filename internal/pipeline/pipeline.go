package pipeline

import "time"

// StageStatus represents the current state of a pipeline stage.
type StageStatus int

const (
	StatusPending StageStatus = iota
	StatusRunning
	StatusDone
	StatusFailed
)

// StageStats tracks real-time metrics for a pipeline stage.
type StageStats struct {
	BytesIn    int64
	BytesOut   int64
	LinesIn    int64
	LinesOut   int64
	Throughput float64 // bytes per second
	StartedAt  time.Time
	Duration   time.Duration
}

// Stage represents a single step in a pipeline.
type Stage struct {
	ID        string
	Name      string
	Command   string
	Args      []string
	DependsOn []string
	Status    StageStatus
	Stats     *StageStats
	Error     error
}

// Pipeline represents a complete data flow pipeline.
type Pipeline struct {
	Name   string
	Stages []*Stage
}

// NewPipeline creates a pipeline with the given name.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{Name: name}
}

// AddStage appends a stage to the pipeline.
func (p *Pipeline) AddStage(s *Stage) {
	p.Stages = append(p.Stages, s)
}
