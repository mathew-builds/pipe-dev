package ui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

const tickInterval = 100 * time.Millisecond

// tickMsg is sent on each animation tick.
type tickMsg time.Time

// Model is the main Bubbletea model for the pipeline TUI.
type Model struct {
	pipeline *pipeline.Pipeline
	runner   *pipeline.Runner
	frame    int
	done     bool
	err      error
}

// NewModel creates a TUI model for the given pipeline.
func NewModel(p *pipeline.Pipeline) Model {
	return Model{
		pipeline: p,
		runner:   pipeline.NewRunner(p),
	}
}

// pipelineEventMsg wraps a pipeline event as a tea.Msg.
type pipelineEventMsg struct {
	event pipeline.Event
	done  bool
}

// Init starts the pipeline runner and returns a command to listen for events.
func (m Model) Init() tea.Cmd {
	go m.runner.Run()
	return tea.Batch(m.waitForEvent(), m.tick())
}

// tick returns a command that sends a tickMsg after the tick interval.
func (m Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		if !m.done {
			m.frame++
			return m, m.tick()
		}
		return m, nil

	case pipelineEventMsg:
		if msg.done {
			m.done = true
			return m, tea.Quit
		}

		if msg.event.Type == pipeline.EventStageFailed && msg.event.Err != nil {
			m.err = msg.event.Err
		}

		// Continue listening for more events.
		return m, m.waitForEvent()
	}

	return m, nil
}

// View renders the TUI.
func (m Model) View() tea.View {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorMauve)).
		MarginBottom(1)

	title := titleStyle.Render(fmt.Sprintf("pipe.dev — %s", m.pipeline.Name))

	flow := RenderFlow(m.pipeline, m.frame)

	statusBar := RenderStatusBar(m.pipeline, m.done)

	parts := []string{title, flow}
	if m.done {
		var msgStyle lipgloss.Style
		var msg string
		if m.err != nil {
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorRed)).
				MarginTop(1)
			msg = msgStyle.Render("Pipeline failed.")
		} else {
			msgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorGreen)).
				MarginTop(1)
			msg = msgStyle.Render("Pipeline complete.")
		}
		parts = append(parts, msg)
	}
	parts = append(parts, statusBar)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	v := tea.NewView(content + "\n")
	return v
}

// waitForEvent returns a tea.Cmd that reads the next event from the runner.
func (m Model) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.runner.Events
		if !ok {
			return pipelineEventMsg{done: true}
		}
		return pipelineEventMsg{event: event}
	}
}
