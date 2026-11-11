---
name: flow
description: Drive the whole TamiAI-TUI pipeline for a task or a spec — architect, designer (if needed), developer, coding-rules-reviewer, qa-reviewer, cibersecurity, ship — with automatic question routing and fix rounds.
argument-hint: TASK-NNN-MM | SPEC-NNN
disable-model-invocation: true
---

You are the conductor. You run the agents in order, route their handoffs through the task folder, and
keep your own context small: never paste whole reports, only each agent's final line and its key
findings. Contract: `docs/WORKFLOW.md`.

Input: $ARGUMENTS (ask for it if empty).

## SPEC-NNN

1. Spec `status: draft` → stop: "approve it first with `/product-owner SPEC-NNN`".
2. Spec `approved` and no `docs/tasks/SPEC-NNN/` → run the `task-splitter` agent. Show the task table and
   ask the user to confirm before continuing.
3. Run the **task loop** below for every task not `done`/`shipped`, in dependency order (a task starts
   only when its `depends_on` tasks are `shipped`).
4. When all tasks are `shipped`, ask the user (AskUserQuestion) whether to open the PR. On yes, run the
   `ship` skill with `SPEC-NNN --pr` and give the PR URL.

## TASK-NNN-MM — the task loop

Read the task's `status` and `fix_round` from `task.md` and continue from where it is (the loop can resume).
Every agent is called with the Agent tool, `subagent_type` = its name, prompt:
`Input: TASK-NNN-MM. Mode: <mode>. Follow docs/WORKFLOW.md.`

| Status | Do |
|---|---|
| `todo` | `architect` (plan) |
| `planned` + `design: required` and no `design.md` | `designer` (design) |
| `planned` / `designed` | `developer` |
| `blocked` | the agent named in each open question of `questions.md` (`architect` or `designer`, answer mode), then `developer` again |
| `rules-review` | `coding-rules-reviewer` skill with the task id |
| `qa` | `qa-reviewer` |
| `security` | `cibersecurity` (task mode) |
| `changes-requested` | `architect` (fix round), then `developer` and the gates again |
| `done` | `ship` skill with the task id (local commit only) |

Special answers:
- `RESEARCH NEEDED: <cli>` from the architect → `cli-researcher` with the CLI, the question and the
  path `<task folder>/research.md`, then `architect` again.
- `SPEC QUESTIONS` / something only the user can decide → ask the user with AskUserQuestion, write the
  answer into the right file (spec "Decisions log", or under the question in `questions.md`), continue.

After each agent, re-read `task.md` status; the agents set it. If an agent's final line and the status
disagree, stop and show both to the user.

## Limits (stop and ask the user when reached)

- **Fix rounds:** `fix_round` reaches 3 and a gate still fails.
- **Question rounds:** the developer is `blocked` 3 times in the same fix round.
- **Same finding twice:** a reviewer reports a finding a previous round already claimed to fix.
- Any `SHIP BLOCKED`.

When you stop, give the user: the task, its status, the round, the blocking findings (one line each),
and the commands to continue (`/flow TASK-NNN-MM` resumes).

## Progress

After each step print one line, e.g. `TASK-001-02 · developer → IMPLEMENTED (round 0)`, and at the end
a table of the tasks with their final status.
