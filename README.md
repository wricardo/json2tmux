# json2tmux

Creates tmux sessions from JSON. Reads a session descriptor on stdin, outputs bash commands.

## Install

```bash
# macOS / Linux (one-liner)
curl -fsSL https://raw.githubusercontent.com/wricardo/json2tmux/main/install.sh | bash

# or with Go
go install github.com/wricardo/json2tmux@latest
```

Or build from source:

```bash
git clone https://github.com/wricardo/json2tmux
cd json2tmux
go build -o json2tmux .
```

## Usage

```bash
json2tmux session.json | bash          # positional file arg
json2tmux -f session.json | bash       # explicit -f flag
json2tmux -attach session.json | bash  # launch + attach in one step
cat session.json | json2tmux | bash    # stdin
```

Run `json2tmux --help` to print the full schema reference.

See [`examples/`](examples/) for ready-to-use session files.

## SSH Fleets

Open SSH connections to multiple servers in one shot. Each pane is its own connection — useful for monitoring a cluster, running parallel deploys, or keeping prod shells organized.

```json
{
  "Name": "fleet",
  "Windows": [
    {
      "Name": "web",
      "Pane": {
        "Command": "ssh web-01.prod.example.com",
        "SplitType": "horizontal",
        "Split": [{ "Command": "ssh web-02.prod.example.com" }]
      }
    },
    {
      "Name": "db",
      "Pane": {
        "Command": "ssh db-primary.prod.example.com",
        "SplitType": "horizontal",
        "Split": [{ "Command": "ssh db-replica.prod.example.com" }]
      }
    }
  ]
}
```

```bash
cat examples/ssh-fleet.json | json2tmux | bash
tmux attach -t fleet
```

See [`examples/ssh-fleet.json`](examples/ssh-fleet.json) for a 4-window example (web, db, workers, monitor).

### Remote server sessions

Bootstrap a session on a remote server. Copy the JSON, SSH in, and run json2tmux there:

```bash
# Copy and bootstrap
scp examples/remote-session.json user@server:~
ssh user@server 'cat remote-session.json | json2tmux | bash'

# Attach from local
ssh -t user@server 'tmux attach -t remote-dev'
```

The session definition lives in version control — everyone gets the same remote layout.

See [`examples/remote-session.json`](examples/remote-session.json) for an editor + logs + shell layout.

## Using with Claude Code

json2tmux pairs well with [Claude Code](https://claude.ai/code). Commit a session JSON to your repo and have Claude launch a reproducible dev environment — or instruct Claude to create and interact with panes directly.

### Reproducible dev environment

Commit a session file (e.g. `examples/claude-code.json`) and launch it:

```bash
cat examples/claude-code.json | json2tmux | bash
tmux attach -t dev
```

Anyone cloning the repo gets the exact same layout.

### Ask Claude to launch a session

With the [tmux MCP server](https://github.com/nickclyde/mcp-server-tmux) enabled in Claude Code, you can ask Claude to:

- **Launch your dev session**: `"Run examples/claude-code.json through json2tmux and attach"` — Claude pipes the file through json2tmux and executes the output.
- **Recreate after a reboot**: `"Restore my dev session"` — Claude re-runs the same JSON, kills the stale session, and rebuilds it identically.
- **Inspect running panes**: Claude uses `mcp__tmux__capture-pane` to read pane output and `mcp__tmux__execute-command` to send keystrokes — useful for checking server logs or test output without leaving the conversation.

### Ask Claude to generate a session file

Describe your layout and Claude will produce the JSON:

> "Create a json2tmux session for this Go project: vim on the left, `go run .` on the right, second window for git, third window idle shell."

> "Create an SSH fleet session for web-01 through web-04 and db-primary and db-replica, side by side."

For SSH fleets, give Claude your host naming convention and it generates the full layout.

### CLAUDE.md tip

Add this to your project `CLAUDE.md` so Claude always knows how to restore your environment:

```markdown
## Dev session

Restore with:
\`\`\`bash
cat examples/claude-code.json | json2tmux | bash && tmux attach -t dev
\`\`\`
```

## JSON Schema

```json
{
  "Name": "MySession",
  "Directory": "~/project",
  "Windows": [
    {
      "Name": "editor",
      "Directory": "~/project",
      "Pane": {
        "Command": "vim .",
        "Directory": "~/project",
        "SplitType": "vertical",
        "Split": [
          {
            "Command": "go run .",
            "SplitType": "horizontal",
            "Split": []
          }
        ]
      }
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | tmux session name. Existing session with this name is killed first. |
| `Directory` | string | Default working directory for the session. |
| `Windows[].Name` | string | tmux window name. |
| `Windows[].Directory` | string | Working directory for the window. |
| `Pane.Command` | string | Command to run in the pane. |
| `Pane.Directory` | string | Working directory passed to child splits. |
| `Pane.SplitType` | string | `"horizontal"` or `"vertical"` (default). |
| `Pane.Split` | array | Recursive child panes. Split depth-first. |
