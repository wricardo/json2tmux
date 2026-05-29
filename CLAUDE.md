# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What it does

Reads a JSON session descriptor (from file arg, -f flag, or stdin), outputs bash commands that create a tmux session. Pipe output to `bash`.

```bash
json2tmux examples/example1.json | bash
tmux attach -t Example1
```

## Commands

```bash
go build .                             # build binary
json2tmux examples/example1.json | bash  # run end-to-end
go test ./...                          # run tests
```

## Architecture

Single `main.go`. Three types:

- `Session` — top-level: name, directory, list of windows. Calls `gomux.KillSession` then `gomux.NewSessionAttr`.
- `Window` — holds one root `Pane`, optional directory. Maps to a tmux window.
- `Pane` — recursive: `Command`, `Directory`, `SplitType` (`"horizontal"` or `"vertical"`), `Split []*Pane`. Splits are created depth-first via `SplitPane()`.

Depends on `github.com/wricardo/gomux` — a thin wrapper that writes tmux shell commands to an `io.Writer`. The output is bash, not direct tmux calls.

`SplitType` default (anything not `"horizontal"`) = vertical split.

## CLI flags

- Positional arg or `-f <file>`: read session JSON from file instead of stdin
- `-exec`: execute tmux commands directly via `bash -s` — no pipe needed
- `-attach`: append/run `tmux attach -t <name>`
- `-dry-run`: print bash commands for inspection
- `--help` / `-h`: print usage and schema reference

## Documentation rules

After every code change, update the relevant docs:

- `README.md` — usage, flags, examples, schema
- `CLAUDE.md` — commands, architecture, CLI flags
- `FAQ.md` — if behavior or common questions change
- `GLOSSARY.md` — if new terms are introduced
- `examples/` — add or update example JSON if new features affect session structure
- `main_test.go` — add tests for any new behavior
