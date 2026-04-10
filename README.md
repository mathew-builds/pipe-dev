# pipe.dev

**See data flow through your terminal pipelines in real-time.**

![demo](assets/demo.gif)

I built pipe.dev because I kept running long shell pipelines and staring at a blank terminal wondering which stage was slow, how much data had flowed through, and whether anything was actually happening. Now I can see it.

## Install

```bash
# Go
go install github.com/mathew-builds/pipe-dev/cmd/pipe@latest

# From source
git clone https://github.com/mathew-builds/pipe-dev.git
cd pipe-dev && make build
```

Or grab a binary from [Releases](https://github.com/mathew-builds/pipe-dev/releases).

## Usage

### Visualize any pipe chain

```bash
pipe "cat access.log | grep 500 | sort | uniq -c | sort -rn"
```

### Run a YAML pipeline

```yaml
# etl.yaml
name: ETL Pipeline
stages:
  - name: extract
    command: cat data.json
  - name: transform
    command: jq '.[] | select(.active)'
  - name: load
    command: wc -l
```

```bash
pipe run etl.yaml
```

### Try the built-in demo

```bash
pipe demo
```

## What you get

- **Animated flow visualization** — particles flow between stages so you can see the pipeline is alive
- **Live byte and line counters** — watch data volume update in real-time as stages process
- **Inspector panel** — press `Tab` to cycle through stages and see their last 100 lines of output
- **Smart SIGPIPE handling** — when `head` exits early, upstream stages complete cleanly (like a real shell)

## Keys

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle through stages |
| `Esc` | Deselect stage |
| `q` / `Ctrl+C` | Quit |

## Why I built this

I work with data pipelines daily. Every time I chain together `grep | sort | uniq | head`, I'm flying blind — no idea which stage is the bottleneck, how much data is flowing, or if the pipeline is even making progress. The terminal just sits there until it's done (or isn't).

pipe.dev is my answer to that. It's [lazygit](https://github.com/jesseduffield/lazygit) for data flows — a TUI that makes the invisible visible.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
