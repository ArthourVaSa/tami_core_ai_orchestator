---
name: developer
description: Implement a planned task test-first, exactly as the architect and designer specified (runs the developer agent, Sonnet).
argument-hint: TASK-NNN-MM
disable-model-invocation: true
---

Run the `developer` subagent (Agent tool, `subagent_type: developer`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. If it answers `BLOCKED`, show the open questions from `questions.md` and offer to run `/architect` or `/designer` on the task to answer them.
