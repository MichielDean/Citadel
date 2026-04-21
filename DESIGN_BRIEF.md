# Design Brief: Clean up tests, docs, and skills for opencode-only

## Requirements Summary

Remove all remaining references to claude, codex, gemini, copilot, ANTHROPIC_API_KEY, CLAUDE.md, GEMINI.md, and `.claude/` from tests, docs, skills, and config files across the codebase. The Go core code has already been migrated to opencode-only (previous droplet). This brief covers the leftover references in test fixtures, documentation, config examples, and type comments. Goal: no references to non-opencode providers, non-opencode API keys, or non-AGENTS.md instruction filenames remain anywhere in the codebase.

## Existing Patterns to Follow

### Naming Conventions

- Table-driven tests use `struct{name string; ...}` with `t.Run(tt.name, ...)`. See `internal/provider/preset_test.go:11-28` (`TestBuiltins_ReturnsExpectedPresetNames`), `internal/cataractae/smoke_test.go:18-61`.
- YAML config uses snake_case keys. See `internal/aqueduct/types.go:190` (`llm:` key), `cmd/ct/assets/cistern.yaml`.
- Environment variables use `UPPER_SNAKE_CASE`. See `cmd/ct/init_test.go:495` (`ANTHROPIC_API_KEY`).

### Error Handling

- Test assertions use `t.Errorf` / `t.Fatalf` with descriptive messages. See `cmd/ct/init_test.go:497-499` (`t.Errorf("ct init next-steps message must not mention %s; output:\n%s", envVar, output)`).
- Operational errors use `fmt.Errorf("pkg: context: %w", err)`. See `internal/aqueduct/provider.go:50`.

### Testing

- `t.Setenv` for environment variable tests. See `internal/cataractae/smoke_test.go:22`.
- `t.Helper()` for test helpers. See `internal/cataractae/smoke_test.go:141`.
- Table-driven tests with `t.Run`. See `cmd/ct/doctor_test.go:1700-1713` (`TestInferLLMProviderFromPreset_KnownPresets`).
- Mock HTTP server for LLM endpoint testing. See `internal/testutil/mockllm/mockllm.go`.

### Idiom Fit

- Mock LLM server uses `httptest.NewServer`. See `internal/testutil/mockllm/mockllm.go:62-68`. This is idiomatic Go — no custom HTTP mocking needed.
- Package-level function variables for test injectability. See `internal/cataractae/session.go:238`.

## Reusability Requirements

- `LLMConfig.Provider` at `internal/aqueduct/types.go:154-155` is the LLM API provider for filtration/refinement, separate from the agent CLI provider. It remains configurable (anthropic, openai, openrouter, ollama, custom) because the LLM API is orthogonal to the agent CLI — opencode-agent uses an LLM backend but can use different LLM APIs. **Do not remove the "anthropic" value from the LLM config** — it is a legitimate LLM API provider name, not an agent CLI provider. Only update the default comment and behavior.
- `mockllm.go` endpoints (`/v1/messages` for Anthropic, `/v1/chat/completions` for OpenAI-compatible) are kept because the LLM refinement feature (`ct filter`) supports multiple LLM backends. **Keep both endpoints**. Only update the comments to use neutral language.

## Coupling Requirements

- The `mockllm` test utility at `internal/testutil/mockllm/mockllm.go` is shared by `cmd/ct/refine_test.go` and `internal/evaluate/*_test.go`. Changes to its comments or test fixture variable names must be consistent across callers.
- The `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY` list in `cmd/ct/init_test.go:495` is a test-only constant — not shared with production code.

## DRY Requirements

No repeated patterns are introduced. The cleanup removes stale references — it does not add new code.

The `removedEnvVars` list at `cmd/ct/init_test.go:495` contains three provider-specific env vars (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`). These are only referenced in one test function. **Keep the list but evaluate each entry**:
- `ANTHROPIC_API_KEY` — not applicable for opencode-only. **Keep in the list** (the test checks that ct init output doesn't mention removed env vars).
- `OPENAI_API_KEY` — not applicable for opencode-only default. **Keep in the list** (same reason).
- `GEMINI_API_KEY` — not applicable for opencode-only default. **Keep in the list** (same reason).

The test concept is sound: verify that ct init doesn't suggest setting removed provider env vars. The list itself is the reference data, not dead code.

## Migration Requirements

No database migrations. This is purely config, documentation, and test cleanup — removing stale references, not adding tables or columns.

However, there is a **config semantics concern**: `LLMConfig.Provider` at `internal/aqueduct/types.go:152-156` currently documents `"anthropic"` as the default LLM provider. After this change, the comment at line 155 should say `"When omitted, defaults to 'anthropic'"` remains correct because the LLM API is independent of the agent CLI — but the comment at line 190 about "default anthropic preset" should be clarified to distinguish the LLM API from the agent provider.

## Test Requirements

### Test naming convention
Tests use `TestXxx_YyyZzz` format. See `cmd/ct/init_test.go:493` (`TestInit_NextStepsMessage_DoesNotMentionRemovedProviderEnvVars`).

### Specific test updates required

| File | Change |
|------|--------|
| `cmd/ct/init_test.go:493-501` | No change needed to the test function. The `removedEnvVars` list is correct — it tests that ct init output doesn't mention provider env vars that no longer apply to the default provider. Verify the test passes after other changes. |
| `cmd/ct/doctor_test.go:1021` | Update comment from `provider=opencode but llm.provider=anthropic` to `provider=opencode but llm.provider=ollama` (or keep as-is since the test verifies a mismatch scenario — anthropic LLM is a valid mismatch). **Decision**: keep the test as-is. The test verifies that an opencode agent + anthropic LLM mismatch produces an advisory, not a hard failure. This is a valid test scenario since LLM API providers are separate from agent CLI providers. |
| `cmd/ct/refine_test.go:50` | The `anthropic` entry is a valid LLM API provider name for the mock server test. **Keep** — this tests the `/v1/messages` endpoint which is the Anthropic Messages API format, not the claude CLI agent. |

## Forbidden Patterns

- Removing LLM API provider names (anthropic, openai, openrouter, ollama, custom) from the `LLMConfig` — these are LLM backends for filtration, not agent CLI providers. The agent CLI is always opencode; the LLM API is independently configurable.
- Removing the `/v1/messages` endpoint from mockllm — it is needed to test Anthropic-format LLM API calls which are a legitimate backend for `ct filter`.
- Removing `GH_TOKEN` from env passthrough or test fixtures — GitHub token is still required for delivery regardless of agent provider.

## API Surface Checklist

### `.env.example`

- [ ] Line 8-12: Remove the `ANTHROPIC_API_KEY=your_anthropic_api_key_here` line and the section comment "Required for headless / non-interactive deployments". Replace with a comment about opencode provider auth and `GH_TOKEN`.
- [ ] Line 18: Remove `# CLAUDE_PATH=/usr/local/bin/claude` comment. Opencode doesn't use a `CLAUDE_PATH` env var.

### `CHANGELOG.md`

- [ ] Add a new top-level entry documenting the opencode-only migration. Entry should state: "Cistern now uses opencode as the only agent CLI provider. The claude, codex, gemini, and copilot provider presets have been removed. ANTHROPIC_API_KEY is no longer required. CLAUDE.md instruction files have been replaced by AGENTS.md across the codebase."

### `internal/aqueduct/types.go`

- [ ] Line 154-155: Update `LLMConfig.Provider` comment. The known values list (`"anthropic", "openai", "openrouter", "ollama", "custom"`) is correct — these are LLM API providers, not agent providers. The default comment `"When omitted, defaults to 'anthropic'"` should be updated to `"When omitted, defaults to 'anthropic'"` — this is the LLM backend default, not the agent default, so "anthropic" is still a reasonable default for the LLM API. **No change needed to the value**. However, update the phrase "default anthropic preset" at line 190 to "default LLM provider" to avoid confusion with agent provider presets.
- [ ] Line 190: Change `"the default anthropic preset is used"` to `"the default LLM provider is used (anthropic)"`. This clarifies the distinction between LLM API provider and agent CLI provider.

### `cmd/ct/init_test.go`

- [ ] Lines 493-501: Verify `TestInit_NextStepsMessage_DoesNotMentionRemovedProviderEnvVars` still passes. The `removedEnvVars` list (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`) is correct — these are removed provider env vars that ct init should not suggest setting. **No source change needed**, just verify the test passes.

### `internal/testutil/mockllm/mockllm.go`

- [ ] Lines 5-7: Update doc comment. Change `"Anthropic Messages API format"` to `"Anthropic Messages API format (for the LLM backend used by ct filter, not the agent CLI)"`. This clarifies that Anthropic here refers to the LLM API backend, not the claude CLI agent.
- [ ] Line 19: Change `t.Setenv("ANTHROPIC_BASE_URL", mock.URL)` comment example to `t.Setenv("ANTHROPIC_BASE_URL", mock.URL) // or OPENAI_BASE_URL for OpenAI-compatible providers`. This shows both LLM backend options.
- [ ] Line 87: Update the `handleMessages` comment from `"Anthropic"` to `"Anthropic Messages API"`. This is a technical API name, not a product reference.

### `cmd/ct/refine_test.go`

- [ ] Line 50: The `{name: "anthropic", endpoint: "/v1/messages"}` test case is correct — it tests the Anthropic LLM API endpoint format. **Keep as-is**. Add a comment above line 45: `// These test the LLM API mock server endpoints (for ct filter), not the agent CLI provider.`

### `cmd/ct/doctor_test.go`

- [ ] Line 1021: Update comment from `agent provider=opencode but llm.provider=anthropic` to `agent provider=opencode but llm.provider=anthropic (LLM API mismatch, not agent CLI)`. Clarifies that "anthropic" is the LLM API provider, not the agent.
- [ ] Line 1031: `provider: anthropic` in YAML is the LLM provider config, not the agent CLI. **Keep as-is** — this is a valid test scenario (opencode agent + anthropic LLM API).
- [ ] Line 1654: Update comment from `opencode agent + anthropic LLM` to `opencode agent + anthropic LLM API`. Makes the LLM vs agent distinction clear.

### `DESIGN_BRIEF.md`

- [ ] Replace this entire file with the completed brief (i.e., this document supersedes the previous "Remove all non-opencode providers from core" brief which covered the Go source migration that has already been implemented).

## Non-Go Files Already Clean (Verified)

The following files were listed in the CONTEXT.md requirements but are already clean after the previous droplet's Go migration. No changes needed:

- `README.md` — already documents opencode as the only provider. No CLAUDE.md, GEMINI.md, or multi-provider references found.
- `.gitignore` — no `.claude/` entry found.
- `docker-entrypoint.sh` — no `~/.claude` symlink or ANTHROPIC_API_KEY references.
- `cataractae/delivery/INSTRUCTIONS.md` — no multi-provider InstructionsFile references; line 194 correctly says `AGENTS.md`.
- `skills/cistern/SKILL.md` — no `~/.claude/.credentials.json` or ANTHROPIC_API_KEY references.
- `skills/cistern/references/commands.md` — no CLAUDE.md or GEMINI.md references.
- `skills/cistern/references/setup.md` — no multi-provider references; documents opencode-first setup.
- `skills/cistern/references/troubleshooting.md` — no multi-provider references; mentions "provider credentials" generically.
- `internal/skills/cistern-git/SKILL.md` — correctly says "AGENTS.md for opencode".
- `ARCHITECTURE.md` — about web UI, no multi-provider references.