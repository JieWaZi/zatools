# Evidence Grounding Discipline

> Shared reference for `/devwiki-ask`, `/devwiki-init`, `/devwiki-ingest`, `/devwiki-scope`, `/devwiki-refresh`, `/devwiki-check`, and `/devwiki-feature-doc`.
> DevWiki outputs must always be grounded in real sources, verified code inspection, or clearly labeled inference.

---

## Core Rule

Every meaningful DevWiki statement must be traceable to at least one of:

1. `raw/` source material
2. `wiki/` structured pages derived from validated sources
3. verified code evidence from the configured code directory

`qmd` is a retrieval accelerator only. It helps surface candidates, but it is **not** the source of truth.

---

## Source Priority

### Raw Source Material

`raw/` is the strongest document-level grounding layer.

Use it for:

- original requirement intent
- original design decisions
- official feature descriptions
- code summaries and postmortems
- API and test documents

Each mirrored document page should retain:

- `source_path`
- `source_hash`

### Structured Wiki Pages

`wiki/` is the maintained knowledge layer, not the origin layer.

Use it for:

- normalized cross-document summaries
- capability aggregation
- change history
- curated links between documents and code

If the wiki conflicts with `raw/`, prefer the real source and route the discrepancy to `/devwiki-refresh` or `/devwiki-check`.

### Code Evidence

Code evidence is required when the question involves implementation reality.

Use:

- `code_refs`
- `api_refs`
- direct file and symbol verification

Do not claim a file, function, route, or endpoint is relevant unless it was actually inspected or strongly verified.

---

## Facts vs Inference

Keep facts separate from inference.

### Facts

- a `source_path` exists or is missing
- a `source_hash` matches or mismatches
- a file exists
- a symbol exists or cannot be found
- a wiki page contains a stated relationship

### Inference

- this request is probably `new`, `modify`, or `unclear`
- this capability likely owns the feature
- this code path is probably the primary implementation
- this change likely supersedes an older one

Inference is allowed only when clearly labeled as inference and backed by visible evidence.

---

## Retrieval Order

Recommended retrieval order:

1. search `wiki`
2. search `raw`
3. search `code`
4. inspect the top-K code candidates locally

This order keeps the agent from overfitting on code before understanding the documented intent.

---

## How To Use `source_hash`

`source_hash` is not decorative metadata.

It is used to answer:

- whether a raw file changed
- whether a wiki mirror is stale
- whether a refresh proposal is deterministic

If a raw file changes, the old `source_hash` must not be carried forward silently.

---

## How To Use `code_refs`

`code_refs` should be treated as structured evidence, not as keyword leftovers.

Each entry should answer:

- which file is relevant
- whether the whole file matters or only one symbol
- why it matters
- how confident the mapping is

If the file is large, prefer `path + symbol` over vague file-only references.

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
- a known capability name

Do not hide uncertainty behind broad narrative text.

---

## What NOT To Do

- Do not treat `qmd` hits as facts
- Do not let a stale wiki page outrank a changed raw source
- Do not invent `code_refs`
- Do not collapse facts and inference into the same sentence
- Do not continue blind retrieval forever; ask the user when evidence stays weak
