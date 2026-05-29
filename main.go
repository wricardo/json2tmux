package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/wricardo/gomux"
)

const usage = `json2tmux — create tmux sessions from JSON

USAGE
  json2tmux [flags] [file]
  cat session.json | json2tmux [flags]
  Input priority: -f flag > positional arg > stdin

FLAGS
  -f <file>  Read session JSON from file
  -exec      Execute via bash directly (no pipe needed)
  -attach    Also run: tmux attach -t <Name>
  -dry-run   Print bash commands, do not execute
  -h         This help

SCHEMA
  {
    "Name":      string,     // session name; existing session killed+recreated
    "Directory": string,     // default working dir for session
    "Windows": [{
      "Name":      string,   // window name
      "Directory": string,   // override working dir for this window
      "Pane": <Pane>
    }]
  }

  Pane {
    "Command":   string,     // shell command to run in pane
    "Directory": string,     // cd here before child splits (quoted, handles spaces)
    "SplitType": string,     // "horizontal"=left/right  "vertical"=top/bottom (default)
    "Split":     [<Pane>]    // child panes; recursive, depth-first
  }

BEHAVIOR
  - Output is bash commands (tmux CLI calls); safe to inspect before running
  - Session kill+recreate is idempotent — run anytime to reset layout
  - Pane.Directory applies cd before each child split, not the root pane
  - Window.Directory sets the window working dir via tmux WindowAttr
  - SplitType default (any value except "horizontal") = vertical

EXAMPLES
  json2tmux dev.json | bash                   # pipe to bash
  json2tmux -exec dev.json                    # execute directly
  json2tmux -exec -attach dev.json            # execute + attach (one command)
  json2tmux -dry-run dev.json                 # inspect commands
  json2tmux -f dev.json | bash                # explicit -f flag
  cat dev.json | json2tmux -exec              # stdin + exec

SSH FLEET
  Set "Command" to an ssh invocation per pane:
  { "Command": "ssh user@host", "SplitType": "horizontal", "Split": [...] }

REMOTE BOOTSTRAP
  scp session.json user@host:~
  ssh user@host 'json2tmux -exec session.json'
  ssh -t user@host 'tmux attach -t <Name>'

CLAUDE CODE
  Ask Claude to generate a session file:
    "Create a json2tmux session: vim + go run side by side, second window idle shell"
  Ask Claude to launch it (requires tmux MCP server):
    "Run examples/dev.json through json2tmux -exec -attach"
  Add to CLAUDE.md for auto-restore:
    json2tmux -exec -attach examples/dev.json
`

func main() {
	var (
		fileFlag = flag.String("f", "", "session JSON file")
		attach   = flag.Bool("attach", false, "append or run 'tmux attach -t <name>'")
		execFlag = flag.Bool("exec", false, "execute tmux commands directly via bash")
		dryRun   = flag.Bool("dry-run", false, "print bash commands without executing")
	)

	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	// resolve input source: -f flag > positional arg > stdin
	var r io.Reader
	switch {
	case *fileFlag != "":
		f, err := os.Open(*fileFlag)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		defer f.Close()
		r = f
	case flag.NArg() > 0:
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		defer f.Close()
		r = f
	default:
		r = os.Stdin
	}

	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatalf("error reading input: %v", err)
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid JSON: %v\n\nRun with --help for usage.\n", err)
		os.Exit(1)
	}

	// always build the bash script into a buffer
	var buf bytes.Buffer
	s.CreateSession(&buf)
	if *attach {
		fmt.Fprintf(&buf, "tmux attach -t %q\n", s.Name)
	}

	switch {
	case *execFlag:
		if err := runBash(buf.String()); err != nil {
			log.Fatalf("exec error: %v", err)
		}
	case *dryRun:
		fmt.Print(buf.String())
	default:
		fmt.Print(buf.String())
	}
}

// runBash pipes the script to `bash -s`, connecting stdio so tmux attach works.
func runBash(script string) error {
	cmd := exec.Command("bash", "-s")
	cmd.Stdin = bytes.NewBufferString(script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Session struct {
	Name      string
	Directory string
	Windows   []Window
}

func (s Session) CreateSession(writer io.Writer) {
	params := gomux.SessionAttr{
		Name:      s.Name,
		Directory: s.Directory,
	}
	gomux.KillSession(s.Name, writer)
	gs := gomux.NewSessionAttr(params, writer)
	for _, w := range s.Windows {
		w.CreateWindow(gs)
	}
}

type Window struct {
	Name      string
	Pane      *Pane
	Directory string
}

func (w Window) CreateWindow(s *gomux.Session) {
	attr := gomux.WindowAttr{
		Name:      w.Name,
		Directory: w.Directory,
	}
	w1 := s.AddWindowAttr(attr)
	w1p0 := w1.Pane(0)
	if w.Pane != nil {
		w.Pane.pane = w1p0
		if w.Pane.Directory != "" {
			w.Pane.pane.Exec(fmt.Sprintf("cd '%s'", w.Pane.Directory))
		}
		w.Pane.ExecCommand()
		w.Pane.SplitPane()
	}
}

// Pane is what's in a tmux window e.g. zero
// or more splits.
type Pane struct {
	pane      *gomux.Pane
	Command   string
	Directory string
	SplitType string
	Split     []*Pane
}

func (p Pane) ExecCommand() {
	p.pane.Exec(p.Command)
}

func (p Pane) SplitPane() {
	for _, split := range p.Split {
		attr := gomux.SplitAttr{
			Directory: p.Directory,
		}
		if split.SplitType == "horizontal" {
			split.pane = p.pane.SplitWAttr(attr)
		} else {
			split.pane = p.pane.VsplitWAttr(attr)
		}
		if split.Directory != "" {
			split.pane.Exec(fmt.Sprintf("cd '%s'", split.Directory))
		}
		split.pane.Exec(split.Command)
		split.SplitPane()
	}
}
