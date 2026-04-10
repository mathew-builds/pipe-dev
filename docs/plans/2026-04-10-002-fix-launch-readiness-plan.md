---
title: "feat: Launch readiness — README, demo GIF, CI, release infrastructure"
type: feat
status: active
date: 2026-04-10
origin: product-completeness-checker audit
---

# feat: Launch readiness — README, demo GIF, CI, release infrastructure

## Overview

The pipe.dev codebase is 100% complete for MVP architecture. Every planned component is built, tested, and working. The gap is entirely in launch infrastructure: no README, no demo GIF, no CI, no release binaries. This plan closes that gap for the April 22-23 launch target.

## Problem Frame

The product completeness audit found 28/35 vision items complete (80%). All 7 missing items are launch infrastructure — the packaging layer that makes an open-source project discoverable, installable, and shareable. Without these, the 500-star goal is unreachable regardless of code quality.

## Requirements Trace

- R1. README.md with personal tone, hero GIF, install instructions, examples
- R2. Demo GIF recorded with VHS showing animated TUI in action
- R3. GitHub Actions CI for test + lint on push/PR
- R4. Goreleaser config for cross-platform binary builds
- R5. First tagged release (v0.1.0) with binaries and `go install` path
- R6. Community files (CONTRIBUTING.md)
- R7. Clean up empty scaffold directories and stale files

## Scope Boundaries

- No new features — this is launch packaging only
- No terminal width responsiveness (post-launch)
- No stderr display in inspector (post-launch)
- No HN/Reddit/Twitter post drafting (separate task)
- No Homebrew tap in v0.1 (goreleaser can add this later)

## Key Technical Decisions

- **VHS for demo GIF:** Charm's VHS produces deterministic, reproducible terminal recordings as GIFs. It's the standard in the Bubbletea ecosystem and produces high-quality output.
- **README structure:** Hero GIF above the fold, one-command install, 3 copy-pasteable examples (one per command), "Why I built this" section. Personal voice per CLAUDE.md instructions.
- **goreleaser for releases:** Standard Go release tool. Builds for linux/darwin amd64/arm64, creates GitHub Release with checksums.
- **Minimal CI:** Single workflow with `go test ./...` and `golangci-lint run`. No coverage badges in v0.1.

## Implementation Units

- [ ] **Unit 1: Clean up scaffolding**

  **Goal:** Remove empty directories and update stale files.

  **Requirements:** R7

  **Files:**
  - Remove: `internal/app/` (empty scaffold)
  - Remove: `internal/tui/` (empty scaffold)
  - Modify: `NEXT_SESSION_PROMPT.md` — mark as stale or remove (its purpose is served)

  **Test expectation:** none — cleanup only

  **Verification:** `find . -type d -empty` returns no project directories. `go build ./...` still works.

- [ ] **Unit 2: GitHub Actions CI**

  **Goal:** Automated test and lint on every push and PR.

  **Requirements:** R3

  **Files:**
  - Create: `.github/workflows/ci.yml`

  **Approach:**
  - Trigger on push to main and pull_request
  - Matrix: Go 1.25, ubuntu-latest + macos-latest
  - Steps: checkout, setup-go, `go test ./...`, `golangci-lint run`
  - Use `golangci/golangci-lint-action`

  **Patterns to follow:** Standard Go CI patterns from Charm ecosystem repos

  **Test scenarios:**
  - Happy path: CI passes on current codebase (all tests pass, lint is clean)
  - Edge case: CI catches a broken PR (test failure blocks merge)

  **Verification:** Push to GitHub, CI runs green.

- [ ] **Unit 3: Goreleaser config**

  **Goal:** Automated cross-platform binary builds on git tag.

  **Requirements:** R4, R5

  **Files:**
  - Create: `.goreleaser.yml`
  - Create: `.github/workflows/release.yml`

  **Approach:**
  - Build targets: linux/darwin, amd64/arm64
  - Binary name: `pipe`
  - Ldflags: `-X github.com/mathew-builds/pipe-dev/pkg/version.Version={{.Version}}`
  - Release workflow triggers on tag push (`v*`)
  - Uses `goreleaser/goreleaser-action`

  **Test scenarios:**
  - Happy path: `goreleaser check` validates the config
  - Happy path: `goreleaser build --snapshot` produces binaries locally

  **Verification:** Config validates. Snapshot build produces 4 binaries (linux/darwin x amd64/arm64).

- [ ] **Unit 4: Record demo GIF with VHS**

  **Goal:** A compelling 10-15 second GIF showing pipe.dev in action.

  **Requirements:** R2

  **Files:**
  - Create: `demo.tape` (VHS tape file)
  - Output: `assets/demo.gif`

  **Approach:**
  - VHS tape that runs `pipe demo`, waits for animation, presses Tab to show inspector, then q to quit
  - Settings: width 120, height 30, font size appropriate for GIF readability
  - The GIF must show: animated particles flowing, live byte counters updating, inspector panel with data, clean completion with all ✓

  **Test scenarios:**
  - Happy path: `vhs demo.tape` produces a GIF under 5MB that shows the full demo flow

  **Verification:** GIF exists, looks good at README width, shows animated connectors and inspector.

- [ ] **Unit 5: Write README.md**

  **Goal:** A compelling README that converts GitHub visitors to stars.

  **Requirements:** R1

  **Files:**
  - Create: `README.md`

  **Approach:**
  - Hero GIF (from Unit 4) above the fold
  - One-liner: "See data flow through your terminal pipelines in real-time."
  - Install section: `go install`, binary download, build from source
  - Usage: one example per command (pipe string, pipe run, pipe demo)
  - "Why I built this" section — personal voice, not corporate
  - Feature highlights: animated connectors, live stats, inspector panel, SIGPIPE handling
  - Contributing section (brief)
  - License (MIT)

  **Patterns to follow:** Charm ecosystem READMEs (glow, mods, vhs) for tone and structure. But more personal per CLAUDE.md instructions.

  **Test scenarios:**
  - Happy path: README renders correctly on GitHub with GIF, install commands, examples

  **Verification:** Push to GitHub, README renders with GIF and all sections visible.

- [ ] **Unit 6: Community files**

  **Goal:** Basic contributor experience files.

  **Requirements:** R6

  **Files:**
  - Create: `CONTRIBUTING.md`

  **Approach:**
  - Brief, friendly contributing guide
  - How to build, test, lint
  - PR expectations
  - No CODE_OF_CONDUCT.md for v0.1 (can add later)

  **Test expectation:** none — documentation only

  **Verification:** File exists and renders on GitHub.

- [ ] **Unit 7: Tag and release v0.1.0**

  **Goal:** First official release with binaries.

  **Requirements:** R5

  **Dependencies:** Units 2, 3, 4, 5, 6

  **Approach:**
  - `git tag v0.1.0`
  - `git push origin v0.1.0`
  - Release workflow triggers goreleaser
  - Verify binaries are attached to GitHub Release
  - Verify `go install github.com/mathew-builds/pipe-dev/cmd/pipe@v0.1.0` works

  **Verification:** GitHub Release page shows v0.1.0 with 4 binaries. `go install` works.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| VHS may not be installed | `go install github.com/charmbracelet/vhs@latest` or `brew install vhs` |
| GIF may be too large for GitHub | Target < 5MB. VHS supports quality/fps settings. |
| goreleaser config may need iteration | `goreleaser check` validates before tagging. Snapshot builds test locally. |
| README tone may feel forced | Review against lazygit, glow, and bat READMEs for calibration. Personal ≠ quirky. |

## Sources & References

- Product completeness audit (this session)
- Architecture.md: `~/Open_Source_Lab/01_Projects/pipe-dev/Architecture.md`
- Feedback: `~/Open_Source_Lab/01_Projects/pipe-dev/feedback/2026-04-09_mvp_review.md`
- VHS: github.com/charmbracelet/vhs
- Goreleaser: goreleaser.com
