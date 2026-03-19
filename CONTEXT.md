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

Phase 2 — four issues found:

1. internal/delivery/handler.go:293-307 — Security: per-IP rate limit bypass via spoofed proxy headers. realIP() unconditionally trusts X-Real-IP and X-Forwarded-For. Any client can set X-Real-IP: <arbitrary IP> on every request to rotate identities and evade per-IP limiting entirely. Fix requires either (a) only honouring proxy headers when RemoteAddr is a known trusted proxy, or (b) dropping proxy-header support and documenting that a trusted proxy must strip those headers before forwarding.

2. Missing integration — handler never mounted, RateLimitConfig never wired. The diff adds internal/delivery and AqueductConfig.RateLimit but includes no code (a) registering NewHandler on any HTTP mux/server or (b) translating RateLimitConfig (Window is string) into delivery.Config (Window is time.Duration). The new package is dead code; the endpoint is unprotected.

3. internal/delivery/ratelimit.go:636-688 — Resource leak: ipCounters and tokCounters grow without bound. Entries are created per unique IP/token and never evicted, even after their window expires. An adversary rotating source IPs or bearer tokens causes unbounded heap growth leading to OOM. Stale entries must be pruned (e.g. delete map key when windowCounter.times becomes empty after pruneAndCount).

4. internal/delivery/handler.go:250 — Logic error: Retry-After hardcoded to "60". The value is a string literal unrelated to rl.cfg.Window. For a configured window of 30s the header tells clients to wait 60s; for 5m it still says 60s. The value must be derived from the actual window duration.

### From: manual

Fixed all 4 Phase 2 issues: (1) Security — realIP() now only trusts X-Real-IP/X-Forwarded-For when RemoteAddr is in a trusted proxy network (loopback 127.0.0.0/8 and ::1/128 by default); added trustedNets to Handler. (2) Integration — added DeliveryAddr to AqueductConfig; wired cisternAdder adapter, RateLimitConfig→delivery.Config translation (time.ParseDuration), NewRateLimiter+NewHandler, http.Server mounted at /droplets in start.go. (3) Resource leak — Allow() deletes map entries whose times slice is empty after pruneAndCount on the reject path. (4) Retry-After — added Window() accessor to RateLimiter; header now derived from actual window duration. All 8 packages pass.

### From: manual

All 4 Phase 2 issues resolved. (1) Security: realIP() gated on trusted proxy networks (loopback default), preventing spoofed-header IP bypass. (2) Integration: delivery HTTP server mounted in aqueduct start command via cisternAdder adapter; RateLimitConfig.Window parsed to time.Duration; DeliveryAddr added to AqueductConfig. (3) Resource leak: stale ipCounters/tokCounters entries evicted on rejection when times slice is empty. (4) Retry-After: derived from Window() accessor on RateLimiter, not hardcoded. All 8 packages build and pass.

### From: scheduler

Implement pass rejected: HEAD has not advanced since last review (commit: 478b485ee8caec82caf06b69db539accd192e2d4). No new commits were found. You must commit your changes before signaling pass.

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
