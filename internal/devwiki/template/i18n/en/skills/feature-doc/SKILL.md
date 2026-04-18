---
name: "devwiki-feature-doc"
description: "Use when Codex needs to reverse-engineer a raw feature document for a specific engineering capability from existing wiki content and source code, especially when documentation is missing, stale, or the user provides an API URL, key file, key function, or route that must be traced through a full call chain."
argument-hint: "<feature-name>"
---

# /devwiki-feature-doc

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


Generate or update `raw/features/<FeatureName>特性文档.md` for a specific capability.

## Input requirements

- A clear feature name is required.
- Strongly prefer at least one entry anchor: `API URL`, `key file`, `key function`, or `page path / route`.
- If only the feature name is available, search `wiki/`, `raw/`, and the local codebase first; if the location is still unstable after a few rounds, ask the user.


## Execution rules

1. Read `references/source-priority.md` before investigating.
2. If the target file already exists, ask whether to update it or create a new document with a different confirmed name.
3. Read `references/trace-playbook.md` and `references/question-rules.md` before tracing code.
4. Run `zatools qmd status` first. If `zatools qmd status` is healthy, prefer `zatools qmd query` across `wiki / raw / code`, then inspect top-K hits with `zatools qmd get` / `zatools qmd multi-get`.
5. Follow the full business path. Do not stop at a controller, handler, or top-level function.
6. If the code cannot be located, the call chain breaks, or dynamic dispatch prevents confirmation, try a few rounds yourself and then ask the user.
7. Read `references/doc-template.md` and `references/section-examples.md` before drafting and follow their structure.
8. Before writing under `raw/features/`, show the evidence summary, target path, and unresolved questions, then wait for confirmation.

## Prohibited shortcuts

- Do not infer behavior from names alone.
- Do not skip critical branches, helper calls, or data transitions.
- Do not present unverified guesses as facts.
