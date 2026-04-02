# Context

## Item: ci-besty

**Title:** TUI cockpit: Castellarius control panel with start/stop/restart
**Status:** in_progress
**Priority:** 2

### Description

Panel registered as module 4 (key: 4) showing ct castellarius status output with live-refresh (5s ticker). Exposes start, stop, restart actions via command palette with confirm overlays. Maps to ct castellarius start/stop/restart. Acceptance: pressing 4 shows castellarius status; palette actions control the service; status refreshes after actions.

## Current Step: implement

- **Type:** agent
- **Role:** implementer
- **Context:** full_codebase

<available_skills>
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
  <skill>
    <name>cistern-github</name>
    <description>Use `gh` CLI for all GitHub operations. Prefer CLI over GitHub MCP servers for lower context usage.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-github/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-besty

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-besty
    ct droplet recirculate ci-besty --to implement

**Pool (cannot currently proceed):**
    ct droplet pool ci-besty

Add notes before signaling:
    ct droplet note ci-besty "What you did / found"

The `ct` binary is on your PATH.
