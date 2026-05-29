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

Usage:
  json2tmux [flags] [file]
  cat session.json | json2tmux [flags]

Flags:
  -f <file>   Session JSON file (alternative to positional arg or stdin)
  -attach     Append 'tmux attach -t <name>' to output (or attach directly with -exec)
  -exec       Execute tmux commands directly instead of printing bash
  -dry-run    Print bash commands for inspection without executing
  -h, --help  Print this help

JSON schema:
  {
    "Name":      string,           // tmux session name (required)
    "Directory": string,           // working directory for the session
    "Windows": [
      {
        "Name":      string,       // tmux window name
        "Directory": string,       // working directory for the window
        "Pane": {
          "Command":   string,     // command to run in this pane
          "Directory": string,     // working directory for child splits
          "SplitType": string,     // "horizontal" or "vertical" (default)
          "Split":     [<Pane>]    // recursive child panes
        }
      }
    ]
  }

Notes:
  - Default output is bash commands; pipe to bash to execute.
  - With -exec, commands run directly — no pipe required.
  - Existing session with same Name is killed before creation.
  - Panes are split depth-first.

Examples:
  json2tmux session.json | bash
  json2tmux -exec session.json
  json2tmux -exec -attach session.json
  json2tmux -f session.json | bash
  cat session.json | json2tmux -exec
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
