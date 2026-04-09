package pipeline

// EventType represents the kind of pipeline event.
type EventType int

const (
	EventStageStarted EventType = iota
	EventStageOutput
	EventStageDone
	EventStageFailed
	EventPipelineDone
)

// Event is emitted by the runner as pipeline execution progresses.
type Event struct {
	Type    EventType
	StageID string
	Stats   *StageStats
	Output  []byte // for EventStageOutput
	Err     error  // for EventStageFailed
}
