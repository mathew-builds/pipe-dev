package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// RenderFlow draws the full pipeline as a horizontal node → connector → node chain.
func RenderFlow(p *pipeline.Pipeline, frame int, selected int) string {
	if len(p.Stages) == 0 {
		return ""
	}

	var parts []string
	for i, stage := range p.Stages {
		node := RenderNode(stage, i == selected)
		parts = append(parts, node)

		if i < len(p.Stages)-1 {
			nodeHeight := lipgloss.Height(node)
			// Connector is active if the stage feeding into it is running.
			active := stage.Status == pipeline.StatusRunning
			parts = append(parts, RenderAnimatedConnector(nodeHeight, frame, active))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}
