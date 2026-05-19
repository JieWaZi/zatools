# zatools qmd Usage

> Shared reference for every DevWiki skill that retrieves, syncs, or refreshes inside a DevWiki workspace.

This file defines two things together:
1. **Retrieval routing** — how to pick between local search, `zatools qmd search`, and `zatools qmd query`
2. **Invocation constraints** — what every real `zatools qmd ...` call must obey (model flags, sandbox prerequisites)

## Retrieval routing (tiered recall)

DevWiki skills recall evidence in cost/latency ascending order. Stop as soon as any tier returns a strong enough top-K; do not keep escalating for its own sake.

### Tier 1: Local `grep` / file search (default starting point)

When the question already contains a concrete anchor, **prefer local search** and do not call `zatools qmd ...`:

- Known symbol name, function name, class name
- Known file name, directory name, path fragment
- API URL, route, error code
- Capability slug, ticket id, commit hash

This tier is usually enough to answer directly, and it is the fastest and most deterministic path on machines without a GPU.

### Tier 2: `zatools qmd search` (keyword recall)

Use `qmd search` when the question has keywords but the exact landing point is unknown:

- Typical queries: "code related to SAML metadata", "where is the auth-failure log emitted", "documents about the payment callback"
- Default recall covers only the `wiki` collection; raw or code directories are included only when the user manually adds those collections to `config/search.yaml`
- Keyword-based, no embeddings required, CPU-friendly

### Tier 3: `zatools qmd query` (semantic recall)

Only escalate to `qmd query` when tiers 1 and 2 are insufficient AND the question is a **concept / design / intent** question:

- Typical queries: "how is the permission system designed overall", "why was this done this way originally", "what is the architecture for rate limiting"
- Requires embedding + rerank; noticeably slow without a GPU
- Upper-tier recall only — never the first choice

### Escalation and stopping

- Stop escalating as soon as any tier returns a strong enough top-K; do not fall through to semantic recall out of habit
- When escalating, use the previous tier's hits as anchors for the next tier (e.g. feed tier-1 file paths into `qmd search` to narrow scope)
- After a few rounds with no useful result, **stop expanding and ask the user for 1–3 concrete anchors**; do not keep blindly escalating

### Hard fallback

Before running `zatools qmd query`, verify that local acceleration is actually usable:

- If no GPU/accelerator is detected, or the current environment can only run embed / rerank on CPU, report "no GPU/accelerator available; falling back to `qmd search` + local text search" and skip `qmd query`
- Never silently block on `qmd query` waiting for completion
- If the tier fails (command error, timeout, cache not writable, ...), treat the `zatools qmd` path as unavailable for this run, fall back to tier 1 / tier 2, and state "degraded retrieval" in the answer

## Invocation constraints

### Single entry point

- Run retrieval and maintenance commands exclusively through `zatools qmd ...`; do not call other implementations directly
- In sandboxed agents such as Codex or Claude Code, verify before running:
  - the agent has permission to execute `zatools qmd ...`
  - the project-root `.cache/` directory is writable

If either check fails, treat the run as "reduced validation coverage" and apply the fallback rules above.

### Model flags must be passed explicitly

When the task is inside a DevWiki workspace:

1. Read `embed_model`, `rerank_model`, and `generate_model` from `config/search.yaml`
2. Append them explicitly on every `zatools qmd ...` command with `--embed-model`, `--rerank-model`, and `--generate-model`
3. If any value is missing, fall back to the CLI built-in defaults

Do not rely on the current working directory being the DevWiki root.

### Don'ts

- Do not use `zatools qmd status` as a prerequisite probe. For retrieval tasks, run the target command directly; only treat it as unavailable when that command fails
- Do not treat `zatools qmd` hits as ground truth — it is only a recall accelerator
- Do not auto-run `embed` after every sync; trigger it on demand
