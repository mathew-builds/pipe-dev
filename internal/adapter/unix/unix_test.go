package unix

import (
	"testing"

	"github.com/mathew-builds/pipe-dev/internal/pipeline"
)

func TestParse(t *testing.T) {
	a := &Adapter{}

	tests := []struct {
		name       string
		input      string
		wantStages int
		wantNames  []string
		wantArgs   [][]string
		wantErr    bool
	}{
		{
			name:       "single command",
			input:      "cat data.json",
			wantStages: 1,
			wantNames:  []string{"cat"},
			wantArgs:   [][]string{{"data.json"}},
		},
		{
			name:       "two commands piped",
			input:      "cat data.json | jq '.[]'",
			wantStages: 2,
			wantNames:  []string{"cat", "jq"},
			wantArgs:   [][]string{{"data.json"}, {"'.[]'"}},
		},
		{
			name:       "four commands piped",
			input:      "cat data.json | jq '.[]' | sort | uniq -c",
			wantStages: 4,
			wantNames:  []string{"cat", "jq", "sort", "uniq"},
			wantArgs:   [][]string{{"data.json"}, {"'.[]'"}, nil, {"-c"}},
		},
		{
			name:       "extra whitespace",
			input:      "  cat foo  |  grep bar  |  wc -l  ",
			wantStages: 3,
			wantNames:  []string{"cat", "grep", "wc"},
			wantArgs:   [][]string{{"foo"}, {"bar"}, {"-l"}},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only pipes",
			input:   "| | |",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := a.Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(p.Stages) != tt.wantStages {
				t.Fatalf("got %d stages, want %d", len(p.Stages), tt.wantStages)
			}

			for i, stage := range p.Stages {
				if stage.Name != tt.wantNames[i] {
					t.Errorf("stage %d: name = %q, want %q", i, stage.Name, tt.wantNames[i])
				}
				if stage.Status != pipeline.StatusPending {
					t.Errorf("stage %d: status = %v, want StatusPending", i, stage.Status)
				}
				if stage.Stats == nil {
					t.Errorf("stage %d: stats is nil", i)
				}

				wantArgs := tt.wantArgs[i]
				if len(wantArgs) == 0 && len(stage.Args) == 0 {
					continue
				}
				if len(stage.Args) != len(wantArgs) {
					t.Errorf("stage %d: got %d args, want %d", i, len(stage.Args), len(wantArgs))
					continue
				}
				for j, arg := range stage.Args {
					if arg != wantArgs[j] {
						t.Errorf("stage %d arg %d: got %q, want %q", i, j, arg, wantArgs[j])
					}
				}
			}
		})
	}
}

func TestName(t *testing.T) {
	a := &Adapter{}
	if a.Name() != "unix" {
		t.Errorf("Name() = %q, want %q", a.Name(), "unix")
	}
}
