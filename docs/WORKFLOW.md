# TamiAI-TUI — Spec-driven workflow

Every piece of work starts as a spec and ends as a reviewed, shipped task. This file is the contract
all agents follow. If an agent's instructions and this file disagree, this file wins.

## Pipeline

```
product-owner ──► task-splitter ──► per task:
  (spec, with you)   (child tasks)    architect ──► designer? ──► developer ──► coding-rules-reviewer
                                         ▲              ▲             │                 │
                                         └──────────────┴─ questions ─┘                 ▼
                                         ▲                                         qa-reviewer
                                         │                                              │
                                         └──── findings (auto, max 3 rounds) ◄── cibersecurity
                                                                                        │ PASS
                                                                                        ▼
                                                                                      ship
```

| Step | Who | Model | Command | Writes |
|---|---|---|---|---|
| 1 | product-owner | Opus | `/product-owner <idea or SPEC-ID>` | `docs/specs/SPEC-NNN-slug.md` |
| 2 | task-splitter | Opus | `/task-splitter SPEC-NNN` | `docs/tasks/SPEC-NNN/` |
| 3 | architect | Opus | `/architect TASK-NNN-MM` | `plan.md` |
| 4 | designer (only if `design: required`) | Sonnet | `/designer TASK-NNN-MM` | `design.md` |
| 5 | developer | Sonnet | `/developer TASK-NNN-MM` | code + `dev-log.md` (+ `questions.md`) |
| 6 | coding-rules-reviewer | Sonnet | `/coding-rules-reviewer TASK-NNN-MM` | `rules-report.md` |
| 7 | qa-reviewer | Sonnet | `/qa-reviewer TASK-NNN-MM` | `qa-report.md` (+ extra `_test.go`) |
| 8 | cibersecurity | Opus | `/cibersecurity TASK-NNN-MM` or `/cibersecurity audit` | `security-report.md` / `SECURITY_REVIEW.md` |
| 9 | ship | Haiku | `/ship TASK-NNN-MM` or `/ship SPEC-NNN --pr` | commit, PR |
| — | all of 3-9 in order | — | `/flow TASK-NNN-MM` or `/flow SPEC-NNN` | — |
| — | cli-researcher | Haiku | on demand (see below) | `research.md` |

## IDs and folders

- Spec: `SPEC-NNN` (3 digits, sequential, never reused). File: `docs/specs/SPEC-001-align-codebase.md`.
- Task: `TASK-NNN-MM` where `NNN` is the parent spec and `MM` the child number. Folder:
  `docs/tasks/SPEC-001/TASK-001-02-move-domain-entities/`.

```
docs/
  CODING_RULES.md
  WORKFLOW.md
  specs/SPEC-001-align-codebase.md
  tasks/SPEC-001/
    README.md                       # index: task table, dependencies, status (task-splitter owns it)
    TASK-001-01-<slug>/
      task.md                       # task-splitter
      research.md                   # cli-researcher (optional)
      plan.md                       # architect
      design.md                     # designer (optional)
      questions.md                  # developer → architect/designer, answered in place
      dev-log.md                    # developer: TDD evidence
      rules-report.md               # coding-rules-reviewer
      qa-report.md                  # qa-reviewer
      security-report.md            # cibersecurity
  security/BACKLOG.md               # Low/Info security findings that do not block a task
SECURITY_REVIEW.md                  # full audit (/cibersecurity audit)
```

## Status

Every spec and task file starts with YAML frontmatter. Only the agent named in the table may set each status.

Spec `status`: `draft` → `approved` (only after you approve it explicitly) → `split` → `done`.

Task `status`:

| Status | Set by | Meaning |
|---|---|---|
| `todo` | task-splitter | created, not planned |
| `planned` | architect | `plan.md` ready |
| `designed` | designer | `design.md` ready |
| `in-progress` | developer | implementing |
| `blocked` | developer | `questions.md` has open questions |
| `rules-review` / `qa` / `security` | developer / reviewers | waiting for that gate |
| `changes-requested` | any reviewer | a gate failed; goes back to the architect |
| `done` | cibersecurity (PASS) | all gates passed, ready to ship |
| `shipped` | ship | committed (and in a PR) |

## Handoff rules

1. **Developer never decides.** Anything ambiguous, contradictory or improvable goes to `questions.md`
   with `to: architect` or `to: designer`, then the developer stops with `BLOCKED`. The architect or
   designer answers in `questions.md` and updates `plan.md`/`design.md`; then the developer resumes.
2. **TDD is mandatory.** The plan lists the tests first. The developer records in `dev-log.md`, for each
   step, the failing (red) test run and then the passing (green) run. The qa-reviewer checks this evidence.
3. **Gates loop back through the architect.** A FAIL from the coding-rules-reviewer, qa-reviewer or
   cibersecurity sets `changes-requested`. The architect adds a "Fix round N" section to `plan.md`, then
   developer → reviewers run again. `/flow` does this automatically, **max 3 rounds per task**; after
   that it stops and asks you.
4. **Security severities:** Critical/High/Medium block the task and start a fix round automatically.
   Low/Info are recorded in `docs/security/BACKLOG.md` and do not block.
5. **Only `ship` touches git history.** No agent commits, pushes or opens PRs. `ship` pushes and opens
   a PR only with `--pr`, after you confirm.
6. **No real tokens in automatic steps.** Agents never run `go run ./cmd/tami run ...` or `serve`
   against real agent CLIs. Tests use fakes.

## Where cli-researcher fits

The cli-researcher (Haiku, cheap) looks up the real flags of an external AI CLI from its `--help` and
its official docs, without ever running a prompt. It is used:

- **By the architect**, when a task integrates or changes an agent CLI (claude, gemini, codex, opencode,
  kimi). The architect writes `research-needed: <cli>` in `plan.md`; `/flow` runs cli-researcher, saves
  `research.md`, and re-runs the architect with it.
- **By cibersecurity**, to check what each CLI's auto-approve flag allows (`--yolo`, `--full-auto`,
  `acceptEdits`), how it stores credentials and what env it inherits.
- **On demand**, when a CLI is updated and TAMI stops working with it.
