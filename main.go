package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/wricardo/gomux"
)

const usage = `json2tmux — create tmux sessions from JSON

Usage:
  cat session.json | json2tmux | bash

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
  - Output is bash commands; pipe to bash to execute.
  - Existing session with same Name is killed before creation.
  - Panes are split depth-first.
`

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(0)
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	var s Session
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid JSON: %v\n\nRun with --help for usage.\n", err)
		os.Exit(1)
	}

	s.CreateSession(os.Stdout)
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
