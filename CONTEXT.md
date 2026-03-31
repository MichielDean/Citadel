# Context

## Item: ci-pay2p

**Title:** Architecti snapshot: append Cistern SKILL.md as final context section
**Status:** in_progress
**Priority:** 2

### Description

In buildArchitectiSnapshot() (internal/castellarius/architecti.go), read openclaw/cistern/SKILL.md from the cistern primary clone at filepath.Join(s.sandboxRoot, "cistern", "_primary", "openclaw", "cistern", "SKILL.md") and append its content verbatim under a '## Cistern Reference' heading as the final section of every snapshot.

Acceptance criteria:
- Skill file content appears verbatim under '## Cistern Reference' in every snapshot, regardless of which repo triggered Architecti
- If the file is unreadable (missing, permission error, etc.), append a single fallback line (e.g. '## Cistern Reference\n(skill file unavailable)') and continue — snapshot must not fail
- No other snapshot sections are modified; dispatch logic in architecti.go is untouched
- Unit tests cover both paths: skill-present (heading + content appear in snapshot) and skill-missing (fallback line appears, no error returned)

## Current Step: docs

- **Type:** agent
- **Role:** docs_writer
- **Context:** full_codebase

## ⚠️ REVISION REQUIRED — Fix these issues before anything else

This droplet was recirculated. The following issues were found and **must** be fixed.
Do not proceed to implementation until you have read and understood each issue.

### Issue 1 (from: security)

No security issues found. The 9-line change reads a hardcoded-path documentation file (SKILL.md) from the project's own primary clone and appends it to an LLM snapshot. Path is fully composed of string literals (no user input), error handling falls back to a safe static string without leaking OS error details, and the content stays within the same trust boundary as all other snapshot data. No injection, path traversal, data exposure, or resource safety concerns.

---

## Recent Step Notes

### From: security

No security issues found. The 9-line change reads a hardcoded-path documentation file (SKILL.md) from the project's own primary clone and appends it to an LLM snapshot. Path is fully composed of string literals (no user input), error handling falls back to a safe static string without leaking OS error details, and the content stays within the same trust boundary as all other snapshot data. No injection, path traversal, data exposure, or resource safety concerns.

### From: qa

All tests pass (full suite clean). Implementation: clean 9-line append at end of buildArchitectiSnapshot, follows existing section-separator pattern, correct fallback on any read error. Tests: both paths covered with real filesystem I/O (not mocked), descriptive names, multiple assertions including negative (no fallback text when file present). All five acceptance criteria verified: verbatim content, fallback on error, snapshot continues on failure, no other sections modified, dispatch logic untouched.

### From: reviewer

No findings. Clean 9-line implementation follows existing section-separator pattern. Error handling correct (fallback on read failure). Tests cover both paths (skill-present, skill-missing). No security, logic, concurrency, or resource leak issues. All package tests pass.

### From: simplifier

No simplifications required. The 9-line implementation in architecti.go follows the file's existing section-separator pattern exactly (// --- Name --- comment + heading + logic). Tests are clean and well-structured. Code is already clear and idiomatic.

<available_skills>
  <skill>
    <name>cistern-droplet-state</name>
    <description>Manage droplet state in the Cistern agentic pipeline using the `ct` CLI.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-droplet-state/SKILL.md</location>
  </skill>
  <skill>
    <name>cistern-git</name>
    <description>---</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-git/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-pay2p

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-pay2p
    ct droplet recirculate ci-pay2p --to implement

**Block (genuinely blocked, cannot proceed):**
    ct droplet block ci-pay2p

Add notes before signaling:
    ct droplet note ci-pay2p "What you did / found"

The `ct` binary is on your PATH.
