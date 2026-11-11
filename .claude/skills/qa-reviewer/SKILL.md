---
name: qa-reviewer
description: Verify TDD, run all tests including race, coverage and mutation, and add acceptance tests for a task (runs the qa-reviewer agent, Sonnet).
argument-hint: TASK-NNN-MM
disable-model-invocation: true
---

Run the `qa-reviewer` subagent (Agent tool, `subagent_type: qa-reviewer`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. On `QA FAIL`, show the findings and offer to run `/architect TASK-NNN-MM` for a fix round.
