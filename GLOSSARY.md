# Glossary

_Domain terms, abbreviations, and concepts used in this project._

---

## G

**gomux** — Third-party Go library (`github.com/wricardo/gomux`) that writes tmux shell commands to an `io.Writer`. Output is bash, not direct tmux API calls.

## J

**json2tmux** — This tool. Reads a JSON session descriptor from stdin, outputs bash commands that create a tmux session.

## P

**Pane** — Recursive struct representing a tmux pane. Holds `Command`, `Directory`, `SplitType`, and child `Split []*Pane`. Splits are created depth-first.

## S

**Session** — Top-level struct: `Name`, `Directory`, list of `Windows`. On `CreateSession`, kills any existing session with the same name then recreates it.

**SplitType** — Field on `Pane`. `"horizontal"` → horizontal split (`SplitWAttr`). Anything else → vertical split (`VsplitWAttr`).

## W

**Window** — Struct holding a name, optional directory, and one root `Pane`. Maps 1:1 to a tmux window.
