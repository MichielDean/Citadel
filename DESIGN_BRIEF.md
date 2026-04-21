# Design Brief: Remove non-opencode providers from commands

## Requirements Summary

The primary removal of non-opencode providers (claude, ollama-claude, codex, gemini, copilot) from the Go codebase was completed in commit 277b499. The remaining work is cleaning up residual references in non-Go files, test assertions about removed providers, and documentation that still mentions removed providers. These are: `.env.example` still lists `ANTHROPIC_API_KEY` and `CLAUDE_PATH`; test code in `cmd/ct/init_test.go` still asserts exclusion of `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GEMINI_API_KEY` from init output; the AGENTS.md instructions file (injected by the Castellarius into cataractae worktrees) still mentions "CLAUDE.md for claude, GEMINI.md for gemini, AGENTS.md for opencode/codex"; the cistern-skill SKILL.md still mentions "CLAUDE.md for claude"; and the `agentJSONOutput` struct in filter.go is opencode's actual JSON output format (not claude-specific) and should be retained.

## Existing Patterns to Follow

### Naming Conventions

- Exported functions use PascalCase; unexported structs use lowercase field names. See `internal/provider/preset.go:26` (`ProviderPreset` exported, all fields PascalCase because the struct is exported).
- Error wrapping uses `fmt.Errorf("pkg: context: %w", err)`. See `internal/provider/preset.go:171` (`fmt.Errorf("provider: read %s: %w", path, err)`).

### Error Handling

- Errors wrapped with domain context using `fmt.Errorf("pkg: context: %w", err)`.
- Operational errors use `slog.Error`/`slog.Warn`. See `internal/cataractae/session.go:321-322`.

### Testing

- Table-driven tests using `t.Run(name, func(t *testing.T) {...})`. See `internal/provider/preset_test.go:43-55`.
- Test helpers use `t.Helper()`. See `internal/provider/preset_test.go:348`.
- `t.Setenv` for environment variable tests. See `cmd/ct/doctor_test.go:1460`.

### Idiom Fit

- Constructor-style initialization via `NewXxx` functions. See `internal/evaluate/evaluate.go:99` (`NewLLMCaller`).
- Package-level function variables for test injectability. See `internal/cataractae/session.go:212` (`resolveCommandFn`).

## Reusability Requirements

- No new types or abstractions needed. All changes are removals or simplifications of existing code.
- The `agentJSONOutput` struct in `cmd/ct/filter.go:19-25` is **opencode's actual output format** (opencode uses `--output-format json` with `type`, `subtype`, `is_error`, `result`, `session_id` fields). It is NOT claude-specific. Keep it.

## Coupling Requirements

No new shared mutable state is introduced. This change removes residual references only.

## DRY Requirements

No new repeated patterns are introduced.

## Migration Requirements

No database migrations. No file system migration concerns — all code-level references have already been migrated to AGENTS.md by the prior commit.

## Test Requirements

### Test files requiring updates

| File | Change |
|------|--------|
| `cmd/ct/init_test.go:495` | `TestInit_NextStepsMessage_DoesNotMentionRemovedProviderEnvVars` — `removedEnvVars` list includes `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`. These are now provider-agnostic env vars (opencode may need ANTHROPIC_API_KEY via its own config). Rename the test to `TestInit_NextStepsMessage_DoesNotMentionRemovedProviders` and remove these three vars from the `removedEnvVars` list since they may still be relevant for opencode's LLM backend configuration. Alternatively, keep the test but change the assertion to verify that `ct init` output does not mention "claude", "codex", "gemini", or "copilot" as provider names — not specific env var names. |

## Forbidden Patterns

- No new code patterns introduced. Only removals and comment simplifications.
- Mixed `AGENTS.md`/`CLAUDE.md` references in the same code path — after this change, only `AGENTS.md` should appear in active code. `CHANGELOG.md` historical entries are exempt.

## API Surface Checklist

### Config Files

- [ ] `.env.example:12` — Remove `ANTHROPIC_API_KEY=your_anthropic_api_key_here` line. Opencode uses ollama or other LLM backends, not Anthropic API keys directly. Replace with a comment about opencode's model configuration if appropriate.
- [ ] `.env.example:18` — Remove `# CLAUDE_PATH=/usr/local/bin/claude` line. The opencode binary is installed via `go install` (as documented in doctor's `providerInstallHint`).

### AGENTS.md Instructions (Cataractae Instructions File)

- [ ] `cataractae/architect/INSTRUCTIONS.md` (line 168 of injected AGENTS.md, but this is in the Castellarius's own source files) — The `cistern-git` skill at `skills/cistern-git/SKILL.md` lines 16-17 and 43 mentions "AGENTS.md for opencode/codex, CLAUDE.md for claude". Simplify to just `"AGENTS.md"`.
- [ ] `skills/cistern/SKILL.md` — Search for any `~/.claude/.credentials.json` reference or CLAUDE.md mention and remove it.

### Test Code

- [x] `cmd/ct/init_test.go:493` — `TestInit_NextStepsMessage_DoesNotMentionRemovedProviders` now checks provider names (claude, codex, gemini, CLAUDE_PATH) instead of env var names (ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY) that may still be relevant for opencode's LLM backend.

### No-Op Items (Already Implemented)

The following items from the prior brief were completed in commit 277b499. They are listed here to confirm verification, not to require new implementation:

- [x] `builtins` var contains exactly one entry: opencode preset
- [x] `InstrFile()` returns `"AGENTS.md"` when empty
- [x] `ResolvePreset("")` falls back to opencode
- [x] `NonInteractiveConfig.PrintFlag` removed
- [x] `NonInteractiveConfig.AllowedToolsFlag` removed
- [x] `ResumeStyle` type and constants removed
- [x] `ContinueFlag` field removed
- [x] `AddDirFlag` field removed
- [x] `SupportsAddDir` field removed
- [x] `buildClaudeCmd` deleted from session.go
- [x] `claudePathFn` and `claudePath` deleted from session.go
- [x] `AgentAliveUnderPIDIn` renamed (was `ClaudeAliveUnderPIDIn`)
- [x] `IsAgentCmdline` only checks for `"opencode"`
- [x] `IsClaudeCmdline` deleted
- [x] `ensureCataractaeIntegrity` checks for AGENTS.md
- [x] `drought_hooks.go` uses AGENTS.md
- [x] `context.go` InstructionsFile comment simplified
- [x] `provider.go` defaults to opencode
- [x] `parse.go` defaults instructionsFile to AGENTS.md
- [x] `providerInstallHint` only has opencode case
- [x] `inferLLMProviderFromPreset` only has opencode→ollama case
- [x] `checkInstructionsFileIntegrity` renamed (was `checkClaudeMdIntegrity`)
- [x] `castellarius.go` `startupRequiredEnvVars` has no `usesClaude`
- [x] `fakeclaude/` directory deleted
- [x] `fakeagent/` comments cleaned up
- [x] `.gitignore` no longer has `.claude/` entry
- [x] `docker-entrypoint.sh` no longer has `~/.claude` symlink setup
- [x] `LLMCaller` PrintFlag and AllowedToolsFlag removed
- [x] `NewLLMCaller` PrintFlag and AllowedToolsFlag parameters removed

## What the Brief Is NOT

- It is NOT a full implementation. Do not write production code.
- It is NOT a test file. Do not write test cases.
- It IS a contract document that the implementer must satisfy and the reviewer must verify.