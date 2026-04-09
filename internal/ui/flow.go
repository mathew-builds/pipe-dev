package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// RenderFlow draws the full pipeline as a horizontal node → connector → node chain.
func RenderFlow(p *pipeline.Pipeline) string {
	if len(p.Stages) == 0 {
		return ""
	}

	var parts []string
	for i, stage := range p.Stages {
		node := RenderNode(stage)
		parts = append(parts, node)

		if i < len(p.Stages)-1 {
			// Match connector height to node height.
			nodeHeight := lipgloss.Height(node)
			parts = append(parts, RenderConnector(nodeHeight))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}
