package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mathew-builds/pipe-dev/internal/adapter/unix"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
	"github.com/mathew-builds/pipe-dev/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "demo":
		fmt.Println("pipe.dev demo — coming soon")
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: pipe run <pipeline.yaml>")
			os.Exit(1)
		}
		fmt.Printf("pipe.dev run %s — coming soon\n", os.Args[2])
	case "--version", "-v":
		fmt.Printf("pipe %s\n", version.Version)
	case "--help", "-h":
		printUsage()
	default:
		// Treat as a Unix pipe command string
		adapter := &unix.Adapter{}
		p, err := adapter.Parse(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		runPipeline(p)
	}
}

func runPipeline(p *pipeline.Pipeline) {
	fmt.Printf("Pipeline: %s (%d stages)\n\n", p.Name, len(p.Stages))

	// Print stage layout.
	for i, stage := range p.Stages {
		arrow := "→"
		if i == 0 {
			arrow = "●"
		}
		cmd := stage.Command
		if len(stage.Args) > 0 {
			cmd += " " + strings.Join(stage.Args, " ")
		}
		fmt.Printf("  %s [%d] %s\n", arrow, i, cmd)
	}
	fmt.Println()

	// Execute pipeline and stream events.
	runner := pipeline.NewRunner(p)
	go runner.Run()

	for event := range runner.Events {
		switch event.Type {
		case pipeline.EventStageStarted:
			stage := findStage(p, event.StageID)
			fmt.Printf("  ▶ %s started\n", stage.Name)

		case pipeline.EventStageDone:
			stage := findStage(p, event.StageID)
			stats := stage.Stats
			fmt.Printf("  ✓ %s done — %s, %d bytes, %d lines\n",
				stage.Name,
				formatDuration(stats.Duration),
				stats.BytesOut,
				stats.LinesOut,
			)

		case pipeline.EventStageFailed:
			stage := findStage(p, event.StageID)
			stderr := ""
			if len(event.Output) > 0 {
				stderr = " — " + strings.TrimSpace(string(event.Output))
			}
			fmt.Fprintf(os.Stderr, "  ✗ %s failed: %v%s\n", stage.Name, event.Err, stderr)

		case pipeline.EventPipelineDone:
			fmt.Println()
			fmt.Println("Pipeline complete.")
		}
	}
}

func findStage(p *pipeline.Pipeline, id string) *pipeline.Stage {
	for _, s := range p.Stages {
		if s.ID == id {
			return s
		}
	}
	return &pipeline.Stage{Name: "unknown"}
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func printUsage() {
	fmt.Print(`pipe.dev — see data flow through your terminal pipelines in real-time

Usage:
  pipe "<cmd1> | <cmd2> | <cmd3>"   Visualize a Unix pipe chain
  pipe run <pipeline.yaml>          Run and visualize a YAML pipeline
  pipe demo                         Built-in demo (no setup needed)

Options:
  --help, -h       Show this help
  --version, -v    Show version

Examples:
  pipe "cat data.json | jq '.[]' | sort | uniq -c"
  pipe run etl-pipeline.yaml
  pipe demo

Learn more: https://github.com/mathew-builds/pipe-dev
`)
}
