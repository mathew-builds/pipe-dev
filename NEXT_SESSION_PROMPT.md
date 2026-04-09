## Session Prompt — pipe.dev

### What Was Done (2026-04-09)
MVP is fully functional. All three commands work:
- `pipe "cmd1 | cmd2 | cmd3"` — parse and execute Unix pipes with TUI
- `pipe run pipeline.yaml` — YAML pipeline execution with TUI
- `pipe demo` — built-in 5-stage showcase pipeline

8 commits from scaffolding through full MVP. 21 tests across 3 packages, all passing.

### Architecture Implemented
```
cmd/pipe/main.go           → CLI entry + TUI launch
internal/pipeline/          → Domain model, runner (TeeReader), events
internal/adapter/unix/      → Unix pipe string parser
internal/adapter/yaml/      → YAML file parser
internal/ui/                → Bubbletea v2 TUI (model, node, connector, flow, theme)
pkg/version/                → Build-time version
examples/                   → Sample YAML pipelines
```

### Key Technical Decisions
- bubbletea v2 import path is `charm.land/bubbletea/v2` (not `github.com/charmbracelet`)
- lipgloss v2 import path is `charm.land/lipgloss/v2`
- exec.Command bypasses shell — no quotes needed in command strings
- Runner uses io.Pipe + countingWriter between stages for byte/line tracking
- Events emitted on channel, TUI reads via tea.Cmd

### What's Next (Post-MVP Polish)
- README.md — personal "I built this" tone, GIF/screenshot of TUI
- CI setup (GitHub Actions for tests + lint)
- Error UX improvements (e.g., graceful non-TTY fallback)
- Inspector panel (select a stage node to see detailed stats)
- Release workflow (goreleaser, homebrew tap)
- Real-world testing with larger/longer-running pipelines

### CLAUDE.md Updates Needed
- Dependencies section: update import paths to charm.land
- Add `examples/` to architecture diagram
- Consider adding a "Status: MVP complete" field

### User Feedback
- User will provide feedback in next session on overall approach and TUI quality
