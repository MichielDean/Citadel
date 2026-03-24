package cataractae

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MichielDean/cistern/internal/cistern"
	"github.com/MichielDean/cistern/internal/provider"
)

func TestBuildClaudeCmd_ContainsAddDir(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	skillsDir := "/home/user/.cistern/skills"
	cmd := s.buildClaudeCmd(skillsDir)
	if !strings.Contains(cmd, "--add-dir") {
		t.Errorf("claudeCmd missing --add-dir flag: %s", cmd)
	}
	if !strings.Contains(cmd, skillsDir) {
		t.Errorf("claudeCmd missing skillsDir %q: %s", skillsDir, cmd)
	}
}

func TestBuildClaudeCmd_QuotesPathWithSpaces(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	skillsDir := "/home/john doe/.cistern/skills"
	cmd := s.buildClaudeCmd(skillsDir)

	// Unquoted form must not appear — it would split at the space.
	if strings.Contains(cmd, "--add-dir /home/john doe/") {
		t.Errorf("claudeCmd contains unquoted path with space — will break shell: %s", cmd)
	}
	// Shell-quoted form must be present.
	want := "--add-dir '/home/john doe/.cistern/skills'"
	if !strings.Contains(cmd, want) {
		t.Errorf("claudeCmd missing shell-quoted skillsDir\nwant substring: %s\ngot: %s", want, cmd)
	}
}

func TestBuildClaudeCmd_WithModel(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp", Model: "haiku"}
	cmd := s.buildClaudeCmd("/home/user/.cistern/skills")
	if !strings.Contains(cmd, "--model haiku") {
		t.Errorf("claudeCmd missing --model flag: %s", cmd)
	}
}

func TestBuildClaudeCmd_WithoutModel(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	cmd := s.buildClaudeCmd("/home/user/.cistern/skills")
	if strings.Contains(cmd, "--model") {
		t.Errorf("claudeCmd should not contain --model when model is empty: %s", cmd)
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/.cistern/skills", "'/home/user/.cistern/skills'"},
		{"/home/john doe/.cistern/skills", "'/home/john doe/.cistern/skills'"},
		{"it's a path", "'it'\\''s a path'"},
	}
	for _, tt := range tests {
		got := shellQuote(tt.input)
		if got != tt.want {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBuildPrompt_WithIdentity_FileFound(t *testing.T) {
	dir := t.TempDir()
	identityDir := filepath.Join(dir, ".cistern", "cataractae", "implementer")
	if err := os.MkdirAll(identityDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(identityDir, "CLAUDE.md"),
		[]byte("# Implementer\n\nYou implement things.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	s := &Session{ID: "test", WorkDir: dir, Identity: "implementer"}
	prompt := s.buildPrompt()

	if !strings.Contains(prompt, "## Your Role") {
		t.Error("prompt missing '## Your Role' section when identity file is present")
	}
	if !strings.Contains(prompt, "You implement things.") {
		t.Error("prompt missing identity file content")
	}
	if !strings.Contains(prompt, baseCataractaePrompt) {
		t.Error("prompt missing constitutional base")
	}
}

func TestBuildPrompt_WithIdentity_FileMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir) // no CLAUDE.md at cistern identity path

	s := &Session{ID: "test", WorkDir: dir, Identity: "implementer"}
	prompt := s.buildPrompt()

	// Fallback: prompt contains the actual missing path, not just any occurrence of "Read".
	if !strings.Contains(prompt, "cataractae/implementer/CLAUDE.md") {
		t.Error("prompt missing fallback path 'cataractae/implementer/CLAUDE.md' when identity file is missing")
	}
	if !strings.Contains(prompt, "implementer") {
		t.Error("prompt missing identity name in fallback")
	}
	if strings.Contains(prompt, "## Your Role") {
		t.Error("prompt should not contain '## Your Role' when identity file is missing")
	}
}

func TestResolveIdentityPath_CisternHome(t *testing.T) {
	dir := t.TempDir()
	cisternPath := filepath.Join(dir, ".cistern", "cataractae", "reviewer", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(cisternPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cisternPath, []byte("# Reviewer"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	s := &Session{Identity: "reviewer"}
	got := s.resolveIdentityPath()
	if got != cisternPath {
		t.Errorf("resolveIdentityPath = %q, want %q", got, cisternPath)
	}
}

func TestResolveIdentityPath_FallbackSandbox(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir) // no CLAUDE.md at cistern identity path

	s := &Session{Identity: "implementer"}
	got := s.resolveIdentityPath()
	want := "cataractae/implementer/CLAUDE.md"
	if got != want {
		t.Errorf("resolveIdentityPath = %q, want %q", got, want)
	}
}

func TestClaudePath_EnvOverride(t *testing.T) {
	t.Setenv("CLAUDE_PATH", "/usr/local/bin/my-claude")
	got := claudePath()
	if got != "/usr/local/bin/my-claude" {
		t.Errorf("claudePath() = %q, want %q", got, "/usr/local/bin/my-claude")
	}
}

func TestClaudePath_LookPath(t *testing.T) {
	t.Setenv("CLAUDE_PATH", "")
	// Place a fake "claude" executable on PATH so exec.LookPath finds it.
	dir := t.TempDir()
	fakeClaude := filepath.Join(dir, "claude")
	if err := os.WriteFile(fakeClaude, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))

	got := claudePath()
	if got != fakeClaude {
		t.Errorf("claudePath() = %q, want %q", got, fakeClaude)
	}
}

// TestClaudePresetBackwardCompat is the backward-compatibility regression test
// for the session.go refactor (ci-sc2wl).
//
// Given: an AqueductConfig with no provider block (defaults to the "claude"
// built-in preset), when buildPresetCmd is called with that preset it must
// produce a command string byte-for-byte identical to what the legacy
// buildClaudeCmd produces today.
//
// This test must stay green before ci-sc2wl merges.
func TestClaudePresetBackwardCompat(t *testing.T) {
	// Normalise claudePath() to "claude" so both code paths agree on the
	// executable name regardless of what is installed on this machine.
	t.Setenv("CLAUDE_PATH", "claude")

	var claudePreset provider.ProviderPreset
	for _, p := range provider.Builtins() {
		if p.Name == "claude" {
			claudePreset = p
			break
		}
	}
	if claudePreset.Name == "" {
		t.Fatal("claude preset not found in Builtins()")
	}

	skillsDir := "/home/user/.cistern/skills"

	t.Run("without model", func(t *testing.T) {
		s := &Session{ID: "test", WorkDir: "/tmp"}
		want := s.buildClaudeCmd(skillsDir)
		got, err := s.buildPresetCmd(claudePreset, skillsDir)
		if err != nil {
			t.Fatalf("buildPresetCmd: %v", err)
		}
		if got != want {
			t.Errorf("backward compat broken (no model):\nwant: %q\ngot:  %q", want, got)
		}
	})

	t.Run("with model", func(t *testing.T) {
		s := &Session{ID: "test", WorkDir: "/tmp", Model: "haiku"}
		want := s.buildClaudeCmd(skillsDir)
		got, err := s.buildPresetCmd(claudePreset, skillsDir)
		if err != nil {
			t.Fatalf("buildPresetCmd: %v", err)
		}
		if got != want {
			t.Errorf("backward compat broken (with model):\nwant: %q\ngot:  %q", want, got)
		}
	})

	t.Run("skills dir with spaces", func(t *testing.T) {
		s := &Session{ID: "test", WorkDir: "/tmp"}
		dir := "/home/john doe/.cistern/skills"
		want := s.buildClaudeCmd(dir)
		got, err := s.buildPresetCmd(claudePreset, dir)
		if err != nil {
			t.Fatalf("buildPresetCmd: %v", err)
		}
		if got != want {
			t.Errorf("backward compat broken (spaces in path):\nwant: %q\ngot:  %q", want, got)
		}
	})

	// This subtest verifies the LookPath contract: when claudePath() resolves to
	// an absolute path (e.g. /usr/local/bin/claude via exec.LookPath), the preset's
	// Command field must carry that same resolved path for buildPresetCmd to produce
	// a command identical to buildClaudeCmd. The test patches claudePathFn directly
	// so that neither CLAUDE_PATH nor a real binary installation is required.
	t.Run("LookPath resolution — preset Command must carry resolved absolute path", func(t *testing.T) {
		// Clear the parent's CLAUDE_PATH=claude to exercise the LookPath code path.
		t.Setenv("CLAUDE_PATH", "")

		// Patch claudePathFn to simulate LookPath resolving to an absolute path.
		const resolvedPath = "/opt/test/claude"
		orig := claudePathFn
		claudePathFn = func() string { return resolvedPath }
		t.Cleanup(func() { claudePathFn = orig })

		// The preset must carry the same resolved path; without it the commands diverge.
		resolvedPreset := claudePreset
		resolvedPreset.Command = resolvedPath

		s := &Session{ID: "test", WorkDir: "/tmp"}
		want := s.buildClaudeCmd(skillsDir)
		got, err := s.buildPresetCmd(resolvedPreset, skillsDir)
		if err != nil {
			t.Fatalf("buildPresetCmd: %v", err)
		}
		if got != want {
			t.Errorf("LookPath compat broken — preset.Command must match resolved path:\nwant: %q\ngot:  %q", want, got)
		}
	})
}

// TestClaudeDefaultFallback is the non-negotiable regression gate for the
// provider-preset refactor.
//
// Given: an AqueductConfig with no provider block (zero value — no Provider
// field set), the resolved preset must be "claude", and the command string
// produced by buildPresetCmd must match buildClaudeCmd() output
// character-for-character.
func TestClaudeDefaultFallback(t *testing.T) {
	// Normalise claudePathFn so both code paths agree on the binary name.
	t.Setenv("CLAUDE_PATH", "claude")

	// Resolve preset: empty provider name must return the "claude" built-in.
	preset := provider.ResolvePreset("")
	if preset.Name != "claude" {
		t.Fatalf("ResolvePreset(\"\") = %q, want %q", preset.Name, "claude")
	}

	skillsDir := "/home/user/.cistern/skills"

	t.Run("without model", func(t *testing.T) {
		s := &Session{ID: "test", WorkDir: "/tmp"}
		want := s.buildClaudeCmd(skillsDir)
		got, err := s.buildPresetCmd(preset, skillsDir)
		if err != nil {
			t.Fatalf("buildPresetCmd error: %v", err)
		}
		if got != want {
			t.Errorf("default fallback command mismatch (no model):\nwant: %q\ngot:  %q", want, got)
		}
	})

	t.Run("with model", func(t *testing.T) {
		s := &Session{ID: "test", WorkDir: "/tmp", Model: "haiku"}
		want := s.buildClaudeCmd(skillsDir)
		got, err := s.buildPresetCmd(preset, skillsDir)
		if err != nil {
			t.Fatalf("buildPresetCmd error: %v", err)
		}
		if got != want {
			t.Errorf("default fallback command mismatch (with model):\nwant: %q\ngot:  %q", want, got)
		}
	})
}

// buildTestBin compiles the Go package at importPath into a temp directory
// and returns the absolute path to the resulting binary.
func buildTestBin(t *testing.T, name, importPath string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), name)
	out, err := exec.Command("go", "build", "-o", bin, importPath).CombinedOutput()
	if err != nil {
		t.Fatalf("build %s: %v\n%s", name, err, out)
	}
	return bin
}

// TestFakeagent_SpawnOutcomeCycle exercises the full Session.Spawn →
// Session.isAlive → droplet outcome pipeline using the fakeagent binary.
//
// The test is skipped when tmux is unavailable (e.g. in minimal CI
// environments) so that 'go test ./...' never hard-fails on missing
// infrastructure.
func TestFakeagent_SpawnOutcomeCycle(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available — skipping integration test")
	}

	// Build fakeagent and ct.
	fakeagentBin := buildTestBin(t, "fakeagent", "github.com/MichielDean/cistern/internal/testutil/fakeagent")
	ctBin := buildTestBin(t, "ct", "github.com/MichielDean/cistern/cmd/ct")

	// Add both binaries to a temporary PATH so fakeagent can call 'ct'.
	binDir := filepath.Dir(ctBin)
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	// Create an isolated cistern DB.
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "test.db")
	t.Setenv("CT_DB", dbPath)
	t.Setenv("CT_NO_ASCII_LOGO", "1")

	c, err := cistern.New(dbPath, "fa")
	if err != nil {
		t.Fatalf("cistern.New: %v", err)
	}
	defer c.Close()

	// Add a test droplet so fakeagent has an ID to pass.
	droplet, err := c.Add("testrepo", "fakeagent test", "desc", 1, 2)
	if err != nil {
		t.Fatalf("cistern.Add: %v", err)
	}

	// Write CONTEXT.md into the WorkDir with the droplet ID.
	workDir := t.TempDir()
	contextContent := fmt.Sprintf("# Context\n\n## Item: %s\n\n**Title:** fakeagent test\n", droplet.ID)
	if err := os.WriteFile(filepath.Join(workDir, "CONTEXT.md"), []byte(contextContent), 0o644); err != nil {
		t.Fatalf("write CONTEXT.md: %v", err)
	}

	// Point CLAUDE_PATH at the fakeagent binary.
	t.Setenv("CLAUDE_PATH", fakeagentBin)

	// Spawn the session.
	sessionID := "ci-t3xo9-fa-" + droplet.ID
	s := &Session{
		ID:      sessionID,
		WorkDir: workDir,
	}
	if err := s.Spawn(); err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	t.Cleanup(func() { s.kill() })

	// Wait for the session to die (fakeagent exits after calling ct droplet pass).
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if !s.isAlive() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if s.isAlive() {
		t.Fatal("session still alive after 15s — fakeagent did not exit")
	}

	// Verify the outcome was recorded.
	got, err := c.Get(droplet.ID)
	if err != nil {
		t.Fatalf("cistern.Get: %v", err)
	}
	if got.Outcome != "pass" {
		t.Errorf("droplet outcome = %q, want %q", got.Outcome, "pass")
	}
}

// TestIsAgentAlive_ProcessNameMatches_ReturnsTrue verifies that isAgentAlive
// returns true when the pane's current command matches one of the preset's
// ProcessNames.
func TestIsAgentAlive_ProcessNameMatches_ReturnsTrue(t *testing.T) {
	orig := tmuxDisplayMessage
	tmuxDisplayMessage = func(id string) (string, error) { return "claude", nil }
	t.Cleanup(func() { tmuxDisplayMessage = orig })

	s := &Session{
		ID:     "test-session",
		Preset: provider.ProviderPreset{ProcessNames: []string{"claude", "node"}},
	}
	if !s.isAgentAlive() {
		t.Error("isAgentAlive() = false, want true when pane_current_command is in ProcessNames")
	}
}

// TestIsAgentAlive_ProcessNameNotMatched_ReturnsFalse verifies that isAgentAlive
// returns false when the pane's current command is not in ProcessNames — this is
// a zombie session (tmux alive, agent dead).
func TestIsAgentAlive_ProcessNameNotMatched_ReturnsFalse(t *testing.T) {
	orig := tmuxDisplayMessage
	tmuxDisplayMessage = func(id string) (string, error) { return "bash", nil }
	t.Cleanup(func() { tmuxDisplayMessage = orig })

	s := &Session{
		ID:     "test-session",
		Preset: provider.ProviderPreset{ProcessNames: []string{"claude", "node"}},
	}
	if s.isAgentAlive() {
		t.Error("isAgentAlive() = true, want false when pane_current_command is not in ProcessNames")
	}
}

// TestIsAgentAlive_EmptyProcessNames_ReturnsTrue verifies that isAgentAlive
// returns true when no ProcessNames are configured — the preset has no way to
// detect zombie sessions so it conservatively assumes the agent is alive.
func TestIsAgentAlive_EmptyProcessNames_ReturnsTrue(t *testing.T) {
	orig := tmuxDisplayMessage
	tmuxDisplayMessage = func(id string) (string, error) { return "bash", nil }
	t.Cleanup(func() { tmuxDisplayMessage = orig })

	s := &Session{
		ID:     "test-session",
		Preset: provider.ProviderPreset{},
	}
	if !s.isAgentAlive() {
		t.Error("isAgentAlive() = false, want true when ProcessNames is empty (no detection configured)")
	}
}

// TestIsAgentAlive_TmuxError_ReturnsFalse verifies that isAgentAlive returns
// false when the tmux display-message call fails — treat an unqueryable session
// as a dead agent.
func TestIsAgentAlive_TmuxError_ReturnsFalse(t *testing.T) {
	orig := tmuxDisplayMessage
	tmuxDisplayMessage = func(id string) (string, error) {
		return "", errors.New("tmux: can't find session: test-session")
	}
	t.Cleanup(func() { tmuxDisplayMessage = orig })

	s := &Session{
		ID:     "test-session",
		Preset: provider.ProviderPreset{ProcessNames: []string{"claude"}},
	}
	if s.isAgentAlive() {
		t.Error("isAgentAlive() = true, want false when tmux command errors")
	}
}

// TestBuildPresetCmd_EmptyCommand_ReturnsError verifies that buildPresetCmd
// returns a descriptive error when the preset has no command configured.
// A misconfigured provider with Name set but Command empty must not silently
// produce a broken tmux command.
func TestBuildPresetCmd_EmptyCommand_ReturnsError(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	preset := provider.ProviderPreset{Name: "custom"} // Command is deliberately empty
	_, err := s.buildPresetCmd(preset, "/skills")
	if err == nil {
		t.Fatal("expected error for preset with empty Command, got nil")
	}
	if !strings.Contains(err.Error(), "custom") {
		t.Errorf("error %q should mention the preset name", err.Error())
	}
	if !strings.Contains(err.Error(), "no command configured") {
		t.Errorf("error %q should mention 'no command configured'", err.Error())
	}
}

// TestBuildPresetCmd_PromptFlag_AppendedWhenNonEmpty verifies that buildPresetCmd
// uses preset.PromptFlag to deliver the prompt when it is set.
func TestBuildPresetCmd_PromptFlag_AppendedWhenNonEmpty(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	preset := provider.ProviderPreset{
		Name:       "myagent",
		Command:    "myagent",
		PromptFlag: "--prompt",
	}
	cmd, err := s.buildPresetCmd(preset, "/skills")
	if err != nil {
		t.Fatalf("buildPresetCmd: %v", err)
	}
	if !strings.Contains(cmd, "--prompt") {
		t.Errorf("buildPresetCmd output missing PromptFlag: %s", cmd)
	}
}

// TestBuildPresetCmd_PromptFlag_OmittedWhenEmpty verifies that buildPresetCmd
// does not append any prompt flag when PromptFlag is empty. Presets for CLIs
// that do not accept -p (e.g. opencode) must have PromptFlag="" to avoid
// spawn failures from unrecognized flags.
func TestBuildPresetCmd_PromptFlag_OmittedWhenEmpty(t *testing.T) {
	s := &Session{ID: "test", WorkDir: "/tmp"}
	preset := provider.ProviderPreset{
		Name:    "opencode",
		Command: "opencode",
		// PromptFlag deliberately empty — prompt delivered via instructions file
	}
	cmd, err := s.buildPresetCmd(preset, "/skills")
	if err != nil {
		t.Fatalf("buildPresetCmd: %v", err)
	}
	if strings.Contains(cmd, " -p ") || strings.Contains(cmd, " --prompt") {
		t.Errorf("buildPresetCmd with empty PromptFlag should not contain a prompt flag: %s", cmd)
	}
	if !strings.HasPrefix(cmd, "opencode") {
		t.Errorf("buildPresetCmd output = %q, want prefix %q", cmd, "opencode")
	}
}

// TestCollectEnvArgs_GHToken_AlwaysForwarded_PresetPath verifies that GH_TOKEN
// is included in env args when using the preset path. This is a regression test
// for the ci-sc2wl refactor: the legacy path forwarded GH_TOKEN but the preset
// path only iterated EnvPassthrough (which did not include GH_TOKEN for claude).
func TestCollectEnvArgs_GHToken_AlwaysForwarded_PresetPath(t *testing.T) {
	t.Setenv("GH_TOKEN", "ghtoken-preset-123")
	t.Setenv("ANTHROPIC_API_KEY", "") // isolate to GH_TOKEN check

	s := &Session{
		ID:     "test",
		Preset: provider.ProviderPreset{Name: "claude", Command: "claude"},
	}
	args := s.collectEnvArgs()
	if !containsEnvPair(args, "GH_TOKEN", "ghtoken-preset-123") {
		t.Errorf("collectEnvArgs (preset path) missing GH_TOKEN; args: %v", args)
	}
}

// TestCollectEnvArgs_GHToken_AlwaysForwarded_LegacyPath verifies that GH_TOKEN
// is included in env args when using the legacy (no-preset) path.
func TestCollectEnvArgs_GHToken_AlwaysForwarded_LegacyPath(t *testing.T) {
	t.Setenv("GH_TOKEN", "ghtoken-legacy-456")

	s := &Session{ID: "test"} // Preset.Name is empty — legacy path
	args := s.collectEnvArgs()
	if !containsEnvPair(args, "GH_TOKEN", "ghtoken-legacy-456") {
		t.Errorf("collectEnvArgs (legacy path) missing GH_TOKEN; args: %v", args)
	}
}

// TestCollectEnvArgs_GHToken_AbsentWhenNotSet verifies that GH_TOKEN is not
// included in env args when it is unset in the environment.
func TestCollectEnvArgs_GHToken_AbsentWhenNotSet(t *testing.T) {
	t.Setenv("GH_TOKEN", "")

	s := &Session{ID: "test", Preset: provider.ProviderPreset{Name: "claude"}}
	args := s.collectEnvArgs()
	for _, a := range args {
		if strings.Contains(a, "GH_TOKEN") {
			t.Errorf("collectEnvArgs contains GH_TOKEN when unset; args: %v", args)
		}
	}
}

// containsEnvPair checks whether args contains "-e" followed by "key=val".
func containsEnvPair(args []string, key, val string) bool {
	target := key + "=" + val
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-e" && args[i+1] == target {
			return true
		}
	}
	return false
}

// TestIsAgentAlive_PassesSessionIDToDisplayMessage verifies that isAgentAlive
// forwards the session ID to tmuxDisplayMessage.
func TestIsAgentAlive_PassesSessionIDToDisplayMessage(t *testing.T) {
	var capturedID string
	orig := tmuxDisplayMessage
	tmuxDisplayMessage = func(id string) (string, error) {
		capturedID = id
		return "claude", nil
	}
	t.Cleanup(func() { tmuxDisplayMessage = orig })

	s := &Session{
		ID:     "myrepo-alice",
		Preset: provider.ProviderPreset{ProcessNames: []string{"claude"}},
	}
	s.isAgentAlive()
	if capturedID != "myrepo-alice" {
		t.Errorf("tmuxDisplayMessage called with id = %q, want %q", capturedID, "myrepo-alice")
	}
}
