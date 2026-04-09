package main

import (
	"fmt"
	"os"

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
		fmt.Printf("pipe.dev visualizing: %s — coming soon\n", os.Args[1])
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
