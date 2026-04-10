# CLAUDE.md — pipe.dev

**pipe.dev** — See data flow through your terminal pipelines in real-time.
"lazygit for data flows."

## Quick Context

- **Author:** Mathew Joseph (github.com/mathew-builds)
- **Tech:** Go + Bubbletea v2 + Lipgloss (Catppuccin Mocha theme)
- **Goal:** GitHub Trending, 500+ stars in 90 days, GTV visa evidence
- **Planning docs:** ~/Open_Source_Lab/01_Projects/pipe-dev/

## Architecture

```
cmd/pipe/main.go                → CLI entry point
internal/pipeline/pipeline.go   → Domain model (Pipeline, Stage, StageStats)
internal/pipeline/runner.go     → Executes stages, countingWriter interception
internal/pipeline/event.go      → Event types (StageStarted/Done/Failed/PipelineDone)
internal/pipeline/ringbuffer.go → Thread-safe circular buffer for output capture
internal/adapter/adapter.go     → Adapter interface
internal/adapter/unix/unix.go   → Unix pipe string parser
internal/adapter/yaml/yaml.go   → YAML pipeline file parser
internal/ui/model.go            → Main Bubbletea model (tick loop, event handling)
internal/ui/flow.go             → Flow visualization (nodes + connectors)
internal/ui/node.go             → Stage node rendering
internal/ui/connector.go        → Animated flowing particles between stages
internal/ui/inspector.go        → Data preview panel (Tab to select stage)
internal/ui/statusbar.go        → Progress counter + key hints
internal/ui/helpers.go          → Formatting utilities (bytes, lines, duration)
internal/ui/theme.go            → Catppuccin Mocha color palette
pkg/version/                    → Build-time version
```

### Key Patterns
- **Adapter interface:** `Parse(input) -> Pipeline`. UI is source-agnostic.
- **countingWriter interception:** stdout piped through io.Pipe with countingWriter for byte/line monitoring and ring buffer capture.
- **Bubbletea messages:** Runner emits StageStarted/Done/Failed/PipelineDone. UI reads events via blocking tea.Cmd on runner channel.
- **SIGPIPE handling:** Non-final stages receiving SIGPIPE (from downstream early exit) are treated as successful completion.

## MVP Commands

```bash
pipe "cmd1 | cmd2 | cmd3"    # Visualize Unix pipes
pipe run pipeline.yaml        # Run YAML pipeline
pipe demo                     # Built-in demo showcase
```

## Build & Run

```bash
make build    # Build binary
make demo     # Build and run demo
make test     # Run tests
make lint     # Run linter
```

## Dependencies

- Go 1.25+
- charm.land/bubbletea/v2
- charm.land/lipgloss/v2
- gopkg.in/yaml.v3

## What NOT to Do

- Do NOT add adapters beyond unix and yaml in v0.1
- Do NOT add a plugin loader yet (architecture supports it, but not for MVP)
- Do NOT use corporate voice in README — personal "I built this" tone
- Do NOT over-engineer — ship fast, polish later
