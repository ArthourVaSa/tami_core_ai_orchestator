---
name: cibersecurity
description: Security review of a task (last gate before ship) or a full repository audit with `audit` (runs the cibersecurity agent, Opus).
argument-hint: TASK-NNN-MM | audit
disable-model-invocation: true
---

Run the `cibersecurity` subagent (Agent tool, `subagent_type: cibersecurity`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. On `SECURITY FAIL`, show the blocking findings and offer `/architect TASK-NNN-MM` for a fix round. On `AUDIT READY`, show the top 5 risks and offer to turn the approved findings into tasks with `/task-splitter`.
