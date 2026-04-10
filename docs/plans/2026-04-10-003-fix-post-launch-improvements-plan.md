---
title: "fix: Post-launch improvements — race conditions, stderr display, terminal width"
type: fix
status: backlog
date: 2026-04-10
origin: research findings from consolidation + flow analysis sessions
---

# fix: Post-launch improvements

Items discovered during the consolidation and completeness audit sessions that are valid improvements but not launch blockers. Track here so they don't get lost.

## Items

1. **Add `ws.Signaled()` check before `ws.Signal()` in SIGPIPE handler** — More defensive pattern. Currently safe because we're inside an `exec.ExitError` check, but `Signaled()` guard is the canonical Go pattern.

2. **Race condition on `Stage.Status` and `Stage.Error`** — These are written by runner goroutines and read by UI without synchronization. `BytesOut`/`LinesOut` use atomics but `Status` doesn't. Real concern on Apple Silicon (ARM memory model). Fix: make `Status` an `atomic.Int32` or add mutex.

3. **`grep` exit code 1 (no matches) treated as failure** — In pipeline semantics, `grep` returning 1 means "no matches found", not an error. Common pipelines like `pipe "cat file | grep pattern | wc -l"` show grep as failed. Fix: treat exit code 1 on non-final stages as success, or add a `StatusWarning` state.

4. **Inspector doesn't show stderr for failed stages** — When a stage fails, stderr is captured in `EventStageFailed.Output` but the inspector shows "waiting for output..." instead. Fix: store stderr on Stage struct, display in inspector with red styling.

5. **Terminal width overflow** — 5 nodes + 4 connectors = ~170 chars. Overflows 80-column terminals. Fix: detect terminal width via `tea.WindowSizeMsg`, adapt layout (compress connectors, wrap rows, or truncate).

6. **`signal.Reset(syscall.SIGPIPE)` in main()** — Prevents the Go process itself from dying ungracefully when pipe.dev's output is piped to another command that exits early. Standard Charm ecosystem pattern.

7. **Ring buffer doesn't flush partial final lines** — If a stage's last output doesn't end with `\n`, the content in `countingWriter.lineBuf` is never written to the ring buffer. Fix: flush on stage completion.

8. **`stageID()` test helper breaks for i >= 10** — Uses `string(rune('0'+i))` which produces non-digit chars for 10+. Not a current problem but fragile.

9. **Status bar doesn't reflect failed stages in count** — Shows "4/5 stages" when 1 failed, giving impression 1 is still running. Fix: "4/5 stages (1 failed)" or count failed as completed.

10. **Missing test coverage for `node.go`, `flow.go`, `model.go`, `helpers.go`** — 4 of 8 UI source files have no tests. Pure render functions in helpers.go and node.go are trivially testable.
