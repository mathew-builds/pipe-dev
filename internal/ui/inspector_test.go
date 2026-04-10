package ui

import (
	"strings"
	"testing"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

func TestInspectorShowsOutput(t *testing.T) {
	rb := pipeline.NewRingBuffer(10)
	rb.Write("hello world")
	rb.Write("foo bar")

	stage := &pipeline.Stage{
		ID:      "s1",
		Name:    "grep",
		Command: "grep",
		Args:    []string{"foo"},
		Status:  pipeline.StatusRunning,
		Stats:   &pipeline.StageStats{Output: rb},
	}

	result := RenderInspector(stage, 60)
	if !strings.Contains(result, "hello world") {
		t.Error("inspector should show buffered output")
	}
	if !strings.Contains(result, "foo bar") {
		t.Error("inspector should show all buffered lines")
	}
}

func TestInspectorEmptyOutput(t *testing.T) {
	rb := pipeline.NewRingBuffer(10)
	stage := &pipeline.Stage{
		ID:      "s1",
		Name:    "grep",
		Command: "grep",
		Status:  pipeline.StatusPending,
		Stats:   &pipeline.StageStats{Output: rb},
	}

	result := RenderInspector(stage, 60)
	if result == "" {
		t.Error("inspector should render something even with no output")
	}
}
