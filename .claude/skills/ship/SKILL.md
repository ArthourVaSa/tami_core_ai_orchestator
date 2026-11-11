---
name: ship
description: Git specialist. Commits a task that passed qa-reviewer and cibersecurity onto its spec branch, and with --pr pushes and opens the pull request. Use only when the task status is done.
argument-hint: TASK-NNN-MM | SPEC-NNN --pr
context: fork
model: haiku
allowed-tools: Read, Edit, Bash, Grep, Glob
---

You prepare the git work for TamiAI-TUI. Input: $ARGUMENTS

## 1. Gate check (stop on any failure, report it, change nothing)

For each task to ship (one task, or for `SPEC-NNN --pr` every task in `docs/tasks/SPEC-NNN/`):
- `task.md` has `status: done` (or `shipped` already, when opening the PR).
- `qa-report.md` and `security-report.md` exist, have `verdict: PASS`, and their `round` equals the
  task's `fix_round`.
- `rules-report.md` has `verdict: PASS`.
Then run the `verify` checks: `gofmt -l .` prints nothing, `go vet ./...` and `go test -race ./...` pass.

## 2. Branch

Branch name: `spec/SPEC-NNN-<spec slug>`. If you are on `main`, create it (`git switch -c`) or switch to
it if it exists. Never commit to `main`.

## 3. Commit (one per task)

- Stage only what belongs to the task: its folder under `docs/tasks/`, the files listed in `plan.md`
  ("Files" table), test files listed in `qa-report.md` ("Tests added"), and the spec/README files
  whose status changed. Use `git add <paths>`, never `git add -A` or `git add .`.
- If other files are modified, leave them unstaged and list them in your output.
- Never stage secrets: refuse to commit `.env*`, `*.pem`, `*.key`, `~/.tami` contents, or anything that
  looks like an API key; report it instead.
- Message (Conventional Commits, English):
  ```
  <type>(<scope>): <task title> (TASK-NNN-MM)

  <2-4 lines: what changed and why, from task.md and plan.md>

  Spec: SPEC-NNN
  Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
  ```
  `type`: feat, fix, refactor, test, docs, chore. `scope`: main package touched.
- After committing, set the task `status: shipped` in `task.md` and in `docs/tasks/SPEC-NNN/README.md`,
  and amend them into the same commit.

Without `--pr`, stop here: **never push**.

## 4. Pull request (only with --pr)

- Every task of the spec must be `shipped`; otherwise list the missing ones and stop.
- `gh auth status` must show an account that can push to `origin`; if not, stop and say which account is active.
- `git push -u origin <branch>`, then `gh pr create --base main --title "SPEC-NNN: <spec title>" --body-file <tmp file>` with:
  ```markdown
  ## Summary
  <goal from the spec, 2-3 lines>

  ## Tasks
  | ID | Title | Commit |

  ## Verification
  - Tests: <go test -race result line>
  - Coverage domain/usecase: <from qa-reports>
  - Mutation: <from qa-reports>
  - Security: <PASS per task; backlog items added, if any>

  Spec: docs/specs/SPEC-NNN-<slug>.md

  🤖 Generated with [Claude Code](https://claude.com/claude-code)
  ```
- Set the spec `status: done` only after the PR is created.

## Output

`SHIPPED: <ids> on <branch>` with the commit hashes (and the PR URL with --pr), plus any files you left
unstaged. On failure: `SHIP BLOCKED: <reason>`.
