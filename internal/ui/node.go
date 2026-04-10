package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

const nodeWidth = 28

// RenderNode draws a single pipeline stage as a styled box.
func RenderNode(stage *pipeline.Stage) string {
	var (
		borderColor color.Color
		statusIcon  string
		statusText  string
		statsLine   string
	)

	switch stage.Status {
	case pipeline.StatusPending:
		borderColor = lipgloss.Color(ColorOverlay0)
		statusIcon = "○"
		statusText = "pending"
	case pipeline.StatusRunning:
		borderColor = lipgloss.Color(ColorBlue)
		statusIcon = "◉"
		statusText = "running"
	case pipeline.StatusDone:
		borderColor = lipgloss.Color(ColorGreen)
		statusIcon = "✓"
		statusText = "done"
	case pipeline.StatusFailed:
		borderColor = lipgloss.Color(ColorRed)
		statusIcon = "✗"
		statusText = "failed"
	}

	// Build command display (truncate if needed).
	cmd := stage.Command
	if len(stage.Args) > 0 {
		cmd += " " + strings.Join(stage.Args, " ")
	}
	maxCmd := nodeWidth - 4
	if len(cmd) > maxCmd {
		cmd = cmd[:maxCmd-1] + "…"
	}

	// Status line with color.
	statusStyle := lipgloss.NewStyle().Foreground(borderColor)
	status := statusStyle.Render(statusIcon + " " + statusText)

	// Stats line (only when running or done).
	if stage.Stats != nil && stage.Status >= pipeline.StatusRunning {
		s := stage.Stats
		if stage.Status == pipeline.StatusDone {
			statsLine = fmt.Sprintf("%s  %s  %s",
				formatBytes(s.LoadBytesOut()),
				formatLines(s.LoadLinesOut()),
				formatDuration(s.Duration),
			)
		} else {
			statsLine = fmt.Sprintf("%s  %s",
				formatBytes(s.LoadBytesOut()),
				formatLines(s.LoadLinesOut()),
			)
		}
	}

	// Compose the node content.
	content := cmd + "\n" + status
	if statsLine != "" {
		statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSubtext))
		content += "\n" + statsStyle.Render(statsLine)
	}

	// Box style.
	box := lipgloss.NewStyle().
		Width(nodeWidth).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	return box.Render(content)
}
