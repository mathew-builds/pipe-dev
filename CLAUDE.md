# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

**pipe.dev** — Real-time visualization of data flowing through terminal pipelines. "lazygit for data flows."

Go + Bubbletea v2 + Lipgloss v2, Catppuccin Mocha theme. Module: `github.com/mathew-builds/pipe-dev`.

## Build & Test Commands

```bash
make build              # Build binary with version from git tags
make test               # Run all tests (go test ./... -v)
make lint               # golangci-lint run
make demo               # Build + run built-in demo
make clean              # Remove binary

# Run a single test:
go test ./internal/pipeline -v -run TestRunTwoStages
```

## Architecture

Three layers: **Adapters** parse input into a Pipeline, the **Runner** executes it, the **UI** visualizes it.

```
Adapter.Parse(input) → Pipeline → Runner.Run() → Events channel → Bubbletea UI
```

### Adapters (`internal/adapter/`)

Interface: `Parse(input string) (*pipeline.Pipeline, error)`. Two implementations:
- `unix/` — splits shell pipe string on `|`, tokenizes with `strings.Fields()` (no shell quote parsing)
- `yaml/` — reads YAML file with `name` + `stages` array

### Runner (`internal/pipeline/runner.go`)

Starts all stages concurrently (like real shell pipes), chained via `io.Pipe`. Each stage's stdout is wrapped in a `countingWriter` that:
1. Atomically updates `StageStats.BytesOut` / `LinesOut` (safe for concurrent UI reads)
2. Captures complete lines into a `RingBuffer` (last 100 lines, for the inspector panel)
3. Forwards bytes to the next stage's stdin (or discards for final stage)

Emits events on a channel: `EventStageStarted` → `EventStageDone`/`EventStageFailed` → `EventPipelineDone`.

**SIGPIPE handling:** Non-final stages receiving SIGPIPE (e.g., downstream `head` exits early) are treated as successful completion — this matches Unix semantics.

### UI (`internal/ui/`)

Bubbletea model with a 100ms tick loop driving animation. Reads events via a **blocking `tea.Cmd`** that directly reads from `runner.Events` channel (not `tea.Sub`).

- `flow.go` — horizontal stage layout with animated connectors
- `connector.go` — 6-frame particle animation between stages
- `node.go` — stage box with name, status icon, live stats
- `inspector.go` — Tab-selectable panel showing stage output from ring buffer
- `theme.go` — Catppuccin Mocha palette (all colors defined here)

### Domain Model (`internal/pipeline/pipeline.go`)

`Pipeline` → `[]Stage` → `StageStats`. Stats use `atomic.AddInt64`/`atomic.LoadInt64` for lock-free concurrent access between runner (writer) and UI (reader).

## Testing Conventions

- Table-driven tests with `name`/`input`/`expected` structs
- Integration-style: tests run real commands (`echo`, `grep`, `wc`)
- No assertion libraries — direct `if != { t.Errorf }` comparisons
- Helper functions: `buildPipeline()`, `collectEvents()` to reduce boilerplate

## Dependencies

- Go 1.25+
- `charm.land/bubbletea/v2` — TUI framework
- `charm.land/lipgloss/v2` — Styling
- `gopkg.in/yaml.v3` — YAML parsing

## Constraints

- Do NOT add adapters beyond unix and yaml in v0.1
- Do NOT add a plugin loader yet
- Do NOT use corporate voice in README — personal "I built this" tone
- Do NOT over-engineer — ship fast, polish later
