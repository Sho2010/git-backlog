#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SANDBOX="$ROOT/tmp/sandbox-repo"
BASE_URL="${BACKLOG_BASE_URL:-https://example.backlog.jp}"

rm -rf "$SANDBOX"
mkdir -p "$SANDBOX"
cd "$SANDBOX"

git init -q -b main
git config user.email "sandbox@example.com"
git config user.name "git-backlog sandbox"
git config backlog.baseUrl "$BASE_URL"
git config backlog.cacheTTL 1m

git commit -q --allow-empty -m "init"

git branch TEST-1-first-match
git branch BUG-42-bar-baz
git branch feature/PROJ-100-nested
git branch MULTI-1-and-OTHER-2-keys
git branch no-key-branch
git branch lower-99-not-match

cat <<EOF
Sandbox ready: $SANDBOX
  baseUrl       = $BASE_URL  (override with BACKLOG_BASE_URL)
  cacheTTL      = 1m
  issuePattern  = (default) [A-Z][A-Z0-9_]*-[0-9]+

Branches:
  TEST-1-first-match           -> TEST-1
  BUG-42-bar-baz               -> BUG-42
  feature/PROJ-100-nested      -> PROJ-100
  MULTI-1-and-OTHER-2-keys     -> MULTI-1   (first match wins)
  no-key-branch                -> (silent exit 0)
  lower-99-not-match           -> (silent exit 0; lowercase rejected)

Try:
  cd $SANDBOX
  git switch TEST-1-first-match
  export BACKLOG_API_KEY=...   # PAT
  $ROOT/git-backlog            # = git-backlog current
EOF
