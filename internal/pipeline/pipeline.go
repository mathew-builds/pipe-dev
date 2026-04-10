package pipeline

import (
	"sync/atomic"
	"time"
)

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
	BytesOut  int64
	LinesOut  int64
	StartedAt time.Time
	Duration  time.Duration
	Output    *RingBuffer // last N lines of stage output
}

// LoadBytesOut atomically reads BytesOut (safe for concurrent UI reads).
func (s *StageStats) LoadBytesOut() int64 { return atomic.LoadInt64(&s.BytesOut) }

// LoadLinesOut atomically reads LinesOut (safe for concurrent UI reads).
func (s *StageStats) LoadLinesOut() int64 { return atomic.LoadInt64(&s.LinesOut) }

// AddBytesOut atomically increments BytesOut.
func (s *StageStats) AddBytesOut(n int64) { atomic.AddInt64(&s.BytesOut, n) }

// AddLinesOut atomically increments LinesOut.
func (s *StageStats) AddLinesOut(n int64) { atomic.AddInt64(&s.LinesOut, n) }

// Stage represents a single step in a pipeline.
type Stage struct {
	ID        string
	Name      string
	Command   string
	Args      []string
	Status StageStatus
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
