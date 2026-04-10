package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// particleFrames defines the animation cycle for flowing particles.
var particleFrames = []string{
	"─▸──▸──",
	"──▸──▸─",
	"───▸──▸",
	"▸───▸──",
	"─▸───▸─",
	"──▸───▸",
}

// RenderAnimatedConnector draws an animated arrow between two nodes.
// When active is true, particles flow right. When false, a static arrow is shown.
func RenderAnimatedConnector(height int, frame int, active bool) string {
	var track string
	if active {
		track = particleFrames[frame%len(particleFrames)]
	} else {
		track = " ──→ "
	}

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBlue))
	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorOverlay0))

	style := inactiveStyle
	if active {
		style = activeStyle
	}

	if height <= 1 {
		return style.Render(track)
	}

	mid := height / 2
	lines := make([]string, height)
	for i := range lines {
		if i == mid {
			lines[i] = track
		} else {
			lines[i] = strings.Repeat(" ", len(track))
		}
	}

	return style.Render(strings.Join(lines, "\n"))
}

// RenderConnector draws a static arrow between two nodes (backwards compat).
func RenderConnector(height int) string {
	return RenderAnimatedConnector(height, 0, false)
}
