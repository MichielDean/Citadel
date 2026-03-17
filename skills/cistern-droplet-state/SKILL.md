# Cistern Droplet State

## Description

This skill teaches Cataractae how to manage droplet state using the `ct` CLI.
Every Cataracta MUST signal its outcome before exiting. Failing to signal leaves
the droplet stranded in the pipeline.

## Finding Your Droplet ID

The droplet ID is in `CONTEXT.md` — look for the `## Item:` line at the top.
Example: `## Item: ci-hvwi9` → droplet ID is `ci-hvwi9`.

## Commands

### Signal Completion

**Pass** — work is complete, tests pass, commit is done:
```
ct droplet pass <id> --notes "What was done and verified."
```

**Recirculate** — work needs rework by a prior stage:
```
ct droplet recirculate <id> --notes "What needs fixing and why."
ct droplet recirculate <id> --to implement --notes "Specific issues for the implementer."
```

**Block** — genuinely cannot proceed (missing dependency, ambiguous requirements):
```
ct droplet block <id> --notes "Exactly what is blocking and what is needed."
```

### Add a Note (without changing state)
```
ct droplet note <id> "Observation or progress note."
```

## When to Use Each Signal

| Signal | When |
|--------|------|
| `pass` | Work complete, requirements met, tests pass, commit made |
| `recirculate` | Found issues that require changes by a prior stage |
| `block` | Cannot proceed — missing input, broken dependency, unclear requirements |

## Rules

1. **Always add `--notes`** when signaling pass, recirculate, or block.
2. **Commit before passing** — reviewers receive your diff. No commit = empty diff.
3. **Never push to origin** — local commit only.
4. **Pass only if tests pass** — run `go test ./...` (or equivalent) first.
5. **One signal per session** — call exactly one of pass/recirculate/block before exiting.

## Examples

```bash
# After successful implementation
ct droplet pass ci-hvwi9 --notes "Implemented X. Added 5 tests. All pass."

# After finding a bug during review
ct droplet recirculate ci-hvwi9 --notes "Required: nil dereference in foo.go:42 when input is empty."

# When blocked on missing dependency
ct droplet block ci-hvwi9 --notes "Blocked: ci-0ll74 (skills infra) must merge first."

# Progress note without state change
ct droplet note ci-hvwi9 "Running tests — 3 of 7 packages done."
```
