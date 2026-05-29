# FAQ

## What does json2tmux do?

Reads a JSON session descriptor (file arg, `-f` flag, or stdin), outputs bash commands that create a tmux session. Pipe the output to `bash` to execute.

## How do I run it?

```bash
json2tmux examples/example1.json | bash   # file arg (simplest)
json2tmux -f examples/example1.json | bash
json2tmux -attach examples/example1.json | bash  # launch + attach
cat examples/example1.json | json2tmux | bash     # stdin
tmux attach -t Example1
```

## What flags are available?

| Flag | Description |
|------|-------------|
| `<file>` | Positional file argument — reads JSON from file |
| `-f <file>` | Explicit file flag — same as positional arg |
| `-attach` | Appends `tmux attach -t <name>` to output |
| `-dry-run` | Accepted flag; output is always bash commands |
| `-h`, `--help` | Print usage and full schema reference |

## What does the JSON structure look like?

```json
{
  "Name": "MySession",
  "Directory": "/home/user/project",
  "Windows": [
    {
      "Name": "editor",
      "Directory": "/home/user/project",
      "Pane": {
        "Command": "vim .",
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

## What split types are supported?

- `"horizontal"` — splits pane left/right
- `"vertical"` (or anything else) — splits pane top/bottom (default)

## What happens if a session with the same name already exists?

It gets killed and recreated. No prompt, no merge.

## Does it run commands directly or output shell commands?

Outputs bash commands only. Nothing executes until you pipe to `bash`.

## Can panes be nested?

Yes. `Split` is recursive — each pane can have its own `Split` array. Splits are created depth-first.

## Does Directory on a Window override the Session Directory?

Yes. Window `Directory` is passed to the tmux window. If omitted, tmux inherits the session directory.

## Where does the pane Directory apply?

`Directory` on a `Pane` is passed to the split attributes (`SplitAttr.Directory`) for child panes, not the root pane of the window.
