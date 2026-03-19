# Context

## Item: ci-vwi46

**Title:** Add rate limiting to the delivery cataracta API endpoint
**Status:** in_progress
**Priority:** 1

### Description

Prevent abuse of the droplet ingestion endpoint. Apply per-IP and per-token limits with configurable thresholds in cistern.yaml.

## Current Step: implement

- **Type:** agent
- **Role:** implementer
- **Context:** full_codebase

## Recent Step Notes

### From: manual

Phase 2: internal/delivery/ratelimit_test.go and handler_test.go — goroutine leak in tests. NewRateLimiter() always starts go rl.cleanupLoop(); the goroutine runs until Close() is called. 10 test functions in ratelimit_test.go (AllowsWithinIPLimit, DeniesAtIPLimit, AllowsWithinTokenLimit, DeniesAtTokenLimit, DifferentIPsAreIndependent, DifferentTokensAreIndependent, ResetsAfterWindow, TokenLimitBlocksEvenIfIPIsUnder, IPLimitBlocksEvenIfTokenIsUnder, PartialIncrementIsAtomic) and all 14 test functions in handler_test.go call newTestLimiter/newTestHandler which calls NewRateLimiter but never call Close(). Each leaks a goroutine for the duration of the test binary. Three tests correctly use defer rl.Close() (DefaultConfig, RejectPathEvictsEmptyCounter, EvictExpiredCleansUpAllowPathEntries) — apply the same pattern to all remaining test functions, or refactor newTestLimiter/newTestHandler to register cleanup via t.Cleanup(rl.Close).

### From: manual

Fixed goroutine leaks in ratelimit_test.go and handler_test.go: added testing.TB parameter to newTestLimiter and newTestHandler, registered t.Cleanup(rl.Close) inside each helper. Removed the two now-redundant explicit defer rl.Close() calls from tests that used newTestLimiter. All 8 packages build and tests pass (go test ./internal/delivery/... ok).

### From: manual

Fixed goroutine leaks in test helpers: added testing.TB parameter to newTestLimiter (ratelimit_test.go) and newTestHandler (handler_test.go), each now calls t.Cleanup(rl.Close) so every limiter is closed when the test finishes. Removed the two pre-existing explicit defer rl.Close() calls from TestRateLimiter_RejectPathEvictsEmptyCounter and TestRateLimiter_EvictExpiredCleansUpAllowPathEntries (which used newTestLimiter) to avoid double-close panics. All call sites updated to pass t. go test ./internal/delivery/... passes.

### From: scheduler

Implement pass rejected: HEAD has not advanced since last review (commit: 567b00b2735dc808d848557f554d031d5d444238). No new commits were found. You must commit your changes before signaling pass.

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
    ct droplet pass ci-vwi46

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-vwi46
    ct droplet recirculate ci-vwi46 --to implement

**Block (genuinely blocked, cannot proceed):**
    ct droplet block ci-vwi46

Add notes before signaling:
    ct droplet note ci-vwi46 "What you did / found"

The `ct` binary is on your PATH.
