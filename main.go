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
	Name      string
	Directory string
	Windows   []Window
}

func (s Session) CreateSession(writer io.Writer) {
	params := gomux.SessionAttr{
		Name:      s.Name,
		Directory: s.Directory,
	}
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
		w.Pane.ExecCommand()
		w.Pane.SplitPane()
	}
}

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
			fmt.Println(p.SplitType)
			split.pane = p.pane.VsplitWAttr(attr)
		}
		split.pane.Exec(split.Command)
		split.SplitPane()
	}
}
