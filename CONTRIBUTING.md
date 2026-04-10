# Contributing to pipe.dev

Thanks for your interest! Here's how to get started.

## Setup

```bash
git clone https://github.com/mathew-builds/pipe-dev.git
cd pipe-dev
make build
make test
```

Requires Go 1.25+ and [golangci-lint](https://golangci-lint.run/).

## Making changes

1. Fork and create a branch from `main`
2. Make your changes
3. Run `make test` and `make lint`
4. Open a PR with a clear description of what and why

## Code style

- Follow existing patterns in the codebase
- Tests are table-driven with real commands (no mocks)
- No assertion libraries — use `if != { t.Errorf }` style

## What's in scope for v0.1

Bug fixes and polish. New adapters and features are planned for future releases — check [Issues](https://github.com/mathew-builds/pipe-dev/issues) before starting large work.
