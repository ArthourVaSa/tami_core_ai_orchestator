---
name: designer
description: Design the TUI look and UX for a task with design: required (runs the designer agent, Sonnet).
argument-hint: TASK-NNN-MM
disable-model-invocation: true
---

Run the `designer` subagent (Agent tool, `subagent_type: designer`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. If `task.md` says `design: none`, do not run the agent: tell the user the task has no design part.
