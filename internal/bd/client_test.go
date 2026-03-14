package bd

import "testing"

func TestParseStepPrefix(t *testing.T) {
	tests := []struct {
		input    string
		wantStep string
		wantText string
	}{
		{"[implement] code looks good", "implement", "code looks good"},
		{"[review] needs changes to auth.go", "review", "needs changes to auth.go"},
		{"no prefix here", "", "no prefix here"},
		{"[broken", "", "[broken"},
		{"[] empty step", "", "[] empty step"},
		{"[qa] ", "qa", ""},
	}

	for _, tt := range tests {
		step, text := parseStepPrefix(tt.input)
		if step != tt.wantStep || text != tt.wantText {
			t.Errorf("parseStepPrefix(%q) = (%q, %q), want (%q, %q)",
				tt.input, step, text, tt.wantStep, tt.wantText)
		}
	}
}

func TestNewClient_defaults(t *testing.T) {
	c := NewClient("", "")
	if c.BdPath != "bd" {
		t.Errorf("BdPath = %q, want %q", c.BdPath, "bd")
	}
	if c.DBPath != "" {
		t.Errorf("DBPath = %q, want empty", c.DBPath)
	}
}

func TestNewClient_custom(t *testing.T) {
	c := NewClient("/usr/local/bin/bd", "/path/to/beads")
	if c.BdPath != "/usr/local/bin/bd" {
		t.Errorf("BdPath = %q, want %q", c.BdPath, "/usr/local/bin/bd")
	}
	if c.DBPath != "/path/to/beads" {
		t.Errorf("DBPath = %q, want %q", c.DBPath, "/path/to/beads")
	}
}

func TestBaseArgs_withDB(t *testing.T) {
	c := NewClient("bd", "/path/to/beads")
	args := c.baseArgs()
	if len(args) != 3 || args[0] != "--db" || args[1] != "/path/to/beads" || args[2] != "--json" {
		t.Errorf("baseArgs() = %v, want [--db /path/to/beads --json]", args)
	}
}

func TestBaseArgs_noDB(t *testing.T) {
	c := NewClient("bd", "")
	args := c.baseArgs()
	if len(args) != 1 || args[0] != "--json" {
		t.Errorf("baseArgs() = %v, want [--json]", args)
	}
}
