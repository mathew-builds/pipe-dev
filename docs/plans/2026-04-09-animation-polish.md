# Animation & Polish Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform the static MVP TUI into an animated, interactive pipeline visualizer that creates a "holy cow" moment for the demo GIF.

**Architecture:** Add a tick-based animation loop to the Bubbletea model, animated `>>>` particle connectors between nodes, live-updating byte/line counters during execution, a status bar with key hints, and an inspector panel showing live data at the selected stage. Fix the runner pipe-leak bug first.

**Tech Stack:** Go 1.25, Bubbletea v2 (`charm.land/bubbletea/v2`), Lipgloss v2 (`charm.land/lipgloss/v2`), sync/atomic for thread-safe counters.

---

## Task 1: Fix runner pipe leak on Start() failure

**Why:** When `cmd.Start()` fails on stage i, the pipe writer from the previous stage is never closed. The previous stage blocks forever trying to write to a full pipe. This is a deadlock bug.

**Files:**
- Modify: `internal/pipeline/runner.go:83-88`
- Test: `internal/pipeline/runner_test.go`

**Step 1: Write the failing test**

Add to `runner_test.go`:

```go
func TestRunFailedMiddleStage(t *testing.T) {
	// Stage 0 produces output, stage 1 is a bad command, stage 2 should still get events.
	p := buildPipeline("echo hello", "command_that_does_not_exist_xyz", "wc -l")
	r := NewRunner(p)

	// This must not deadlock. Use a timeout.
	done := make(chan error, 1)
	go func() {
		done <- r.Run()
	}()

	select {
	case <-done:
		// Good — it completed.
	case <-time.After(5 * time.Second):
		t.Fatal("Run() deadlocked — pipe leak on Start() failure")
	}

	events := collectEvents(r)

	var failed int
	for _, e := range events {
		if e.Type == EventStageFailed {
			failed++
		}
	}

	if failed == 0 {
		t.Error("expected at least one StageFailed event")
	}

	if p.Stages[1].Status != StatusFailed {
		t.Errorf("stage 1 status = %v, want StatusFailed", p.Stages[1].Status)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pipeline/ -run TestRunFailedMiddleStage -v -timeout 10s`
Expected: FAIL (deadlock or timeout)

**Step 3: Write minimal fix**

In `runner.go`, after the `cmd.Start()` failure block (line 83-88), close the pipe writer for the failed stage before continuing:

```go
if err := cmd.Start(); err != nil {
	stage.Status = StatusFailed
	stage.Error = err
	r.emit(Event{Type: EventStageFailed, StageID: stage.ID, Err: err})

	// Close the pipe writer so the previous stage gets a write error
	// instead of blocking forever.
	if i < len(cmds)-1 {
		if cw, ok := cmds[i].Stdout.(*countingWriter); ok {
			if pw, ok := cw.w.(*io.PipeWriter); ok {
				pw.CloseWithError(err)
			}
		}
	}
	// Also close the pipe reader so the previous stage's pipe writer
	// doesn't block. The stdin for this stage is a pipe reader.
	if i > 0 {
		if pr, ok := cmds[i].Stdin.(*io.PipeReader); ok {
			pr.CloseWithError(err)
		}
	}
	continue
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pipeline/ -run TestRunFailedMiddleStage -v -timeout 10s`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -v`
Expected: All 22 tests pass (21 existing + 1 new)

**Step 6: Commit**

```bash
git add internal/pipeline/runner.go internal/pipeline/runner_test.go
git commit -m "fix: close pipes on stage Start() failure to prevent deadlock"
```

---

## Task 2: Real-time stats with atomic counters

**Why:** Currently `stage.Stats.BytesOut` is only set after `cmd.Wait()` completes. The UI can't show live counters during execution. We need the countingWriter to update stats in real-time using atomic operations (safe for concurrent read from UI goroutine).

**Files:**
- Modify: `internal/pipeline/runner.go`
- Modify: `internal/pipeline/pipeline.go`
- Test: `internal/pipeline/runner_test.go`

**Step 1: Write the failing test**

Add to `runner_test.go`:

```go
func TestLiveStatsUpdatedDuringExecution(t *testing.T) {
	// Use a slow command so we can check stats mid-execution.
	// "seq 1 100000" produces enough output to check before it finishes.
	p := buildPipeline("seq 1 100000")
	r := NewRunner(p)

	go r.Run()

	// Wait for stage to start.
	e := <-r.Events
	if e.Type != EventStageStarted {
		t.Fatalf("expected StageStarted, got %v", e.Type)
	}

	// Give the command a moment to produce some output.
	time.Sleep(50 * time.Millisecond)

	// Stats should be updating in real-time (not zero while running).
	stats := p.Stages[0].Stats
	bytesOut := stats.LoadBytesOut()
	if bytesOut == 0 {
		t.Error("BytesOut should be > 0 during execution")
	}

	// Drain remaining events.
	for range r.Events {
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pipeline/ -run TestLiveStatsUpdatedDuringExecution -v`
Expected: FAIL — `LoadBytesOut` method doesn't exist yet

**Step 3: Add atomic accessors to StageStats**

In `pipeline.go`, add atomic load/store methods:

```go
import "sync/atomic"

// StageStats tracks real-time metrics for a pipeline stage.
type StageStats struct {
	BytesIn    int64
	BytesOut   int64
	LinesIn    int64
	LinesOut   int64
	Throughput float64 // bytes per second
	StartedAt  time.Time
	Duration   time.Duration
}

// Atomic accessors for concurrent read/write safety.
func (s *StageStats) LoadBytesOut() int64  { return atomic.LoadInt64(&s.BytesOut) }
func (s *StageStats) LoadLinesOut() int64  { return atomic.LoadInt64(&s.LinesOut) }
func (s *StageStats) AddBytesOut(n int64)  { atomic.AddInt64(&s.BytesOut, n) }
func (s *StageStats) AddLinesOut(n int64)  { atomic.AddInt64(&s.LinesOut, n) }
```

**Step 4: Update countingWriter to write directly to StageStats**

In `runner.go`, modify the countingWriter to hold a reference to StageStats and update atomically:

```go
// countingWriter wraps a writer and counts bytes/lines passing through.
type countingWriter struct {
	w     io.Writer
	stats *StageStats
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	if n > 0 && cw.stats != nil {
		cw.stats.AddBytesOut(int64(n))
		cw.stats.AddLinesOut(int64(bytes.Count(p[:n], []byte{'\n'})))
	}
	return n, err
}
```

Update the runner's `Run()` method to pass stats directly:

Where counters are created (around line 40), remove the separate `byteCounter` struct. Instead, when wiring stages (lines 43-57), create countingWriters with the stage's Stats:

```go
// Build commands and ensure stats are initialized.
for i, stage := range stages {
	cmds[i] = exec.Command(stage.Command, stage.Args...)
	stage.Stats = &StageStats{StartedAt: time.Now()}
}

// Chain stages: stage[i].stdout → countingWriter → stage[i+1].stdin
for i := 0; i < len(cmds)-1; i++ {
	pr, pw := io.Pipe()
	cmds[i].Stdout = &countingWriter{w: pw, stats: stages[i].Stats}
	cmds[i+1].Stdin = pr
	_ = pw
}

// Last stage writes to a counting writer backed by a buffer.
lastBuf := &bytes.Buffer{}
cmds[len(cmds)-1].Stdout = &countingWriter{w: lastBuf, stats: stages[len(cmds)-1].Stats}
```

Remove the separate `byteCounter` struct entirely. Remove the line that sets `stage.Stats.BytesOut = counter.bytes` and `stage.Stats.LinesOut = counter.lines` after `cmd.Wait()` — they're now updated in real-time.

In the goroutine after `cmd.Wait()`, just set Duration:

```go
go func(idx int) {
	defer wg.Done()
	err := cmd.Wait()
	stage.Stats.Duration = time.Since(stage.Stats.StartedAt)

	// Close pipe writer so next stage gets EOF.
	if idx < len(cmds)-1 {
		if cw, ok := cmds[idx].Stdout.(*countingWriter); ok {
			if pw, ok := cw.w.(*io.PipeWriter); ok {
				pw.Close()
			}
		}
	}

	if err != nil {
		stage.Status = StatusFailed
		stage.Error = err
		r.emit(Event{
			Type:    EventStageFailed,
			StageID: stage.ID,
			Err:     err,
			Output:  stderrBuf.Bytes(),
		})
	} else {
		stage.Status = StatusDone
		r.emit(Event{
			Type:    EventStageDone,
			StageID: stage.ID,
			Stats:   stage.Stats,
		})
	}
}(i)
```

Also: move `stage.Stats = &StageStats{StartedAt: time.Now()}` to the build phase (before Start), and set `stage.Status = StatusRunning` right before `cmd.Start()`.

**Step 5: Update node.go to use atomic accessors**

In `node.go`, lines 58-70, change `s.BytesOut` to `s.LoadBytesOut()` and `s.LinesOut` to `s.LoadLinesOut()`:

```go
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
```

**Step 6: Run tests**

Run: `go test ./... -v`
Expected: All tests pass. The `TestByteCounterAccuracy` test should still pass because atomic adds produce the same result.

**Step 7: Commit**

```bash
git add internal/pipeline/pipeline.go internal/pipeline/runner.go internal/pipeline/runner_test.go internal/ui/node.go
git commit -m "feat: real-time stats via atomic counters for live UI updates"
```

---

## Task 3: Add tick-based animation loop

**Why:** The Bubbletea model currently only updates on pipeline events. Without a tick loop, the UI can't animate connectors or show live counter updates between events. This is the foundation for all visual polish.

**Files:**
- Modify: `internal/ui/model.go`

**Step 1: Add tick message type and frame counter**

```go
import "time"

const tickInterval = 100 * time.Millisecond

type tickMsg time.Time

type Model struct {
	pipeline *pipeline.Pipeline
	runner   *pipeline.Runner
	done     bool
	err      error
	frame    int
}
```

**Step 2: Update Init() to start tick loop alongside event listener**

```go
func (m Model) Init() tea.Cmd {
	go m.runner.Run()
	return tea.Batch(m.waitForEvent(), m.tick())
}

func (m Model) tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
```

**Step 3: Handle tickMsg in Update()**

```go
case tickMsg:
	if !m.done {
		m.frame++
		return m, m.tick()
	}
	return m, nil
```

**Step 4: Build and run demo to verify**

Run: `make demo`
Expected: Demo runs. The tick loop increments frames but there's no visible change yet (animation comes in Task 4). No crashes.

**Step 5: Commit**

```bash
git add internal/ui/model.go
git commit -m "feat: add tick-based animation loop to TUI model"
```

---

## Task 4: Animated connectors

**Why:** The static ` ──→ ` arrows need to become animated `>>>` particles flowing between nodes. This is THE visual that makes pipe.dev feel alive — the core of the demo GIF.

**Files:**
- Modify: `internal/ui/connector.go`
- Modify: `internal/ui/flow.go` (pass frame to connector)
- Modify: `internal/ui/model.go` (pass frame to flow)

**Step 1: Write a test for the connector animation frames**

Create `internal/ui/connector_test.go`:

```go
package ui

import "testing"

func TestAnimatedConnectorFrames(t *testing.T) {
	// Different frames should produce different output.
	frame0 := RenderAnimatedConnector(3, 0, true)
	frame1 := RenderAnimatedConnector(3, 1, true)
	frame2 := RenderAnimatedConnector(3, 2, true)

	if frame0 == frame1 && frame1 == frame2 {
		t.Error("all frames are identical — no animation")
	}

	// Non-active connector should be static.
	static0 := RenderAnimatedConnector(3, 0, false)
	static1 := RenderAnimatedConnector(3, 1, false)
	if static0 != static1 {
		t.Error("non-active connector should not animate")
	}
}

func TestAnimatedConnectorContainsParticles(t *testing.T) {
	result := RenderAnimatedConnector(3, 0, true)
	if result == "" {
		t.Error("connector should not be empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestAnimatedConnector -v`
Expected: FAIL — `RenderAnimatedConnector` doesn't exist

**Step 3: Implement animated connector**

Replace `connector.go` contents:

```go
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const connectorWidth = 7 // visual width of the connector track

// particleFrames defines the animation cycle for flowing particles.
// Each string is one frame of the connector track.
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
```

**Step 4: Run connector tests**

Run: `go test ./internal/ui/ -run TestAnimatedConnector -v`
Expected: PASS

**Step 5: Update flow.go to pass frame and active state**

```go
package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// RenderFlow draws the full pipeline as a horizontal node → connector → node chain.
func RenderFlow(p *pipeline.Pipeline, frame int) string {
	if len(p.Stages) == 0 {
		return ""
	}

	var parts []string
	for i, stage := range p.Stages {
		node := RenderNode(stage)
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
```

**Step 6: Update model.go View() to pass frame**

Change line 74 in model.go:

```go
flow := RenderFlow(m.pipeline, m.frame)
```

**Step 7: Build and run demo**

Run: `make demo`
Expected: You should see `>>>` style particles flowing between nodes while stages are running. Static arrows for completed/pending connectors.

**Step 8: Commit**

```bash
git add internal/ui/connector.go internal/ui/connector_test.go internal/ui/flow.go internal/ui/model.go
git commit -m "feat: animated flowing particles between pipeline stages"
```

---

## Task 5: Status bar

**Why:** Users need to see overall pipeline progress and know what keys do. The status bar anchors the bottom of the TUI.

**Files:**
- Create: `internal/ui/statusbar.go`
- Test: `internal/ui/statusbar_test.go`
- Modify: `internal/ui/model.go` (render status bar instead of plain footer)

**Step 1: Write the test**

Create `internal/ui/statusbar_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestStatusBar -v`
Expected: FAIL — `RenderStatusBar` doesn't exist

**Step 3: Implement status bar**

Create `internal/ui/statusbar.go`:

```go
package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

// RenderStatusBar draws the bottom status bar with progress and key hints.
func RenderStatusBar(p *pipeline.Pipeline, done bool) string {
	// Count completed stages.
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

	// Left side: progress.
	progress := fmt.Sprintf(" %d/%d stages", completed, len(p.Stages))
	if totalBytes > 0 {
		progress += fmt.Sprintf("  %s processed", formatBytes(totalBytes))
	}

	// Right side: key hints.
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
```

**Step 4: Run tests**

Run: `go test ./internal/ui/ -run TestStatusBar -v`
Expected: PASS

**Step 5: Integrate into model.go View()**

Replace the footer section in `View()`:

```go
func (m Model) View() tea.View {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorMauve)).
		MarginBottom(1)

	title := titleStyle.Render(fmt.Sprintf("pipe.dev — %s", m.pipeline.Name))
	flow := RenderFlow(m.pipeline, m.frame)
	statusBar := RenderStatusBar(m.pipeline, m.done)

	var message string
	if m.done {
		if m.err != nil {
			msgStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorRed)).
				MarginTop(1)
			message = msgStyle.Render("Pipeline failed.")
		} else {
			msgStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorGreen)).
				MarginTop(1)
			message = msgStyle.Render("Pipeline complete.")
		}
	}

	parts := []string{title, flow}
	if message != "" {
		parts = append(parts, message)
	}
	parts = append(parts, statusBar)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return tea.NewView(content + "\n")
}
```

**Step 6: Build and run demo**

Run: `make demo`
Expected: Status bar visible at bottom showing `0/5 stages  q quit  tab select`, updating as stages complete.

**Step 7: Commit**

```bash
git add internal/ui/statusbar.go internal/ui/statusbar_test.go internal/ui/model.go
git commit -m "feat: add status bar with progress counter and key hints"
```

---

## Task 6: Ring buffer for output capture

**Why:** The inspector panel needs to show the last N lines of data flowing through each stage. We need a thread-safe ring buffer that the countingWriter populates and the UI reads from.

**Files:**
- Create: `internal/pipeline/ringbuffer.go`
- Create: `internal/pipeline/ringbuffer_test.go`
- Modify: `internal/pipeline/runner.go` (wire ring buffer into countingWriter)
- Modify: `internal/pipeline/pipeline.go` (add RingBuffer to StageStats)

**Step 1: Write ring buffer tests**

Create `internal/pipeline/ringbuffer_test.go`:

```go
package pipeline

import "testing"

func TestRingBufferBasic(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Write("line 1")
	rb.Write("line 2")
	rb.Write("line 3")

	lines := rb.Lines()
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "line 1" || lines[2] != "line 3" {
		t.Errorf("wrong order: %v", lines)
	}
}

func TestRingBufferOverwrite(t *testing.T) {
	rb := NewRingBuffer(3)
	rb.Write("a")
	rb.Write("b")
	rb.Write("c")
	rb.Write("d") // should overwrite "a"

	lines := rb.Lines()
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != "b" || lines[1] != "c" || lines[2] != "d" {
		t.Errorf("expected [b c d], got %v", lines)
	}
}

func TestRingBufferEmpty(t *testing.T) {
	rb := NewRingBuffer(5)
	lines := rb.Lines()
	if len(lines) != 0 {
		t.Errorf("got %d lines, want 0", len(lines))
	}
}

func TestRingBufferPartialFill(t *testing.T) {
	rb := NewRingBuffer(10)
	rb.Write("only one")
	lines := rb.Lines()
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}
	if lines[0] != "only one" {
		t.Errorf("got %q, want %q", lines[0], "only one")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pipeline/ -run TestRingBuffer -v`
Expected: FAIL — file doesn't exist

**Step 3: Implement ring buffer**

Create `internal/pipeline/ringbuffer.go`:

```go
package pipeline

import "sync"

// RingBuffer is a thread-safe circular buffer that stores the last N lines.
type RingBuffer struct {
	mu    sync.Mutex
	lines []string
	size  int
	pos   int
	count int
}

// NewRingBuffer creates a ring buffer that holds at most n lines.
func NewRingBuffer(n int) *RingBuffer {
	return &RingBuffer{
		lines: make([]string, n),
		size:  n,
	}
}

// Write adds a line to the buffer, overwriting the oldest if full.
func (rb *RingBuffer) Write(line string) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.lines[rb.pos] = line
	rb.pos = (rb.pos + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// Lines returns the buffered lines in chronological order.
func (rb *RingBuffer) Lines() []string {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.count == 0 {
		return nil
	}
	result := make([]string, rb.count)
	start := rb.pos - rb.count
	if start < 0 {
		start += rb.size
	}
	for i := 0; i < rb.count; i++ {
		result[i] = rb.lines[(start+i)%rb.size]
	}
	return result
}
```

**Step 4: Run ring buffer tests**

Run: `go test ./internal/pipeline/ -run TestRingBuffer -v`
Expected: All 4 PASS

**Step 5: Add RingBuffer to StageStats and wire into countingWriter**

In `pipeline.go`, add to StageStats:

```go
type StageStats struct {
	BytesIn    int64
	BytesOut   int64
	LinesIn    int64
	LinesOut   int64
	Throughput float64
	StartedAt  time.Time
	Duration   time.Duration
	Output     *RingBuffer // last N lines of stage output
}
```

In `runner.go`, update the countingWriter to also capture lines:

```go
type countingWriter struct {
	w       io.Writer
	stats   *StageStats
	lineBuf []byte // accumulates partial lines
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	if n > 0 && cw.stats != nil {
		cw.stats.AddBytesOut(int64(n))
		newLines := int64(bytes.Count(p[:n], []byte{'\n'}))
		cw.stats.AddLinesOut(newLines)

		// Capture lines for the ring buffer.
		if cw.stats.Output != nil {
			cw.lineBuf = append(cw.lineBuf, p[:n]...)
			for {
				idx := bytes.IndexByte(cw.lineBuf, '\n')
				if idx < 0 {
					break
				}
				line := string(cw.lineBuf[:idx])
				cw.stats.Output.Write(line)
				cw.lineBuf = cw.lineBuf[idx+1:]
			}
		}
	}
	return n, err
}
```

When creating StageStats in the runner, initialize the RingBuffer:

```go
stage.Stats = &StageStats{
	StartedAt: time.Now(),
	Output:    NewRingBuffer(100),
}
```

**Step 6: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 7: Commit**

```bash
git add internal/pipeline/ringbuffer.go internal/pipeline/ringbuffer_test.go internal/pipeline/pipeline.go internal/pipeline/runner.go
git commit -m "feat: add ring buffer for capturing last N lines of stage output"
```

---

## Task 7: Inspector panel

**Why:** Users should be able to select a stage and see data flowing through it in real-time. This is the interactive "wow" feature that differentiates pipe.dev from just watching text scroll by.

**Files:**
- Create: `internal/ui/inspector.go`
- Create: `internal/ui/inspector_test.go`
- Modify: `internal/ui/model.go` (add stage selection + render inspector)
- Modify: `internal/ui/node.go` (highlight selected node)

**Step 1: Write inspector tests**

Create `internal/ui/inspector_test.go`:

```go
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
	if !strings.Contains(result, "waiting") && !strings.Contains(result, "no output") {
		// Should show some placeholder when empty
		if result == "" {
			t.Error("inspector should render something even with no output")
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestInspector -v`
Expected: FAIL — `RenderInspector` doesn't exist

**Step 3: Implement inspector panel**

Create `internal/ui/inspector.go`:

```go
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
```

**Step 4: Run inspector tests**

Run: `go test ./internal/ui/ -run TestInspector -v`
Expected: PASS

**Step 5: Add stage selection to model.go**

Add `selected int` field to Model and handle Tab/arrow keys:

```go
type Model struct {
	pipeline *pipeline.Pipeline
	runner   *pipeline.Runner
	done     bool
	err      error
	frame    int
	selected int // index of selected stage (-1 = none)
}

func NewModel(p *pipeline.Pipeline) Model {
	return Model{
		pipeline: p,
		runner:   pipeline.NewRunner(p),
		selected: -1,
	}
}
```

In `Update()`, add key handling:

```go
case tea.KeyPressMsg:
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		m.selected++
		if m.selected >= len(m.pipeline.Stages) {
			m.selected = -1
		}
	case "shift+tab":
		m.selected--
		if m.selected < -1 {
			m.selected = len(m.pipeline.Stages) - 1
		}
	case "esc":
		m.selected = -1
	}
```

In `View()`, add inspector rendering after the flow:

```go
parts := []string{title, flow}
if m.selected >= 0 && m.selected < len(m.pipeline.Stages) {
	inspector := RenderInspector(m.pipeline.Stages[m.selected], 60)
	parts = append(parts, inspector)
}
if message != "" {
	parts = append(parts, message)
}
parts = append(parts, statusBar)
```

**Step 6: Highlight selected node in node.go**

Add a `selected` parameter to `RenderNode`:

```go
func RenderNode(stage *pipeline.Stage, selected bool) string {
```

At the end of the function, if selected, add an extra bold border or use a brighter color:

```go
box := lipgloss.NewStyle().
	Width(nodeWidth).
	Padding(0, 1).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(borderColor)

if selected {
	box = box.BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ColorMauve))
}
```

Update `flow.go` to pass selection state:

```go
func RenderFlow(p *pipeline.Pipeline, frame int, selected int) string {
	// ...
	for i, stage := range p.Stages {
		node := RenderNode(stage, i == selected)
		// ...
	}
}
```

Update `model.go` `View()` call:

```go
flow := RenderFlow(m.pipeline, m.frame, m.selected)
```

**Step 7: Build and run demo**

Run: `make demo`
Expected: Press Tab to cycle through stages. Selected stage shows a double-border in mauve. Inspector panel appears below the flow showing live output data.

**Step 8: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

**Step 9: Commit**

```bash
git add internal/ui/inspector.go internal/ui/inspector_test.go internal/ui/model.go internal/ui/node.go internal/ui/flow.go
git commit -m "feat: add inspector panel with Tab stage selection and live output preview"
```

---

## Task 8: Final integration test

**Why:** Verify everything works together end-to-end before moving to README/release tasks.

**Files:**
- Test: manual `make demo` + visual verification

**Step 1: Run full test suite**

Run: `go test ./... -v -count=1`
Expected: All tests pass (original 21 + new tests)

**Step 2: Run demo and visually verify**

Run: `make demo`

Verify:
- [ ] Animated `>>>` particles flow between running stages
- [ ] Particles stop (static arrows) when stage completes
- [ ] Live byte/line counters update during execution
- [ ] Status bar shows `N/5 stages` progress
- [ ] Tab cycles through stages
- [ ] Selected stage has double border in mauve
- [ ] Inspector panel shows actual output data
- [ ] Pipeline completes cleanly, shows "Pipeline complete."
- [ ] `q` quits at any point

**Step 3: Run with a real-world pipeline**

Run: `make build && ./pipe "find . -name '*.go' | xargs wc -l | sort -n | tail -5"`

Verify it works with a longer-running pipeline and displays meaningful data.

**Step 4: Commit any fixes needed**

If any issues found, fix and commit individually.

---

## Summary

| Task | What | Files Changed | New Tests |
|------|------|---------------|-----------|
| 1 | Fix pipe leak bug | runner.go, runner_test.go | 1 |
| 2 | Real-time atomic stats | pipeline.go, runner.go, node.go, runner_test.go | 1 |
| 3 | Tick animation loop | model.go | 0 (visual) |
| 4 | Animated connectors | connector.go, flow.go, model.go | 2 |
| 5 | Status bar | statusbar.go, model.go | 2 |
| 6 | Ring buffer | ringbuffer.go, pipeline.go, runner.go | 4 |
| 7 | Inspector panel | inspector.go, node.go, flow.go, model.go | 2 |
| 8 | Integration test | none | manual |

**Total: ~8 commits, ~12 new tests, 7 new/modified files**

After this plan is complete, the TUI will be demo-GIF-ready. The next plan should cover README, VHS recording, goreleaser, and CI setup.
