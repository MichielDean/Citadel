# Context

## Item: ci-xilu9

**Title:** Bug: in-repo skills not available in non-cistern repo sandboxes
**Status:** in_progress
**Priority:** 2

### Description

When aqueducts run against ScaledTest or PortfolioWebsite sandboxes, skills defined as in-repo paths (e.g. path: skills/cistern-droplet-state/SKILL.md) fail to copy because those paths only exist in the cistern repo, not in ScaledTest or PortfolioWebsite.

Root cause: internal/cataractae/runner.go resolves skill.Path relative to w.SandboxDir. For ScaledTest sandboxes, w.SandboxDir is ~/.cistern/sandboxes/ScaledTest/julia — which doesn't contain skills/cistern-droplet-state/.

Fix options:
1. Treat skills with path: as external skills if the path doesn't exist in the sandbox, falling back to ~/.cistern/skills/<name>/SKILL.md
2. When ct repo add creates new sandboxes, copy the cistern repo's skills/ directory into each new sandbox
3. Change aqueduct.yaml to use external skill references (no path:) for universal skills like cistern-droplet-state, and install them to ~/.cistern/skills/ during ct init

Option 1 is the most robust — graceful fallback means it works regardless of how the sandbox was created.

Affected: every ScaledTest and PortfolioWebsite aqueduct run logs 'warning: copy skill cistern-droplet-state' and agents run without their state management skill.

## Current Step: simplify

- **Type:** agent
- **Role:** simplifier
- **Context:** full_codebase

## Recent Step Notes

### From: scheduler

Implement pass rejected: HEAD has not advanced since last review (commit: 49299a429ca39f61358609292752e65b6909b32d). No new commits were found. You must commit your changes before signaling pass.

### From: manual

Committed existing fixes (os.UserHomeDir error check, shell-quoting skillsDir, buildClaudeCmd extraction + 5 tests in session_test.go). All 9 packages pass. HEAD advanced to d1b3ca6.

### From: manual

FINAL APPROACH: symlinks instead of copies.

Replace the file copy in runner.go with symlink creation:

1. internal/cataractae/runner.go — instead of copyFile(src, dest), use os.Symlink(src, dest):
   - For in-repo skills that don't exist in the sandbox, symlink from ~/.cistern/skills/<name>/SKILL.md
   - For external skills, symlink from ~/.cistern/skills/<name>/SKILL.md  
   - Check if symlink/file already exists before creating (idempotent)
   - If src doesn't exist in sandbox AND doesn't exist in ~/.cistern/skills/, log warning and skip

2. Result: ~/.cistern/skills/ is the single source of truth. Sandboxes contain symlinks pointing there. No copying, always current, deterministic.

3. ct init: seed ~/.cistern/skills/ with all built-in skills.
4. ct doctor: verify all aqueduct.yaml-referenced skills exist in ~/.cistern/skills/.
5. ct repo add: no sandbox seeding needed — symlinks are created at job start by the runner.

The <location> in CONTEXT.md stays as .claude/skills/<name>/SKILL.md (relative to sandbox). Claude reads the symlink transparently.

### From: manual

All three review issues resolved and committed (d1b3ca6): (1) os.UserHomeDir() error checked — spawn returns error if HOME unset; (2) skillsDir shell-quoted via shellQuote() helper; (3) buildClaudeCmd() extracted as testable method with 5 tests in session_test.go. All 9 packages pass.

<available_skills>
  <skill>
    <name>cistern-droplet-state</name>
    <description>Manage droplet state in the Cistern agentic pipeline using the `ct` CLI.</description>
    <location>.claude/skills/cistern-droplet-state/SKILL.md</location>
  </skill>
  <skill>
    <name>code-simplifier</name>
    <description>code-simplifier</description>
    <location>.claude/skills/code-simplifier/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-xilu9

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-xilu9
    ct droplet recirculate ci-xilu9 --to implement

**Block (genuinely blocked, cannot proceed):**
    ct droplet block ci-xilu9

Add notes before signaling:
    ct droplet note ci-xilu9 "What you did / found"

The `ct` binary is on your PATH.
