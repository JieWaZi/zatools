# Evidence Grounding Discipline

> Shared reference for `/devwiki-ask`, `/devwiki-init`, `/devwiki-ingest`, `/devwiki-refresh`, `/devwiki-check`, and `/devwiki-feature-doc`.
> DevWiki outputs must always be grounded in real sources, verified code inspection, or clearly labeled inference.

---

## Core Rule

Every meaningful DevWiki statement must be traceable to at least one of:

1. `raw/` source material
2. `wiki/capabilities/` or `wiki/features/`
3. verified code evidence from the configured code directory

`qmd` is a retrieval accelerator only. It helps surface candidates, but it is **not** the source of truth.

---

## Source Priority

### Raw Source Material

`raw/` is the strongest source layer.

Use it for:

- original requirement intent
- original design decisions
- original feature notes
- test plans and test records

Feature pages should retain this evidence through `sources.path` and `sources.hash`.

### Structured Wiki Pages

`wiki/` is the maintained knowledge layer, not the origin layer.

Use it for:

- business capability summaries
- feature-level workflow summaries
- curated links between capabilities and features
- curated code, API, and test entry points on feature pages

If wiki content conflicts with `raw/`, prefer the real source and route the discrepancy to `/devwiki-refresh` or `/devwiki-check`.

### Code Evidence

Code evidence is required when the question involves implementation reality.

Use:

- `code_refs`
- `api_entries`
- `test_refs`
- direct file and symbol verification

Do not claim a file, function, route, or endpoint is relevant unless it was actually inspected or strongly verified.

---

## Facts vs Inference

Keep facts separate from inference.

### Facts

- a raw `path` exists or is missing
- a `hash` matches or mismatches
- a file exists
- a symbol exists or cannot be found
- a wiki page contains a stated relationship

### Inference

- this request is probably `new`, `modify`, or `unclear`
- this capability likely owns the feature
- this feature likely covers the requested change
- this code path is probably the primary implementation

Inference is allowed only when clearly labeled as inference and backed by visible evidence.

---

## Retrieval Order

Recommended retrieval order:

1. search `wiki`
2. search `raw`
3. decide whether the pages already support the answer
4. only search `code` when the pages are insufficient or the question explicitly asks about implementation reality
5. inspect the top-K code candidates locally only when code evidence is actually required

This order keeps the agent from overfitting on code before understanding the documented intent.

For `/devwiki-ask`, answer from documents first, then decide whether code is worth opening.

For `/devwiki-ask`, answer from wiki/raw first, then decide whether code is worth opening.

If the documents already support the answer, do not perform another code-expansion pass just for reassurance.

If the pages already support the answer, do not perform another code-expansion pass just for reassurance.

---

## How To Use `sources.hash`

`sources.hash` is not decorative metadata.

It is used to answer:

- whether a raw file changed
- whether a feature page is stale
- whether a refresh proposal is deterministic

If a raw file changes, the old hash must not be carried forward silently.

---

## How To Use `code_refs`

`code_refs` should be treated as structured evidence, not as keyword leftovers.

Each entry should answer:

- which file is relevant
- whether the whole file matters or only one symbol
- why it matters
- how confident the mapping is

If the file is large, prefer `path + symbol` over vague file-only references.

`code_refs` belong on feature pages, not capability pages.

---

## Low-Confidence Protocol

When retrieval remains weak after a few bounded rounds:

- stop expanding the search
- summarize what evidence was found
- ask the user 1 to 3 concrete questions

Prefer asking for:

- an API URL
- a key file
- a key function
- a page route
- a known capability or feature name

Do not hide uncertainty behind broad narrative text.

---

## What NOT To Do

- Do not treat `qmd` hits as facts
- Do not let a stale wiki page outrank a changed raw source
- Do not invent `code_refs`
- Do not collapse facts and inference into the same sentence
- Do not continue blind retrieval forever; ask the user when evidence stays weak
