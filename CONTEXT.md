# Context

## Item: ci-7ecn4

**Title:** Add ct import CLI subcommand
**Status:** in_progress
**Priority:** 1

### Description

Add 'ct import' subcommand in cmd/ct/import.go. Usage: ct import <provider> <issue-key> [flags]. Flags: --repo (target repo, required), --filter (run through filtration before filing), --priority (override mapped priority), --complexity (override, default 1). Flow: (1) resolve provider from first arg (e.g. 'jira'), (2) load provider config from cistern.yaml trackers section, (3) call provider.FetchIssue(key), (4) map ExternalIssue fields to droplet fields, (5) set external_ref to 'provider:key', (6) if --filter flag: run ct filter with pre-populated title/description, (7) else: add droplet directly via client.AddDroplet(). Print created droplet ID on success. Provider registry maps string names to TrackerProvider constructors. Depends on ci-xrgv2, ci-ikbj2, ci-g6so3.

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
    ct droplet pass ci-7ecn4

**Recirculate (needs rework — send back upstream):**
    ct droplet recirculate ci-7ecn4
    ct droplet recirculate ci-7ecn4 --to implement

**Pool (cannot currently proceed):**
    ct droplet pool ci-7ecn4

Add notes before signaling:
    ct droplet note ci-7ecn4 "What you did / found"

The `ct` binary is on your PATH.
