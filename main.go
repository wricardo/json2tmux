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

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)

	var s Session
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		log.Fatal(err)
	}

	s.CreateSession(os.Stdout)
}

type Session struct {
	Name    string
    Directory string
	Windows []Window
}

func (s Session) CreateSession(writer io.Writer) {
    var gs *gomux.Session
    if s.Directory != "" {
        gs = gomux.NewSessionOnDir(s.Name, s.Directory, writer)
    }else {
        gs = gomux.NewSession(s.Name, writer)
    }
	for _, w := range s.Windows {
		w.CreateWindow(gs)
	}
}

type Window struct {
	Name string
	Pane *Pane
}

func (w Window) CreateWindow(s *gomux.Session) {
	w1 := s.AddWindow(w.Name)
	w1p0 := w1.Pane(0)
	if w.Pane != nil {
		w.Pane.pane = w1p0
		w.Pane.ExecCommand()
		w.Pane.SplitPane()
	}
}

type Pane struct {
	pane      *gomux.Pane
	Command   string
	SplitType string
	Split     []*Pane
}

func (p Pane) ExecCommand() {
	p.pane.Exec(p.Command)
}

func (p Pane) SplitPane() {
	for _, split := range p.Split {
		if split.SplitType == "horizontal" {
			split.pane = p.pane.Split()
		} else {
			fmt.Println(p.SplitType)
			split.pane = p.pane.Vsplit()
		}
		split.pane.Exec(split.Command)
		split.SplitPane()
	}
}
