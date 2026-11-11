---
name: task-splitter
description: Split an approved spec into child tasks with IDs under docs/tasks/SPEC-NNN/ (runs the task-splitter agent, Opus).
argument-hint: SPEC-NNN
disable-model-invocation: true
---

Run the `task-splitter` subagent (Agent tool, `subagent_type: task-splitter`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. If it answers `NOT APPROVED`, tell the user to approve the spec first (`/product-owner SPEC-NNN`). If it answers `SPEC QUESTIONS`, show them and offer to run `/product-owner SPEC-NNN` with them.
