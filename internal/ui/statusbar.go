package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// RenderStatusBar draws the bottom status bar with progress and key hints.
func RenderStatusBar(p *pipeline.Pipeline, done bool) string {
	var completed int
	var totalBytes int64
	for _, s := range p.Stages {
		if s.Status == pipeline.StatusDone {
			completed++
		}
		if s.Stats != nil {
			totalBytes += s.Stats.LoadBytesOut()
		}
	}

	progress := fmt.Sprintf(" %d/%d stages", completed, len(p.Stages))
	if totalBytes > 0 {
		progress += fmt.Sprintf("  %s processed", formatBytes(totalBytes))
	}

	hints := "q quit  tab select"
	if done {
		hints = "q quit"
	}

	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Background(lipgloss.Color(ColorSurface0)).
		Padding(0, 1)

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSubtext)).
		Background(lipgloss.Color(ColorSurface0)).
		Padding(0, 1)

	left := leftStyle.Render(progress)
	right := rightStyle.Render(hints)

	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(ColorSurface0)).
		MarginTop(1)

	return barStyle.Render(left + "  " + right)
}
