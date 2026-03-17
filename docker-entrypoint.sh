#!/usr/bin/env bash
set -euo pipefail

# ── 1. Install / update Claude CLI ───────────────────────────────────────────
echo "==> Updating Claude CLI..."
npm install -g @anthropic-ai/claude-code --quiet

# ── 2. Persist Claude auth in the volume ─────────────────────────────────────
# The claude CLI stores its config in ~/.claude by default.
# Symlink it into the volume so auth tokens survive container restarts.
mkdir -p /root/.cistern/auth/claude
if [ ! -e /root/.claude ]; then
    ln -sf /root/.cistern/auth/claude /root/.claude
fi

# ── 3. GitHub auth ────────────────────────────────────────────────────────────
if ! gh auth status >/dev/null 2>&1; then
    if [ -n "${GH_TOKEN:-}" ]; then
        echo "==> Authenticating gh with GH_TOKEN..."
        echo "$GH_TOKEN" | gh auth login --with-token
    else
        echo "==> GitHub authentication required (opens browser)."
        echo "    On headless servers set GH_TOKEN in your .env file instead."
        gh auth login --web
    fi
fi

# ── 4. Claude auth ────────────────────────────────────────────────────────────
# If ANTHROPIC_API_KEY is set, the claude CLI uses it directly — no login needed.
# Otherwise check for persisted auth tokens in the volume.
if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
    if [ -z "$(ls -A /root/.cistern/auth/claude 2>/dev/null)" ]; then
        echo "==> Claude authentication required (opens browser)."
        echo "    On headless servers set ANTHROPIC_API_KEY in your .env file instead."
        claude auth login
    fi
fi

# ── 5. Cistern init (idempotent) ──────────────────────────────────────────────
# ct init uses writeFileIfAbsent internally — existing files are skipped.
# Output is suppressed to avoid noise on every container start.
ct init > /dev/null 2>&1

# ── 6. Hand off ───────────────────────────────────────────────────────────────
exec "$@"
