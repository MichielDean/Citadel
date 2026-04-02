# Context

## Item: ci-cv5jf

**Title:** Remove ct audit command
**Status:** in_progress
**Priority:** 2

### Description

Delete the ct audit subcommand entirely. It provides little value: the audit agent hallucinates findings for wrong codebases, files them under incorrect prefixes, and the results require manual triage that defeats the purpose.

Files to delete:
- cmd/ct/audit.go — command implementation
- cmd/ct/audit_test.go — all tests
- internal/testutil/fakeauditagent/main.go and its directory (only used by audit_test.go; failagent stays, it is also used by filter_test.go)

Documentation to update:
- openclaw/cistern/references/commands.md — remove the ct audit section (lines ~341-382)

No other Go source files reference auditCmd or AuditFinding outside audit.go and audit_test.go. The --cancelled flag comment in cistern.go and the skills.go example string are unrelated and must not be touched.

Acceptance criteria:
- ct audit run returns 'unknown command' or is entirely absent from ct help
- All tests pass (go test ./...)
- No dangling references to AuditFinding, auditCmd, auditRunCmd, auditSystemPrompt, or fakeauditagent remain in non-deleted files
- commands.md no longer documents ct audit

## Current Step: delivery

- **Type:** agent
- **Role:** delivery

## ⚠️ REVISION REQUIRED — Fix these issues before anything else

This droplet was recirculated. The following issues were found and **must** be fixed.
Do not proceed to implementation until you have read and understood each issue.

### Issue 1 (from: security)

No security issues found. Diff is entirely deletions (audit.go, audit_test.go, fakeauditagent) plus documentation and config updates. No new code paths, input surfaces, auth checks, or injection vectors introduced. No dangling references to deleted symbols. SKILL.md update improves security posture by warning against ANTHROPIC_API_KEY misuse.

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
    <description>Each droplet has an isolated worktree at `~/.cistern/sandboxes/&lt;repo&gt;/&lt;droplet-id&gt;/`.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-git/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-cv5jf

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-cv5jf
    ct droplet recirculate ci-cv5jf --to implement

**Pool (cannot currently proceed):**
    ct droplet pool ci-cv5jf

Add notes before signaling:
    ct droplet note ci-cv5jf "What you did / found"

The `ct` binary is on your PATH.
