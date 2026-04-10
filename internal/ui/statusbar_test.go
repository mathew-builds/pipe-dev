package ui

import (
	"strings"
	"testing"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

func TestStatusBarShowsProgress(t *testing.T) {
	p := pipeline.NewPipeline("test")
	p.AddStage(&pipeline.Stage{ID: "1", Status: pipeline.StatusDone, Stats: &pipeline.StageStats{BytesOut: 1024}})
	p.AddStage(&pipeline.Stage{ID: "2", Status: pipeline.StatusRunning, Stats: &pipeline.StageStats{}})
	p.AddStage(&pipeline.Stage{ID: "3", Status: pipeline.StatusPending})

	result := RenderStatusBar(p, false)

	if !strings.Contains(result, "1/3") {
		t.Errorf("status bar should show stage progress, got: %s", result)
	}
}

func TestStatusBarShowsKeyHints(t *testing.T) {
	p := pipeline.NewPipeline("test")
	result := RenderStatusBar(p, false)

	if !strings.Contains(result, "q") {
		t.Error("status bar should contain quit hint")
	}
}
