package aqueduct

import (
	"slices"
	"strings"
	"testing"

	"github.com/MichielDean/cistern/internal/provider"
)

// --- ResolveProvider tests ---

// TestResolveProvider_DefaultsToOpencodeWhenNoProvider verifies that an AqueductConfig
// with no provider block resolves to the built-in opencode preset unchanged.
func TestResolveProvider_DefaultsToOpencodeWhenNoProvider(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{{Name: "myrepo", Cataractae: 1}},
	}
	preset, err := cfg.ResolveProvider("myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Name != "opencode" {
		t.Errorf("Name = %q, want %q", preset.Name, "opencode")
	}
	if preset.Command != "opencode" {
		t.Errorf("Command = %q, want %q", preset.Command, "opencode")
	}
}

// TestResolveProvider_TopLevelProviderName verifies that the top-level provider
// name selects the corresponding built-in preset.
func TestResolveProvider_TopLevelProviderName(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Name: "opencode"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Name != "opencode" {
		t.Errorf("Name = %q, want %q", preset.Name, "opencode")
	}
	if preset.Command != "opencode" {
		t.Errorf("Command = %q, want %q", preset.Command, "opencode")
	}
}

// TestResolveProvider_RepoNameOverridesTopLevelName verifies that a repo-level
// provider name takes precedence over the top-level provider name.
func TestResolveProvider_RepoNameOverridesTopLevelName(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Name: "opencode"}},
		},
		Provider: &ProviderConfig{Name: "opencode"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Name != "opencode" {
		t.Errorf("Name = %q, want %q", preset.Name, "opencode")
	}
}

// TestResolveProvider_TopLevelCommandOverride verifies that a top-level command
// override replaces the preset's executable.
func TestResolveProvider_TopLevelCommandOverride(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Command: "my-opencode"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Command != "my-opencode" {
		t.Errorf("Command = %q, want %q", preset.Command, "my-opencode")
	}
}

// TestResolveProvider_RepoCommandOverridesTopLevel verifies that a repo-level
// command override wins over the top-level command override.
func TestResolveProvider_RepoCommandOverridesTopLevel(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Command: "repo-opencode"}},
		},
		Provider: &ProviderConfig{Command: "top-opencode"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Command != "repo-opencode" {
		t.Errorf("Command = %q, want %q", preset.Command, "repo-opencode")
	}
}

// TestResolveProvider_ArgsAreAppended verifies that top-level and repo-level args
// are both appended to the preset's built-in args.
func TestResolveProvider_ArgsAreAppended(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Args: []string{"--repo-arg"}}},
		},
		Provider: &ProviderConfig{Args: []string{"--top-arg"}},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Contains(preset.Args, "--top-arg") {
		t.Errorf("Args %v missing --top-arg", preset.Args)
	}
	if !slices.Contains(preset.Args, "--repo-arg") {
		t.Errorf("Args %v missing --repo-arg", preset.Args)
	}
}

// TestResolveProvider_EnvIsMerged verifies that env entries from top-level and
// repo-level configs are merged, with repo-level values winning for shared keys.
func TestResolveProvider_EnvIsMerged(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{
				Env: map[string]string{"REPO_KEY": "repo-val", "SHARED": "from-repo"},
			}},
		},
		Provider: &ProviderConfig{
			Env: map[string]string{"TOP_KEY": "top-val", "SHARED": "from-top"},
		},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.ExtraEnv["TOP_KEY"] != "top-val" {
		t.Errorf("ExtraEnv[TOP_KEY] = %q, want %q", preset.ExtraEnv["TOP_KEY"], "top-val")
	}
	if preset.ExtraEnv["REPO_KEY"] != "repo-val" {
		t.Errorf("ExtraEnv[REPO_KEY] = %q, want %q", preset.ExtraEnv["REPO_KEY"], "repo-val")
	}
	// Repo value overrides top-level for shared keys.
	if preset.ExtraEnv["SHARED"] != "from-repo" {
		t.Errorf("ExtraEnv[SHARED] = %q, want %q", preset.ExtraEnv["SHARED"], "from-repo")
	}
}

// TestResolveProvider_UnknownProviderReturnsError verifies that an unknown preset
// name causes ResolveProvider to return an error.
func TestResolveProvider_UnknownProviderReturnsError(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Name: "nonexistent"},
	}
	_, err := cfg.ResolveProvider("r")
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "nonexistent")
	}
}

// TestResolveProvider_UnknownRepoFallsBackToTopLevel verifies that an unknown
// repo name causes only the top-level provider to be applied.
func TestResolveProvider_UnknownRepoFallsBackToTopLevel(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Name: "opencode"},
	}
	preset, err := cfg.ResolveProvider("nonexistent-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Name != "opencode" {
		t.Errorf("Name = %q, want %q", preset.Name, "opencode")
	}
}

// TestResolveProvider_CustomProviderStartsEmpty verifies that "custom" as the
// provider name starts from an empty preset (no built-in base).
func TestResolveProvider_CustomProviderStartsEmpty(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{
			Name:    "custom",
			Command: "my-agent",
			Args:    []string{"--custom-flag"},
		},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error for custom provider: %v", err)
	}
	if preset.Command != "my-agent" {
		t.Errorf("Command = %q, want %q", preset.Command, "my-agent")
	}
	// No built-in args should be inherited.
	if slices.Contains(preset.Args, "--dangerously-skip-permissions") {
		t.Error("custom provider inherited built-in opencode args; want clean slate")
	}
}

// TestResolveProvider_EmptyRepoNameReturnsTopLevel verifies that an empty repo
// name applies only the top-level provider config.
func TestResolveProvider_EmptyRepoNameReturnsTopLevel(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Command: "override-cmd"},
	}
	preset, err := cfg.ResolveProvider("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Command != "override-cmd" {
		t.Errorf("Command = %q, want %q", preset.Command, "override-cmd")
	}
}

// TestResolveProvider_NoProviderBlockUsesBuiltinFields verifies that the full
// built-in preset is returned when no provider config is present.
func TestResolveProvider_NoProviderBlockUsesBuiltinFields(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{{Name: "r", Cataractae: 1}},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// opencode built-in has ModelFlag "--model"
	if preset.ModelFlag != "--model" {
		t.Errorf("ModelFlag = %q, want %q", preset.ModelFlag, "--model")
	}
}

// TestResolveProvider_TopLevelModelAppliedToPreset verifies that model set at the
// top-level provider block is surfaced as DefaultModel on the resolved preset.
func TestResolveProvider_TopLevelModelAppliedToPreset(t *testing.T) {
	cfg := &AqueductConfig{
		Repos:    []RepoConfig{{Name: "r", Cataractae: 1}},
		Provider: &ProviderConfig{Model: "ollama/llama3"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.DefaultModel != "ollama/llama3" {
		t.Errorf("DefaultModel = %q, want %q", preset.DefaultModel, "ollama/llama3")
	}
}

// TestResolveProvider_RepoLevelModelOverridesTopLevelModel verifies that a
// repo-level model value takes precedence over the top-level model value.
func TestResolveProvider_RepoLevelModelOverridesTopLevelModel(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Model: "ollama/mistral"}},
		},
		Provider: &ProviderConfig{Model: "ollama/llama3"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.DefaultModel != "ollama/mistral" {
		t.Errorf("DefaultModel = %q, want %q (repo should override top-level)", preset.DefaultModel, "ollama/mistral")
	}
}

// TestResolveProvider_TopLevelCommandOverrideNotAppliedWhenRepoSpecifiesDifferentCommand
// verifies that when a repo-level config provides its own command, the top-level
// command override does not override it.
func TestResolveProvider_TopLevelCommandOverrideNotAppliedWhenRepoSpecifiesDifferentCommand(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Name: "opencode", Command: "/usr/local/bin/opencode"}},
		},
		Provider: &ProviderConfig{Name: "opencode", Command: "/opt/opencode"},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.Name != "opencode" {
		t.Errorf("Name = %q, want %q", preset.Name, "opencode")
	}
	if preset.Command != "/usr/local/bin/opencode" {
		t.Errorf("Command = %q, want %q (repo command override should win)", preset.Command, "/usr/local/bin/opencode")
	}
}

// TestResolveProvider_TopLevelGenericOverridesApplyRegardlessOfRepoProvider verifies
// that when the top-level config has no explicit name (generic overrides), its
// command/args/env/model overrides are applied even when the repo selects a named provider.
func TestResolveProvider_TopLevelGenericOverridesApplyRegardlessOfRepoProvider(t *testing.T) {
	cfg := &AqueductConfig{
		Repos: []RepoConfig{
			{Name: "r", Cataractae: 1, Provider: &ProviderConfig{Name: "opencode"}},
		},
		Provider: &ProviderConfig{Env: map[string]string{"GLOBAL_VAR": "val"}},
	}
	preset, err := cfg.ResolveProvider("r")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preset.ExtraEnv["GLOBAL_VAR"] != "val" {
		t.Errorf("ExtraEnv[GLOBAL_VAR] = %q, want %q (generic top-level env should apply)", preset.ExtraEnv["GLOBAL_VAR"], "val")
	}
}

// --- ValidateModelForProvider tests ---

// TestValidateModelForProvider_NoModelReturnsEmptyWarning verifies that a step
// with no model set produces no warning.
func TestValidateModelForProvider_NoModelReturnsEmptyWarning(t *testing.T) {
	step := WorkflowCataractae{Name: "s", Type: CataractaeTypeAgent}
	preset := provider.ProviderPreset{Name: "opencode", ModelFlag: "--model"}
	if warn := ValidateModelForProvider(step, preset); warn != "" {
		t.Errorf("expected empty warning, got %q", warn)
	}
}

// TestValidateModelForProvider_ModelWithModelFlagReturnsEmptyWarning verifies
// that a step with a model set against a preset that supports ModelFlag is fine.
func TestValidateModelForProvider_ModelWithModelFlagReturnsEmptyWarning(t *testing.T) {
	m := "ollama/llama3"
	step := WorkflowCataractae{Name: "s", Type: CataractaeTypeAgent, Model: &m}
	preset := provider.ProviderPreset{Name: "opencode", ModelFlag: "--model"}
	if warn := ValidateModelForProvider(step, preset); warn != "" {
		t.Errorf("expected empty warning, got %q", warn)
	}
}

// TestValidateModelForProvider_ModelWithoutModelFlagReturnsWarning verifies
// that setting a model against a preset with no ModelFlag produces a warning.
func TestValidateModelForProvider_ModelWithoutModelFlagReturnsWarning(t *testing.T) {
	m := "some-model"
	step := WorkflowCataractae{Name: "my-step", Type: CataractaeTypeAgent, Model: &m}
	preset := provider.ProviderPreset{Name: "custom", ModelFlag: ""}
	warn := ValidateModelForProvider(step, preset)
	if warn == "" {
		t.Fatal("expected warning, got empty string")
	}
	if !strings.Contains(warn, "my-step") {
		t.Errorf("warning %q should mention step name %q", warn, "my-step")
	}
	if !strings.Contains(warn, "custom") {
		t.Errorf("warning %q should mention provider name %q", warn, "custom")
	}
	if !strings.Contains(warn, "some-model") {
		t.Errorf("warning %q should mention model value %q", warn, "some-model")
	}
}
