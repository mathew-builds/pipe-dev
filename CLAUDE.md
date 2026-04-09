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
cmd/pipe/main.go           → CLI entry point
internal/pipeline/          → Domain model (Pipeline, Stage, StageStats)
internal/pipeline/runner.go → Executes stages, TeeReader interception
internal/adapter/           → Adapter interface + unix/yaml implementations
internal/ui/                → Bubbletea TUI components (flow, node, connector, inspector)
pkg/version/                → Build-time version
```

### Key Patterns
- **Adapter interface:** `Parse(input) -> Pipeline`. UI is source-agnostic.
- **TeeReader interception:** stdout piped through TeeReader for byte/line monitoring.
- **Bubbletea messages:** Runner emits StageStarted/Output/Done/Failed. UI subscribes via tea.Sub.

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

- Go 1.23+
- github.com/charmbracelet/bubbletea/v2
- github.com/charmbracelet/lipgloss/v2
- github.com/charmbracelet/bubbles/v2

## What NOT to Do

- Do NOT add adapters beyond unix and yaml in v0.1
- Do NOT add a plugin loader yet (architecture supports it, but not for MVP)
- Do NOT use corporate voice in README — personal "I built this" tone
- Do NOT over-engineer — ship fast, polish later
