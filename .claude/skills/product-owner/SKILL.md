---
name: product-owner
description: Define the product with the user and write or refine a spec (runs the product-owner agent, Opus, as a conversation).
argument-hint: <feature idea> | SPEC-NNN
disable-model-invocation: true
---

This is a conversation between the user and the `product-owner` subagent. You only relay; do not
write or edit the spec yourself.

1. Start the `product-owner` subagent (Agent tool, `subagent_type: product-owner`, give it a `name`
   such as `po-<short-slug>` so it can be messaged again) with:
   > Input: $ARGUMENTS
   > Follow docs/WORKFLOW.md. If the input is a SPEC id, refine that spec.

   If no argument was given, first ask the user what they want to build.

2. If it answers `QUESTIONS`, ask the user those questions with AskUserQuestion (keep its options and
   its recommendation first). Send the answers back to the **same** agent with SendMessage (load it
   with ToolSearch `select:SendMessage` if needed), so it keeps its context. Repeat until it answers
   `SPEC READY`.

3. On `SPEC READY`, show the user the summary and the path, and ask: approve, request changes, or
   stop.
   - **Approve:** set `status: approved` in the spec's frontmatter yourself (that is the only edit you
     make) and add a line to its "Decisions log". Offer `/task-splitter SPEC-NNN` or `/flow SPEC-NNN`.
   - **Changes:** send the user's feedback to the same agent and go back to step 2.
