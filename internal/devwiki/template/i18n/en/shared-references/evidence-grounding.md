# Evidence Grounding Discipline

> Shared reference for `devwiki-query`, `devwiki-ingest`, `devwiki-maintain`, `devwiki-code-to-doc`, and `devwiki-project-router`.
> DevWiki outputs must always be grounded in real sources, verified code inspection, or clearly labeled inference.

## Core Rule

Every meaningful DevWiki statement must be traceable to at least one of:

1. `raw/` source material
2. `wiki/capabilities/`, `wiki/features/`, `wiki/workflows/`, or `wiki/troubleshooting/`
3. verified code evidence from the configured code directory

`qmd` is a retrieval accelerator only. It is not the source of truth.

## Source Priority

`raw/` is the strongest source layer. Use it for original requirements, design decisions, feature notes, test plans, and test records. Pages should retain this evidence through inline `sources.path` and `sources.hash`.

`wiki/` is the maintained knowledge layer. Use capability pages for business capability summaries, feature pages for design and behavior, workflow pages for engineering location and code references, and troubleshooting pages for diagnosis and fixes.

If wiki content conflicts with `raw/`, prefer the real source and show the conflict in the answer or proposal; save it to `wiki/outputs/` only when the user asks for a report.

Code evidence is required when the question involves implementation reality. Use workflow or troubleshooting `code_refs`, `api_entries`, `test_refs`, and direct file/symbol verification. Do not claim a file, function, route, or endpoint is relevant unless it was inspected or strongly verified.

## Facts vs Inference

Keep facts separate from inference.

Facts: a raw path exists, a hash matches, a file exists, a symbol exists, or a wiki page states a relationship.

Inference: a request is probably new/modified, a capability likely owns a feature, a feature likely covers a request, or a code path is probably primary. Label inference clearly and include visible evidence.

## Retrieval Order

Recommended order:

1. Search the wiki layer matching the user's intent.
2. Search `raw/`.
3. Decide whether pages already support the answer.
4. Search code only when page evidence is insufficient or implementation reality is explicitly requested.
5. Inspect top-K code candidates locally only when code evidence is required.

If documents already support the answer, do not perform another code-expansion pass just for reassurance.

## How To Use `sources.hash`

`sources.hash` answers whether a raw file changed, whether a page is stale, and whether a refresh proposal is deterministic. If a raw file changes, the old hash must not be carried forward silently.

## How To Use `code_refs`

`code_refs` are structured evidence. Each entry should say which file is relevant, whether the whole file or only one symbol matters, why it matters, and confidence.

`code_refs` belong on workflow or troubleshooting pages, not capability or feature pages.

## Low-Confidence Protocol

When retrieval remains weak after bounded rounds:

- stop expanding
- summarize found evidence
- ask the user 1 to 3 concrete questions

Do not hide uncertainty behind broad narrative text.

## What NOT To Do

- Do not treat `qmd` hits as facts.
- Do not let a stale wiki page outrank changed raw source.
- Do not invent `code_refs`.
- Do not mix facts and inference in one sentence.
- Do not keep searching indefinitely when evidence remains thin.
