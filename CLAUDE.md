# TamiAI-TUI — multi-agent orchestrator (Go)

TAMI delegates coding tasks across several AI CLIs (claude, gemini, codex, opencode, kimi) to spread
token usage across accounts. A local Ollama model (free) classifies each task as
`simple | code | complex`, and the orchestrator picks the cheapest available agent, with automatic
fallback when an account hits its limit. The product is becoming a full TUI (Bubble Tea).

`python_prototype/` is the original prototype: reference only, never modified.

## How work is done

All work is **spec-driven**. Read `docs/WORKFLOW.md` (pipeline, IDs, statuses, handoffs) and
`docs/CODING_RULES.md` (the only coding standard) before changing code.

```
/product-owner <idea>   → docs/specs/SPEC-NNN-*.md   (you approve it)
/task-splitter SPEC-NNN → docs/tasks/SPEC-NNN/TASK-NNN-MM-*/
/flow TASK-NNN-MM       → architect → designer? → developer → coding-rules-reviewer → qa-reviewer → cibersecurity → ship
```

Each step also has its own command: `/architect`, `/designer`, `/developer`, `/coding-rules-reviewer`,
`/qa-reviewer`, `/cibersecurity` (`audit` for a full review), `/ship`, `/verify`.

In the main session: don't implement tasks directly — route them through the agents. Small fixes
outside the workflow are fine only when the user asks for them explicitly.

## Commands

```bash
go run ./cmd/tami agents                       # agent status (free)
go run ./cmd/tami run -dir /tmp/tami-test "x"  # real task: spends tokens, only the user runs it
go run ./cmd/tami serve                        # HTTP gateway on :3500
go test -race ./...                            # tests
```

**Git:** all work goes through a pull request; never commit on or push to `main` (WORKFLOW rule 6).
After cloning, enable the repo's git hooks once: `git config core.hooksPath .githooks`.

Never build binaries inside the repo. User config: `~/.tami/agents.json`; daily usage: `~/.tami/usage.json`.

## Current state

The code predates `docs/CODING_RULES.md` (old layout: `internal/domain/entities`, `ports`, `use_cases`,
`infrastructure`; `I`-prefixed interfaces; Spanish comments). **SPEC-001** migrates it to the rules'
layout (`core`, `domain`, `usecase`, `adapter`, `bootstrap`, `tui`) and to English. Until then, new code
follows the rules and old code is only touched by SPEC-001 tasks.

Language: English for code, comments, specs, tasks and reports.
