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
cat session.json | json2tmux | bash
tmux attach -t MySession
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

## Example

```bash
cat examples/example1.json | json2tmux | bash
tmux attach -t Example1
```

Run `json2tmux --help` to print schema reference.
