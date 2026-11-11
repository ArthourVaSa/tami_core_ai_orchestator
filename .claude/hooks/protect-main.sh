#!/usr/bin/env bash
# PreToolUse hook: stops Claude Code from committing on or pushing to main/master.
# All work goes through a pull request (docs/WORKFLOW.md, handoff rule 5).
set -euo pipefail

input=$(cat)
cmd=$(jq -r '.tool_input.command // ""' <<<"$input")
cwd=$(jq -r '.cwd // "."' <<<"$input")
branch=$(git -C "$cwd" branch --show-current 2>/dev/null || true)
on_protected=false
[[ "$branch" == "main" || "$branch" == "master" ]] && on_protected=true

block() {
  echo "Blocked by .claude/hooks/protect-main.sh: $1. Work on a branch and open a PR (/ship SPEC-NNN --pr)." >&2
  exit 2
}

# Only the `git push ...` segment (up to the next ; & | or newline) is checked, so other commands in
# the same line that mention main (e.g. `gh pr create --base main`) are not false positives.
push_segments=$(grep -oE '(^|[^[:alnum:]_-])git([[:space:]]+-C[[:space:]]+[^[:space:]]+)?[[:space:]]+push([[:space:]][^;&|]*)?' <<<"$cmd" || true)
if [[ -n "$push_segments" ]]; then
  grep -Eq '([[:space:]:+])(main|master)([[:space:]]|$)' <<<"$push_segments" && block "pushing to main is not allowed"
  $on_protected && block "you are on $branch; pushing it is not allowed"
fi

if $on_protected && grep -Eq '(^|[^[:alnum:]_-])git([[:space:]]+-C[[:space:]]+[^[:space:]]+)?[[:space:]]+commit([[:space:]]|$)' <<<"$cmd"; then
  block "committing on $branch is not allowed"
fi

exit 0
