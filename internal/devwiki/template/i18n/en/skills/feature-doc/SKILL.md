---
name: "devwiki-feature-doc"
description: "Use when Codex needs to reverse-engineer or refresh a structured feature page from existing wiki content and source code, especially when documentation is missing, stale, or the user provides an API URL, key file, key function, or route that must be traced."
argument-hint: "<feature-name>"
---

# /devwiki-feature-doc

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


Generate or update `wiki/features/<feature-slug>.md` for a specific feature.

## Input requirements

- A clear feature name is required.
- Strongly prefer at least one entry anchor: `API URL`, `key file`, `key function`, or `page path / route`.
- If only the feature name is available, search `wiki/`, `raw/`, and the local codebase first; if the location is still unstable after a few rounds, ask the user.

## Execution rules

1. Read `references/source-priority.md` before investigating.
2. If the target page already exists, update it in place unless the user explicitly asks for a second page.
3. Read `references/trace-playbook.md` and `references/question-rules.md` before tracing code.
4. Follow the tiered recall rules in `references/zatools-qmd.md`, local-first by default:
   - start with local `grep` / file search to resolve known anchors
   - escalate to `zatools qmd search` when local hits are insufficient
   - only escalate to `zatools qmd query` when concept-level recall is needed and the environment supports acceleration; apply the shared fallback rules when it does not
5. Trace enough of the real flow to explain business flow, constraints, entry points, and implementation anchors. Do not stop at a controller, handler, or top-level function if the feature behavior still is not clear.
6. If the code cannot be located, the call chain breaks, or dynamic dispatch prevents confirmation, try a few rounds yourself and then ask the user.
7. Read `references/doc-template.md` and `references/section-examples.md` before drafting and follow their structure.
8. Before writing under `wiki/features/`, show the evidence summary, target path, and unresolved questions, then wait for confirmation.
9. Keep the page focused: summarize supported capabilities, business flow, constraints, entry points, code clues, and test entry points. Do not expand into a full implementation essay.

## Prohibited shortcuts

- Do not infer behavior from names alone.
- Do not skip critical branches, helper calls, or data transitions.
- Do not present unverified guesses as facts.
