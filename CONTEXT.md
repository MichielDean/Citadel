# Context

## Item: ci-o3yxk

**Title:** TUI cockpit: Filter conversation panel
**Status:** in_progress
**Priority:** 2

### Description

Panel registered as module 8 (key: 8). Interactive multi-turn filtration conversation inside the TUI. Message history displayed as alternating user/LLM blocks. Text input at the bottom; enter submits. First-use: single text box, first line = title, rest = description. Session ID maintained for --resume across turns. n starts a new session clearing history. Submit → spinner → display full response (no streaming required). Manages subprocess invocation currently in callFilterAgent() (filter.go) from within the TUI via tea.ExecProcess or a goroutine with tea.Msg callbacks. Acceptance: pressing 8 opens filter panel; user can have a multi-turn filtration conversation without leaving the TUI.

## Current Step: delivery

- **Type:** agent
- **Role:** delivery

## ⚠️ REVISION REQUIRED — Fix these issues before anything else

This droplet was recirculated. The following issues were found and **must** be fixed.
Do not proceed to implementation until you have read and understood each issue.

### Issue 1 (from: security)

Security review of filter panel TUI (ci-o3yxk). Audited: filter_panel_tui.go (new), cockpit_tui.go (new), filter.go (pre-existing), filter_context.go (pre-existing), doctor_panel_tui.go (new), status_panel_tui.go (new), dashboard_web.go (modified).

Primary concern — command injection via user input in filter panel:
- Keyboard input flows: handleKey() → inputBuf → invokeFilterNew/Resume → callFilterAgent → exec.CommandContext(ctx, preset.Command, args...). No shell is involved; exec.CommandContext passes the prompt as a single argv element. No injection vector.
- Session ID from agent JSON response is similarly passed as argv elements to exec.CommandContext. Even with adversarially-crafted session IDs, no shell interpretation occurs.

Context gathering (gatherFilterContext):
- Title/desc are used only for strings.Contains matching to select cobra subcommand help — never interpolated into SQL or shell commands.
- DB schema query (gatherDBSchema) uses sqlite_master with no user-supplied values in the query string.

Doctor panel: uses os.Executable() for binary path — no user input reaches the command path.

Dashboard web changes: limited to attach() snapshot assembly and resize() frame-flush sequencing. No new attack surface. Origin validation for WebSocket (isAllowedWSOrigin) is unchanged and correctly validates RFC 1918 + loopback.

Error exposure: error messages displayed only in the local TUI; no external disclosure.

No blocking or required issues found.

### Issue 2 (from: security)

No blocking or required security issues. Full input-to-subprocess trace confirms no command injection: user keyboard input and agent session IDs are both passed as argv elements via exec.CommandContext (no shell). Context gathering does not interpolate user input into SQL or shell commands. Dashboard web changes are internal rendering logic with no new attack surface.

---

<available_skills>
  <skill>
    <name>cistern-github</name>
    <description>Use `gh` CLI for all GitHub operations. Prefer CLI over GitHub MCP servers for lower context usage.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-github/SKILL.md</location>
  </skill>
  <skill>
    <name>cistern-droplet-state</name>
    <description>Manage droplet state in the Cistern agentic pipeline using the `ct` CLI.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-droplet-state/SKILL.md</location>
  </skill>
  <skill>
    <name>cistern-git</name>
    <description>Each droplet has an isolated worktree at `~/.cistern/sandboxes/<repo>/<droplet-id>/`.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-git/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-o3yxk

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-o3yxk
    ct droplet recirculate ci-o3yxk --to implement

**Pool (cannot currently proceed):**
    ct droplet pool ci-o3yxk

Add notes before signaling:
    ct droplet note ci-o3yxk "What you did / found"

The `ct` binary is on your PATH.
