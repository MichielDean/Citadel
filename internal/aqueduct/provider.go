package aqueduct

import (
	"fmt"
	"maps"
	"slices"

	"github.com/MichielDean/cistern/internal/provider"
)

// ResolveProvider returns the effective ProviderPreset for the named repo.
//
// Resolution order: built-in preset → top-level AqueductConfig.Provider overrides
// → repo-specific RepoConfig.Provider overrides.
//
// The preset name is resolved with repo-level taking precedence over top-level,
// which takes precedence over the default ("claude"). The special name "custom"
// starts from an empty preset instead of a built-in.
//
// If repoName does not match any configured repo, only the top-level provider
// config is applied. An empty repoName behaves the same way.
//
// An error is returned if the resolved preset name is not a known built-in and
// is not "custom".
func (cfg *AqueductConfig) ResolveProvider(repoName string) (provider.ProviderPreset, error) {
	// Determine effective preset name: repo level > top level > default.
	name := "claude"
	if cfg.Provider != nil && cfg.Provider.Name != "" {
		name = cfg.Provider.Name
	}

	var repoProvider *ProviderConfig
	if idx := slices.IndexFunc(cfg.Repos, func(r RepoConfig) bool { return r.Name == repoName }); idx >= 0 {
		repoProvider = cfg.Repos[idx].Provider
	}
	if repoProvider != nil && repoProvider.Name != "" {
		name = repoProvider.Name
	}

	// Resolve the base preset.
	var result provider.ProviderPreset
	if name == "custom" {
		result = provider.ProviderPreset{Name: "custom"}
	} else {
		builtins := provider.Builtins()
		idx := slices.IndexFunc(builtins, func(p provider.ProviderPreset) bool {
			return p.Name == name
		})
		if idx < 0 {
			return provider.ProviderPreset{}, fmt.Errorf("aqueduct: unknown provider %q", name)
		}
		result = builtins[idx]
	}

	// Apply top-level overrides onto the base preset, but only when the
	// top-level provider name matches the resolved name. If the top-level has
	// no explicit name its overrides are treated as generic and always applied;
	// if it names a specific provider (e.g. "gemini") but the repo has switched
	// to a different one (e.g. "codex"), applying those overrides would silently
	// contaminate the repo's preset with the wrong provider's settings.
	if cfg.Provider != nil && (cfg.Provider.Name == "" || cfg.Provider.Name == name) {
		applyProviderOverrides(&result, cfg.Provider)
	}

	// Apply repo-level overrides on top.
	if repoProvider != nil {
		applyProviderOverrides(&result, repoProvider)
	}

	return result, nil
}

// applyProviderOverrides applies non-zero fields from cfg onto the preset p.
//   - A non-empty Command replaces p.Command.
//   - Args are appended to p.Args.
//   - Env entries are merged into p.ExtraEnv; later calls override earlier ones.
//
// The Name field is intentionally not applied here — it is resolved before this
// function is called.
func applyProviderOverrides(p *provider.ProviderPreset, cfg *ProviderConfig) {
	if cfg.Command != "" {
		p.Command = cfg.Command
	}
	if len(cfg.Args) > 0 {
		p.Args = append(p.Args, cfg.Args...)
	}
	if len(cfg.Env) > 0 {
		if p.ExtraEnv == nil {
			p.ExtraEnv = make(map[string]string, len(cfg.Env))
		}
		maps.Copy(p.ExtraEnv, cfg.Env)
	}
	if cfg.Model != "" {
		p.DefaultModel = cfg.Model
	}
}

// ValidateModelForProvider checks whether a workflow cataractae's Model field can
// be used with the given provider preset.
//
// If the preset has no ModelFlag, the model value cannot be passed to the agent
// CLI and will be silently ignored at launch time. This function returns a
// descriptive warning string in that case so callers can surface it to the user.
// An empty return value means no issue was found.
func ValidateModelForProvider(step WorkflowCataractae, preset provider.ProviderPreset) string {
	if step.Model == nil || preset.ModelFlag != "" {
		return ""
	}
	return fmt.Sprintf(
		"cataractae %q: provider %q has no model flag; model %q will be ignored",
		step.Name, preset.Name, *step.Model,
	)
}
