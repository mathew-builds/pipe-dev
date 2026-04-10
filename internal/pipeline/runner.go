package pipeline

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"
)

// Runner executes a pipeline and emits events as stages progress.
type Runner struct {
	Pipeline *Pipeline
	Events   chan Event
}

// NewRunner creates a runner for the given pipeline.
func NewRunner(p *Pipeline) *Runner {
	return &Runner{
		Pipeline: p,
		Events:   make(chan Event, 64),
	}
}

// Run executes all pipeline stages, chaining stdout→stdin between them.
// It blocks until all stages complete, then closes the Events channel.
func (r *Runner) Run() error {
	stages := r.Pipeline.Stages
	if len(stages) == 0 {
		close(r.Events)
		return nil
	}

	cmds := make([]*exec.Cmd, len(stages))

	// Build commands and initialize stats early so countingWriters can reference them.
	for i, stage := range stages {
		cmds[i] = exec.Command(stage.Command, stage.Args...)
		stage.Stats = &StageStats{StartedAt: time.Now()}
	}

	// Chain stages: stage[i].stdout → countingWriter → stage[i+1].stdin
	for i := 0; i < len(cmds)-1; i++ {
		pr, pw := io.Pipe()

		// Stage i writes to a counting writer that forwards to the pipe.
		cmds[i].Stdout = &countingWriter{w: pw, stats: stages[i].Stats}

		// Stage i+1 reads from the pipe.
		cmds[i+1].Stdin = pr

		// Close the pipe writer when stage i's goroutine finishes
		// (deferred in the start goroutine below).
		_ = pw // captured in closure below
	}

	// Last stage writes to a counting writer backed by a discard buffer.
	lastBuf := &bytes.Buffer{}
	cmds[len(cmds)-1].Stdout = &countingWriter{w: lastBuf, stats: stages[len(cmds)-1].Stats}

	// Capture stderr for error reporting.
	stderrBufs := make([]*bytes.Buffer, len(cmds))
	for i := range cmds {
		stderrBufs[i] = &bytes.Buffer{}
		cmds[i].Stderr = stderrBufs[i]
	}

	// Start all stages concurrently (like a real shell pipe).
	var wg sync.WaitGroup
	for i := range cmds {
		stage := stages[i]
		cmd := cmds[i]
		stderrBuf := stderrBufs[i]

		stage.Status = StatusRunning
		stage.Stats.StartedAt = time.Now() // Reset to accurate start time
		r.emit(Event{Type: EventStageStarted, StageID: stage.ID})

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

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			err := cmd.Wait()
			stage.Stats.Duration = time.Since(stage.Stats.StartedAt)

			// Close the pipe writer so the next stage gets EOF.
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
	}

	wg.Wait()
	r.emit(Event{Type: EventPipelineDone})
	close(r.Events)
	return nil
}

func (r *Runner) emit(e Event) {
	r.Events <- e
}

// countingWriter wraps a writer and atomically updates StageStats in real-time.
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
