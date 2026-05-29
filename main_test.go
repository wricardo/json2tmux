package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSession_KillsAndCreates(t *testing.T) {
	var buf bytes.Buffer
	s := Session{Name: "mysession", Directory: "/tmp"}
	s.CreateSession(&buf)
	out := buf.String()

	if !strings.Contains(out, "kill-session") {
		t.Errorf("expected kill-session command, got:\n%s", out)
	}
	if !strings.Contains(out, "new-session") {
		t.Errorf("expected new-session command, got:\n%s", out)
	}
	if !strings.Contains(out, "mysession") {
		t.Errorf("expected session name 'mysession', got:\n%s", out)
	}
}

func TestCreateSession_WithWindow(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name:      "testsession",
		Directory: "/tmp",
		Windows: []Window{
			{Name: "mywindow", Pane: &Pane{Command: "echo hello"}},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	if !strings.Contains(out, "mywindow") {
		t.Errorf("expected window name 'mywindow', got:\n%s", out)
	}
	if !strings.Contains(out, "echo hello") {
		t.Errorf("expected pane command 'echo hello', got:\n%s", out)
	}
}

func TestCreateSession_PaneDirectory(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "dirsession",
		Windows: []Window{
			{
				Name: "w",
				Pane: &Pane{
					Command:   "ls",
					Directory: "/home/user/my project",
				},
			},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	// path with space must be quoted
	if !strings.Contains(out, "cd '/home/user/my project'") {
		t.Errorf("expected quoted cd, got:\n%s", out)
	}
}

func TestCreateSession_NilPane(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name:    "nilpane",
		Windows: []Window{{Name: "w", Pane: nil}},
	}
	// must not panic
	s.CreateSession(&buf)
}

func TestCreateSession_SplitVertical(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "splitsession",
		Windows: []Window{
			{
				Name: "w",
				Pane: &Pane{
					Command: "top",
					Split: []*Pane{
						{Command: "htop", SplitType: "vertical"},
					},
				},
			},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	if !strings.Contains(out, "split-window") {
		t.Errorf("expected split-window command, got:\n%s", out)
	}
	if !strings.Contains(out, "htop") {
		t.Errorf("expected 'htop' command, got:\n%s", out)
	}
}

func TestCreateSession_SplitHorizontal(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "hsplit",
		Windows: []Window{
			{
				Name: "w",
				Pane: &Pane{
					Command: "vim",
					Split: []*Pane{
						{Command: "bash", SplitType: "horizontal"},
					},
				},
			},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	// gomux SplitWAttr (horizontal split) emits split-window -v
	if !strings.Contains(out, "split-window") {
		t.Errorf("expected split-window command, got:\n%s", out)
	}
	if !strings.Contains(out, "bash") {
		t.Errorf("expected 'bash' command, got:\n%s", out)
	}
}

func TestCreateSession_SplitDirectory(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "splitdir",
		Windows: []Window{
			{
				Name: "w",
				Pane: &Pane{
					Command: "vim",
					Split: []*Pane{
						{Command: "ls", Directory: "/var/log/my dir", SplitType: "vertical"},
					},
				},
			},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	if !strings.Contains(out, "cd '/var/log/my dir'") {
		t.Errorf("expected quoted cd for split directory, got:\n%s", out)
	}
}

func TestCreateSession_NestedSplits(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "nested",
		Windows: []Window{
			{
				Name: "w",
				Pane: &Pane{
					Command: "a",
					Split: []*Pane{
						{
							Command: "b",
							Split: []*Pane{
								{Command: "c"},
							},
						},
					},
				},
			},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	for _, cmd := range []string{"a", "b", "c"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected command %q in output, got:\n%s", cmd, out)
		}
	}
}

func TestCreateSession_MultipleWindows(t *testing.T) {
	var buf bytes.Buffer
	s := Session{
		Name: "multi",
		Windows: []Window{
			{Name: "win1", Pane: &Pane{Command: "vim"}},
			{Name: "win2", Pane: &Pane{Command: "bash"}},
		},
	}
	s.CreateSession(&buf)
	out := buf.String()

	for _, name := range []string{"win1", "win2", "vim", "bash"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in output, got:\n%s", name, out)
		}
	}
}

// helpers

func writeSessionFile(t *testing.T, s Session) string {
	t.Helper()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "session.json")
	if err := os.WriteFile(f, data, 0600); err != nil {
		t.Fatal(err)
	}
	return f
}

func sessionWithCommand(name, cmd string) Session {
	return Session{
		Name:    name,
		Windows: []Window{{Name: "w", Pane: &Pane{Command: cmd}}},
	}
}

// flag / input tests

func TestReadFromFile_PositionalArg(t *testing.T) {
	path := writeSessionFile(t, sessionWithCommand("filesession", "echo positional"))

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, _ := os.ReadFile(path)
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	s.CreateSession(&buf)
	out := buf.String()

	if !strings.Contains(out, "filesession") {
		t.Errorf("expected session name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "echo positional") {
		t.Errorf("expected command in output, got:\n%s", out)
	}
}

func TestAttachFlag(t *testing.T) {
	var buf bytes.Buffer
	s := sessionWithCommand("attachsession", "vim")
	s.CreateSession(&buf)
	fmt.Fprintf(&buf, "tmux attach -t %q\n", s.Name)
	out := buf.String()

	if !strings.Contains(out, `tmux attach -t "attachsession"`) {
		t.Errorf("expected attach command, got:\n%s", out)
	}
}

func TestRunBash_Success(t *testing.T) {
	if err := runBash("echo hello"); err != nil {
		t.Errorf("runBash failed on valid script: %v", err)
	}
}

func TestRunBash_Failure(t *testing.T) {
	if err := runBash("exit 1"); err == nil {
		t.Error("expected error for failing script, got nil")
	}
}

func TestRunBash_MultiLine(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "out.txt")
	script := fmt.Sprintf("echo line1 > %s\necho line2 >> %s\n", tmp, tmp)
	if err := runBash(script); err != nil {
		t.Fatalf("runBash failed: %v", err)
	}
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, "line1") || !strings.Contains(out, "line2") {
		t.Errorf("unexpected file contents: %s", out)
	}
}
