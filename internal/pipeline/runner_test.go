package pipeline

import (
	"testing"
	"time"
)

func buildPipeline(cmds ...string) *Pipeline {
	p := NewPipeline("test")
	for i, raw := range cmds {
		tokens := splitCommand(raw)
		stage := &Stage{
			ID:      stageID(i),
			Name:    tokens[0],
			Command: tokens[0],
			Status:  StatusPending,
			Stats:   &StageStats{},
		}
		if len(tokens) > 1 {
			stage.Args = tokens[1:]
		}
		p.AddStage(stage)
	}
	return p
}

func splitCommand(s string) []string {
	var tokens []string
	cur := ""
	for _, c := range s {
		if c == ' ' && cur != "" {
			tokens = append(tokens, cur)
			cur = ""
		} else if c != ' ' {
			cur += string(c)
		}
	}
	if cur != "" {
		tokens = append(tokens, cur)
	}
	return tokens
}

func stageID(i int) string {
	return "stage-" + string(rune('0'+i))
}

func collectEvents(r *Runner) []Event {
	var events []Event
	for e := range r.Events {
		events = append(events, e)
	}
	return events
}

func TestRunSingleStage(t *testing.T) {
	p := buildPipeline("echo hello")
	r := NewRunner(p)

	err := r.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	events := collectEvents(r)

	// Expect: StageStarted, StageDone, PipelineDone
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[0].Type != EventStageStarted {
		t.Errorf("event 0: got %v, want StageStarted", events[0].Type)
	}
	if events[1].Type != EventStageDone {
		t.Errorf("event 1: got %v, want StageDone", events[1].Type)
	}
	if events[2].Type != EventPipelineDone {
		t.Errorf("event 2: got %v, want PipelineDone", events[2].Type)
	}

	// Check stage status was updated.
	if p.Stages[0].Status != StatusDone {
		t.Errorf("stage status = %v, want StatusDone", p.Stages[0].Status)
	}
}

func TestRunTwoStages(t *testing.T) {
	// echo prints "hello\nworld\n", grep selects "world"
	p := buildPipeline("printf hello\\nworld\\n", "grep world")
	r := NewRunner(p)

	err := r.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	events := collectEvents(r)

	// Expect: 2x StageStarted, 2x StageDone, 1x PipelineDone = 5
	var started, done, pipeDone int
	for _, e := range events {
		switch e.Type {
		case EventStageStarted:
			started++
		case EventStageDone:
			done++
		case EventPipelineDone:
			pipeDone++
		}
	}

	if started != 2 {
		t.Errorf("got %d StageStarted, want 2", started)
	}
	if done != 2 {
		t.Errorf("got %d StageDone, want 2", done)
	}
	if pipeDone != 1 {
		t.Errorf("got %d PipelineDone, want 1", pipeDone)
	}
}

func TestRunThreeStages(t *testing.T) {
	// echo 3 lines, grep for "a", count lines
	p := buildPipeline("printf apple\\nbanana\\navocado\\n", "grep a", "wc -l")
	r := NewRunner(p)

	err := r.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	events := collectEvents(r)

	var started, done int
	for _, e := range events {
		switch e.Type {
		case EventStageStarted:
			started++
		case EventStageDone:
			done++
		}
	}

	if started != 3 {
		t.Errorf("got %d StageStarted, want 3", started)
	}
	if done != 3 {
		t.Errorf("got %d StageDone, want 3", done)
	}

	// First stage should have tracked output bytes.
	stats := p.Stages[0].Stats
	if stats.BytesOut == 0 {
		t.Error("stage 0 BytesOut should be > 0")
	}
	if stats.LinesOut == 0 {
		t.Error("stage 0 LinesOut should be > 0")
	}
}

func TestRunFailedCommand(t *testing.T) {
	p := buildPipeline("command_that_does_not_exist_xyz")
	r := NewRunner(p)

	r.Run()
	events := collectEvents(r)

	var failed int
	for _, e := range events {
		if e.Type == EventStageFailed {
			failed++
		}
	}

	if failed != 1 {
		t.Errorf("got %d StageFailed, want 1", failed)
	}

	if p.Stages[0].Status != StatusFailed {
		t.Errorf("stage status = %v, want StatusFailed", p.Stages[0].Status)
	}
}

func TestRunEmptyPipeline(t *testing.T) {
	p := NewPipeline("empty")
	r := NewRunner(p)

	err := r.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	events := collectEvents(r)
	if len(events) != 0 {
		t.Errorf("got %d events, want 0", len(events))
	}
}

func TestLiveStatsUpdatedDuringExecution(t *testing.T) {
	p := buildPipeline("seq 1 100000")
	r := NewRunner(p)

	go r.Run()

	// Wait for stage to start.
	e := <-r.Events
	if e.Type != EventStageStarted {
		t.Fatalf("expected StageStarted, got %v", e.Type)
	}

	// Give the command a moment to produce some output.
	time.Sleep(50 * time.Millisecond)

	// Stats should be updating in real-time.
	stats := p.Stages[0].Stats
	bytesOut := stats.LoadBytesOut()
	if bytesOut == 0 {
		t.Error("BytesOut should be > 0 during execution")
	}

	// Drain remaining events.
	for range r.Events {
	}
}

func TestByteCounterAccuracy(t *testing.T) {
	// "echo hello" outputs "hello\n" = 6 bytes, 1 line
	p := buildPipeline("echo hello")
	r := NewRunner(p)

	r.Run()
	collectEvents(r)

	stats := p.Stages[0].Stats
	if stats.LoadBytesOut() != 6 {
		t.Errorf("BytesOut = %d, want 6", stats.LoadBytesOut())
	}
	if stats.LoadLinesOut() != 1 {
		t.Errorf("LinesOut = %d, want 1", stats.LoadLinesOut())
	}
}

func TestRunFailedMiddleStage(t *testing.T) {
	// Stage 0 produces output, stage 1 is a bad command, stage 2 should still get events.
	p := buildPipeline("echo hello", "command_that_does_not_exist_xyz", "wc -l")
	r := NewRunner(p)

	// This must not deadlock. Use a timeout.
	done := make(chan error, 1)
	go func() {
		done <- r.Run()
	}()

	select {
	case <-done:
		// Good — it completed.
	case <-time.After(5 * time.Second):
		t.Fatal("Run() deadlocked — pipe leak on Start() failure")
	}

	events := collectEvents(r)

	var failed int
	for _, e := range events {
		if e.Type == EventStageFailed {
			failed++
		}
	}

	if failed == 0 {
		t.Error("expected at least one StageFailed event")
	}

	if p.Stages[1].Status != StatusFailed {
		t.Errorf("stage 1 status = %v, want StatusFailed", p.Stages[1].Status)
	}
}
