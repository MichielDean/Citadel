# Context

## Item: ci-7x196

**Title:** Detect agents that exit inside a live tmux session without signaling outcome
**Status:** in_progress
**Priority:** 2

### Description

The current `isTmuxAlive` check in `scheduler.go` calls `tmux has-session` — this returns true even when the tmux session is alive but the claude process inside it has already exited (e.g. OOM kill, hard token limit, or non-zero exit), leaving a shell prompt in the pane. These sessions are invisible to the heartbeat and stay in_progress indefinitely until the session eventually closes. Fix: extend `isTmuxAlive` (or add a companion `isAgentAlive(sessionID string) bool`) that uses `tmux list-panes -t <session> -F '#{pane_pid}'` to get the pane's root PID, then checks whether any child process matching `claude` is still alive under that PID (e.g. `kill -0` or `/proc/<pid>/cmdline`). In `heartbeatRepo`, when `isTmuxAlive` is true but `isAgentAlive` is false and there is no outcome, kill the orphaned tmux session (`tmux kill-session`), add a diagnostic note (`"Session zombie detected: tmux alive but claude process dead. Session killed. Re-dispatching. [<timestamp>]"`) via `client.AddNote`, then reset the droplet to open. Acceptance: (1) `isAgentAlive` correctly distinguishes a live claude process from a shell-only pane in unit tests with mock tmux output; (2) heartbeat integration test covers the tmux-alive-process-dead path: note is written and droplet is reset; (3) existing path (tmux fully dead) is unchanged and still covered by tests from the prior droplet.

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
    <description>---</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-git/SKILL.md</location>
  </skill>
  <skill>
    <name>cistern-github</name>
    <description>---</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-github/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-7x196

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-7x196
    ct droplet recirculate ci-7x196 --to implement

**Pool (cannot currently proceed):**
    ct droplet pool ci-7x196

Add notes before signaling:
    ct droplet note ci-7x196 "What you did / found"

The `ct` binary is on your PATH.
