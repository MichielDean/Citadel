package proc_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MichielDean/cistern/internal/proc"
)

// writeFakeProcEntry creates a minimal /proc/<pid> directory under procRoot
// with a status file containing the given ppid and a cmdline file with
// null-separated args.
func writeFakeProcEntry(t *testing.T, procRoot, pid, ppid string, argv ...string) {
	t.Helper()
	dir := filepath.Join(procRoot, pid)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	status := fmt.Sprintf("Name:\tsh\nPPid:\t%s\nUid:\t1000\n", ppid)
	if err := os.WriteFile(filepath.Join(dir, "status"), []byte(status), 0644); err != nil {
		t.Fatal(err)
	}
	cmdline := strings.Join(argv, "\x00") + "\x00"
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(cmdline), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestIsProcPIDEntry_ValidAndInvalid exercises the PID name filter.
func TestIsProcPIDEntry_ValidAndInvalid(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123", true},
		{"1", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"12a", false},
		{"net", false},
		{"self", false},
	}
	for _, tc := range cases {
		if got := proc.IsProcPIDEntry(tc.in); got != tc.want {
			t.Errorf("IsProcPIDEntry(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// TestParsePPid_ExtractsPPidLine verifies ParsePPid returns the PPid value.
func TestParsePPid_ExtractsPPidLine(t *testing.T) {
	status := "Name:\tbash\nPid:\t42\nPPid:\t7\nUid:\t1000\n"
	if got := proc.ParsePPid(status); got != "7" {
		t.Errorf("ParsePPid = %q, want %q", got, "7")
	}
}

// TestParsePPid_MissingLine returns empty string when PPid line is absent.
func TestParsePPid_MissingLine(t *testing.T) {
	if got := proc.ParsePPid("Name:\tbash\nPid:\t42\n"); got != "" {
		t.Errorf("ParsePPid = %q, want empty", got)
	}
}

// TestIsAgentCmdline covers positive and negative cases for the opencode agent binary.
func TestIsAgentCmdline(t *testing.T) {
	cases := []struct {
		cmdline string
		want    bool
	}{
		{"/home/user/.opencode/bin/opencode\x00run\x00--dangerously-skip-permissions\x00--model\x00ollama/glm-5.1:cloud\x00hello\x00", true},
		{"opencode\x00run\x00hello\x00", true},
		{"/bin/bash\x00-c\x00sleep 100\x00", false},
		{"sh\x00", false},
		{"", false},
		{"\x00", false},
	}
	for _, tc := range cases {
		if got := proc.IsAgentCmdline(tc.cmdline); got != tc.want {
			t.Errorf("IsAgentCmdline(%q) = %v, want %v", tc.cmdline, got, tc.want)
		}
	}
}

// TestAgentAliveUnderPIDIn_EmptyPID_ReturnsFalse ensures the empty PID guard fires.
func TestAgentAliveUnderPIDIn_EmptyPID_ReturnsFalse(t *testing.T) {
	if proc.AgentAliveUnderPIDIn("", t.TempDir()) {
		t.Error("expected false for empty panePIDStr")
	}
}

// TestAgentAliveUnderPIDIn_NoAgentProcess_ReturnsFalse verifies that a
// process tree with no agent process returns false.
func TestAgentAliveUnderPIDIn_NoAgentProcess_ReturnsFalse(t *testing.T) {
	procRoot := t.TempDir()
	// Pane PID 1 (shell), child 2 (another shell). Neither is an agent.
	writeFakeProcEntry(t, procRoot, "1", "0", "/bin/bash")
	writeFakeProcEntry(t, procRoot, "2", "1", "/usr/bin/python3", "script.py")
	if proc.AgentAliveUnderPIDIn("1", procRoot) {
		t.Error("expected false when no agent descendant exists")
	}
}

// TestAgentAliveUnderPIDIn_DirectAgentChild_ReturnsTrue verifies that a
// direct agent child of the pane PID is detected (opencode).
func TestAgentAliveUnderPIDIn_DirectAgentChild_ReturnsTrue(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"opencode"} {
		t.Run(name, func(t *testing.T) {
			procRoot := t.TempDir()
			writeFakeProcEntry(t, procRoot, "1", "0", "/bin/bash")
			writeFakeProcEntry(t, procRoot, "2", "1", "/usr/local/bin/"+name, "--dangerously-skip-permissions")
			if !proc.AgentAliveUnderPIDIn("1", procRoot) {
				t.Errorf("expected true when direct %s child exists", name)
			}
		})
	}
}

// TestAgentAliveUnderPIDIn_DeepDescendant_ReturnsTrue verifies that an agent
// is found even when it is several levels deep (bash → sh → node → agent).
func TestAgentAliveUnderPIDIn_DeepDescendant_ReturnsTrue(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"opencode"} {
		t.Run(name, func(t *testing.T) {
			procRoot := t.TempDir()
			writeFakeProcEntry(t, procRoot, "1", "0", "/bin/bash")
			writeFakeProcEntry(t, procRoot, "2", "1", "/bin/sh", "-c", "node launcher.js")
			writeFakeProcEntry(t, procRoot, "3", "2", "/usr/bin/node", "launcher.js")
			writeFakeProcEntry(t, procRoot, "4", "3", "/home/user/.local/bin/"+name, "arg")
			if !proc.AgentAliveUnderPIDIn("1", procRoot) {
				t.Errorf("expected true when %s is a deep descendant", name)
			}
		})
	}
}

// TestAgentAliveUnderPIDIn_UnrelatedAgentProcess_ReturnsFalse verifies that
// an agent process that is NOT a descendant of the pane PID is not reported.
func TestAgentAliveUnderPIDIn_UnrelatedAgentProcess_ReturnsFalse(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"opencode"} {
		t.Run(name, func(t *testing.T) {
			procRoot := t.TempDir()
			// Pane PID is 10; agent runs under PID 1 (unrelated).
			writeFakeProcEntry(t, procRoot, "1", "0", "/bin/bash")
			writeFakeProcEntry(t, procRoot, "2", "1", "/usr/bin/"+name)
			writeFakeProcEntry(t, procRoot, "10", "0", "/bin/bash") // our pane, no agent children
			if proc.AgentAliveUnderPIDIn("10", procRoot) {
				t.Errorf("expected false when %s is not a descendant of pane PID", name)
			}
		})
	}
}

// TestAgentAliveUnderPIDIn_NonexistentPanePID_ReturnsFalse handles a pane PID
// that no longer appears in /proc (process already gone).
func TestAgentAliveUnderPIDIn_NonexistentPanePID_ReturnsFalse(t *testing.T) {
	procRoot := t.TempDir()
	writeFakeProcEntry(t, procRoot, "1", "0", "/bin/bash")
	if proc.AgentAliveUnderPIDIn("9999", procRoot) {
		t.Error("expected false when pane PID does not appear in /proc")
	}
}
