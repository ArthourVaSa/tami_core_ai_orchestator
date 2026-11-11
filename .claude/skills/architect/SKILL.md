---
name: architect
description: Plan how a task is built, answer developer questions, or plan a fix round (runs the architect agent, Opus).
argument-hint: TASK-NNN-MM
disable-model-invocation: true
---

Run the `architect` subagent (Agent tool, `subagent_type: architect`) with this prompt:

> Input: $ARGUMENTS
> Follow docs/WORKFLOW.md. Find the files for this id under docs/specs/ or docs/tasks/ yourself.

If no argument was given, ask the user for it instead of guessing.

When it finishes, relay its final line and a short summary to the user. If it answers `RESEARCH NEEDED`, run the `cli-researcher` agent with the CLI and the question, telling it to write `research.md` in the task folder, then run the architect again on the same task.
