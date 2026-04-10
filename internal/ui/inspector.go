package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

const inspectorHeight = 8 // max visible lines

// RenderInspector draws the data preview panel for a selected stage.
func RenderInspector(stage *pipeline.Stage, width int) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorMauve))

	cmd := stage.Command
	if len(stage.Args) > 0 {
		cmd += " " + strings.Join(stage.Args, " ")
	}
	header := headerStyle.Render(fmt.Sprintf("  %s", cmd))

	// Get output lines from ring buffer.
	var outputLines []string
	if stage.Stats != nil && stage.Stats.Output != nil {
		outputLines = stage.Stats.Output.Lines()
	}

	// Take last N lines that fit.
	if len(outputLines) > inspectorHeight {
		outputLines = outputLines[len(outputLines)-inspectorHeight:]
	}

	var body string
	if len(outputLines) == 0 {
		dimStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorOverlay0)).
			Italic(true)
		body = dimStyle.Render("  waiting for output...")
	} else {
		lineStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))
		var rendered []string
		for _, line := range outputLines {
			// Truncate long lines.
			maxLen := width - 4
			if len(line) > maxLen {
				line = line[:maxLen-1] + "…"
			}
			rendered = append(rendered, "  "+lineStyle.Render(line))
		}
		body = strings.Join(rendered, "\n")
	}

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorSurface1)).
		Width(width).
		MarginTop(1)

	content := header + "\n" + body
	return boxStyle.Render(content)
}
