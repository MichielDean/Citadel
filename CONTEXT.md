# Context

## Item: ci-8hhrs

**Title:** Castellarius: detect and recover stuck delivery agents
**Status:** in_progress
**Priority:** 1

### Description

A delivery agent can get stuck in a CI-polling loop even after all checks pass. This happened with sc-zy2y2: the agent polled for 67+ minutes while PR #164 had been fully green and mergeable, but the branch fell behind main after other PRs merged. The agent never detected the branch-behind condition and never attempted a rebase.

The Castellarius should detect this and recover automatically.

## Detection criteria
A delivery cataractae is considered stuck when ALL of the following are true:
- Droplet has been in 'delivery' stage for longer than the delivery timeout_minutes (currently 45m)
- The agent process is still alive (tmux session exists, process running)
- The associated PR exists and is in a recoverable state (OPEN, not CLOSED/MERGED)

## Recovery protocol
When a stuck delivery is detected, the Castellarius should:
1. Kill the stuck agent process (send SIGTERM to the tmux pane)
2. Fetch the PR URL from GitHub for the droplet's branch (search by droplet ID or branch pattern feat/<droplet-id>)
3. Check PR state:
   a. If MERGED: signal pass — the work is done, agent just didn't notice
   b. If OPEN + branch behind main: rebase branch onto main, push --force-with-lease, enable --auto merge, signal pass
   c. If OPEN + CI failing: recirculate the droplet with notes describing the CI failure
   d. If CLOSED (not merged): recirculate with notes
4. Log the recovery action with reason

## Implementation notes
- Add a 'stuck delivery' check to the Castellarius drought protocol or as a periodic health check (every 5 minutes)
- The timeout threshold should be configurable — default: 1.5x the delivery timeout_minutes
- Rebase logic already exists in the delivery cataractae instructions; the Castellarius needs a Go implementation of: git fetch origin main && git rebase origin/main && git push --force-with-lease
- Use the gh CLI for PR state checks (already available in the sandbox environment)
- This should be idempotent — safe to run multiple times on the same stuck droplet

## Current Step: simplify

- **Type:** agent
- **Role:** simplifier
- **Context:** full_codebase

## Recent Step Notes

### From: manual

Phase 2: Three issues.

Issue 1 (unchecked error on critical write): All 8+ calls to client.SetOutcome() use _ = and discard the return error. If SetOutcome fails after killStuckSession has already run, the droplet is stranded with Outcome=="" and a dead session. On the next 5-minute cycle, isTmuxAlive returns false, the item is skipped, and there is no retry path. This is the exact failure mode the feature is supposed to prevent.

Issue 2 (missing tests — core routing logic has zero coverage): recoverStuckDelivery and recoverOpenPR contain 10+ routing branches (MERGED, CLOSED, OPEN+BEHIND succeed, OPEN+BEHIND rebase fail, OPEN+BLOCKED, OPEN+UNSTABLE, OPEN+CLEAN direct-merge, OPEN+CLEAN auto-merge fallback, OPEN+CLEAN both fail, OPEN+DIRTY/UNKNOWN, PR lookup failure, no PR found). None of these branches have tests. The 9 tests in the diff cover only threshold arithmetic, filter skipping, and rebaseAndPush git mechanics — not the Castellarius routing decisions, which are the core purpose of this feature.

Issue 3 (wrong error variable in note): stuck_delivery.go recoverOpenPR CLEAN case: when both direct-merge and auto-merge fail, the recirculate note is built with err (the direct-merge error) instead of autoErr. The note says 'all merge attempts failed' but shows only the direct-merge error, misleading operators about the failure cause.

### From: manual

Fixed three Phase 2 issues:

Issue 3 (wrong error in note): recoverOpenPR CLEAN case now uses autoErr instead of err in the 'all merge attempts failed' note — the regression is caught by TestRecoverOpenPR_CleanBothFail_RecirculatesWithAutoErr.

Issue 1 (unchecked SetOutcome errors): replaced all 12 '_ = client.SetOutcome(...)' calls with s.logSetOutcome(), which logs at Error level if the write fails. Added 4 injectable function fields (findPRFn, killSessionFn, rebaseAndPushFn, ghMergeFn) to Castellarius and initialized defaults in both New() and NewFromParts().

Issue 2 (missing routing tests): added 16 new tests covering all routing branches in recoverStuckDelivery (MERGED, CLOSED, PR-lookup-fail, no-PR, unexpected-state, session-kill) and recoverOpenPR (BEHIND-rebase-ok, BEHIND-rebase-fail, BEHIND-auto-merge-fail, BLOCKED, UNSTABLE, CLEAN-direct-ok, CLEAN-direct-fail-auto-ok, CLEAN-both-fail, DIRTY, UNKNOWN).

All 9 packages pass (25 stuck-delivery tests total, 0 regressions).

### From: scheduler

Implement pass rejected: HEAD has not advanced since last review (commit: dbde664595bb904e1ee96df40b48a5afb8700473). No new commits were found. You must commit your changes before signaling pass.

### From: manual

Implemented stuck-delivery detection and recovery. Committed: 20e20e6 in feat/ci-8hhrs (marcia worktree). All 9 packages pass. 25 stuck-delivery tests covering: threshold calc (2), filter skipping (5), recoverStuckDelivery routing (6: MERGED, CLOSED, PR-lookup-fail, no-PR, unexpected-state, session-kill), recoverOpenPR routing (10: BEHIND-rebase-ok, BEHIND-rebase-fail, BEHIND-auto-merge-fail, BLOCKED, UNSTABLE, CLEAN-direct-ok, CLEAN-direct-fail-auto-ok, CLEAN-both-fail, DIRTY, UNKNOWN), defaultRebaseAndPush git mechanics (2). SetOutcome added to CisternClient interface; 4 injectable fields (findPRFn, killSessionFn, rebaseAndPushFn, ghMergeFn) in New() and NewFromParts(); 5min stuck-delivery goroutine in Run().

<available_skills>
  <skill>
    <name>cistern-droplet-state</name>
    <description>Manage droplet state in the Cistern agentic pipeline using the `ct` CLI.</description>
    <location>/home/lobsterdog/.cistern/skills/cistern-droplet-state/SKILL.md</location>
  </skill>
  <skill>
    <name>code-simplifier</name>
    <description>code-simplifier</description>
    <location>/home/lobsterdog/.cistern/skills/code-simplifier/SKILL.md</location>
  </skill>
</available_skills>

## Signaling Completion

When your work is done, signal your outcome using the `ct` CLI:

**Pass (work complete, move to next step):**
    ct droplet pass ci-8hhrs

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-8hhrs
    ct droplet recirculate ci-8hhrs --to implement

**Block (genuinely blocked, cannot proceed):**
    ct droplet block ci-8hhrs

Add notes before signaling:
    ct droplet note ci-8hhrs "What you did / found"

The `ct` binary is on your PATH.
