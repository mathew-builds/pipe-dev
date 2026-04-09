package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/mathew-builds/pipe-dev/internal/adapter/unix"
	"github.com/mathew-builds/pipe-dev/internal/pipeline"
	"github.com/mathew-builds/pipe-dev/internal/ui"
	"github.com/mathew-builds/pipe-dev/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "demo":
		runDemo()
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

func runDemo() {
	// A showcase pipeline using universally available commands.
	// Generates 1000 numbers, filters, sorts, deduplicates, and counts.
	adapter := &unix.Adapter{}
	p, err := adapter.Parse("seq 1 1000 | grep 7 | sort -r | head -20 | wc -l")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	p.Name = "Demo Pipeline"
	runPipeline(p)
}

func runPipeline(p *pipeline.Pipeline) {
	m := ui.NewModel(p)
	prog := tea.NewProgram(m)
	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
