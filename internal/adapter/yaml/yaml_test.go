package yaml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "pipeline.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return path
}

func TestParseBasic(t *testing.T) {
	path := writeTempYAML(t, `
name: Test Pipeline
stages:
  - name: Generate
    command: seq 1 10
  - name: Filter
    command: grep 5
  - name: Count
    command: wc -l
`)

	a := &Adapter{}
	p, err := a.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if p.Name != "Test Pipeline" {
		t.Errorf("Name = %q, want %q", p.Name, "Test Pipeline")
	}
	if len(p.Stages) != 3 {
		t.Fatalf("got %d stages, want 3", len(p.Stages))
	}

	// Check stage details.
	want := []struct {
		name    string
		command string
		args    []string
	}{
		{"Generate", "seq", []string{"1", "10"}},
		{"Filter", "grep", []string{"5"}},
		{"Count", "wc", []string{"-l"}},
	}

	for i, w := range want {
		s := p.Stages[i]
		if s.Name != w.name {
			t.Errorf("stage %d: Name = %q, want %q", i, s.Name, w.name)
		}
		if s.Command != w.command {
			t.Errorf("stage %d: Command = %q, want %q", i, s.Command, w.command)
		}
		if s.Status != pipeline.StatusPending {
			t.Errorf("stage %d: Status = %v, want Pending", i, s.Status)
		}
		if len(s.Args) != len(w.args) {
			t.Errorf("stage %d: got %d args, want %d", i, len(s.Args), len(w.args))
			continue
		}
		for j, arg := range s.Args {
			if arg != w.args[j] {
				t.Errorf("stage %d arg %d: got %q, want %q", i, j, arg, w.args[j])
			}
		}
	}
}

func TestParseDefaultName(t *testing.T) {
	path := writeTempYAML(t, `
stages:
  - command: echo hello
`)

	a := &Adapter{}
	p, err := a.Parse(path)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if p.Name != "YAML Pipeline" {
		t.Errorf("Name = %q, want %q", p.Name, "YAML Pipeline")
	}

	// Stage name should default to the command name.
	if p.Stages[0].Name != "echo" {
		t.Errorf("stage name = %q, want %q", p.Stages[0].Name, "echo")
	}
}

func TestParseNoStages(t *testing.T) {
	path := writeTempYAML(t, `
name: Empty
stages: []
`)

	a := &Adapter{}
	_, err := a.Parse(path)
	if err == nil {
		t.Fatal("expected error for empty stages")
	}
}

func TestParseEmptyCommand(t *testing.T) {
	path := writeTempYAML(t, `
stages:
  - name: Bad
    command: ""
`)

	a := &Adapter{}
	_, err := a.Parse(path)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestParseFileNotFound(t *testing.T) {
	a := &Adapter{}
	_, err := a.Parse("/nonexistent/pipeline.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	path := writeTempYAML(t, `{{{invalid yaml`)

	a := &Adapter{}
	_, err := a.Parse(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestName(t *testing.T) {
	a := &Adapter{}
	if a.Name() != "yaml" {
		t.Errorf("Name() = %q, want %q", a.Name(), "yaml")
	}
}
