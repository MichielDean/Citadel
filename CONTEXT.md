# Context

## Item: ci-wtopd

**Title:** Add ct droplet search: filter list by title, status, and priority
**Status:** in_progress
**Priority:** 2

### Description

Support --query flag with substring match on title, plus --status and --priority filters. Output respects existing color/icon format.

## Current Step: implement

- **Type:** agent
- **Role:** implementer
- **Context:** full_codebase

<available_skills>
  <skill>
    <name>cistern-droplet-state</name>
    <description>Manage droplet state in the Cistern agentic pipeline using the `ct` CLI.</description>
    <location>.claude/skills/cistern-droplet-state/SKILL.md</location>
  </skill>
  <skill>
    <name>github-workflow</name>
    <description>---</description>
    <location>.claude/skills/github-workflow/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-wtopd

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-wtopd
    ct droplet recirculate ci-wtopd --to implement

**Block (genuinely blocked, cannot proceed):**
    ct droplet block ci-wtopd

Add notes before signaling:
    ct droplet note ci-wtopd "What you did / found"

The `ct` binary is on your PATH.
