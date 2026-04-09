package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderConnector draws an arrow between two nodes.
// The height parameter matches the connector to the node height for vertical centering.
func RenderConnector(height int) string {
	arrow := " ──→ "
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorOverlay0))

	if height <= 1 {
		return style.Render(arrow)
	}

	// Vertically center the arrow.
	mid := height / 2
	lines := make([]string, height)
	for i := range lines {
		if i == mid {
			lines[i] = arrow
		} else {
			lines[i] = strings.Repeat(" ", len(arrow))
		}
	}

	return style.Render(strings.Join(lines, "\n"))
}
