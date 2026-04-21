---
name: "devwiki-refresh"
description: "Use when existing DevWiki knowledge has drifted from current raw documents, capability-feature mapping, code paths, symbols, or test/API entry points."
argument-hint: "[drift scope or issue description]"
---

# /devwiki-refresh

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Repair DevWiki knowledge drift and user-reported errors. `refresh` is proposal-first by default and should not silently rewrite high-impact knowledge.

## Inputs

- `scope` (optional): drift scope or issue description, such as “user-permission feature drift” or “all broken code refs”
- `wiki/capabilities/*.md`
- `wiki/features/*.md`
- `raw/*/*.md`
- `config/project.yaml`

## Outputs

- one `refresh proposal`
- lists of capabilities / features needing repair
- lists of broken `code_refs`, stale `sources.hash`, missing `symbol`, invalid paths, and bad capability-feature links
- confirmed repair results after user approval
- updated `wiki/index.md`
- appended refresh logs in `wiki/log.md`

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory
- `wiki/capabilities/*.md` — verify capability summary, feature links, and boundaries
- `wiki/features/*.md` — verify `sources`, `api_entries`, `code_refs`, and `test_refs`
- `wiki/index.md` — cross-check page existence
- `raw/*/*.md` — compare against current raw sources
- local code directory — verify code-ref paths and whether a `symbol` still exists

### Writes

- proposal-first by default, no immediate writes
- only after confirmation:
  - EDIT `wiki/capabilities/*.md`
  - EDIT `wiki/features/*.md`
  - EDIT `wiki/index.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: Identify the refresh scope

1. Decide whether this is a global refresh or a targeted repair for one capability, one feature, or one source path
2. Classify the issue:
   - stale `sources.hash`
   - missing raw source
   - broken `code_refs.path`
   - missing `symbol`
   - stale `api_entries` or `test_refs`
   - incorrect capability-feature mapping
   - capability summary drift
3. If the user already reported a concrete mistake, stay tightly scoped to that mistake instead of scanning everything blindly

### Step 2: Run deterministic drift checks

Follow the tiered recall rules in `references/zatools-qmd.md`, **local-first by default**:

1. Re-verify `sources.path` and `sources.hash` for feature pages
2. Check whether feature `code_refs`, `api_entries`, and `test_refs` still point to valid entry points
3. If a code directory is configured, verify whether each referenced `symbol` still exists in its file
4. Verify capability-feature reverse links
5. When re-running retrieval, escalate in tiers:
   - start with local `grep` / file search
   - escalate to `zatools qmd search` when local hits are insufficient
   - only escalate to `zatools qmd query` when concept-level recall is needed; apply the hard fallback when no GPU/accelerator is available
6. Separate findings into:
   - deterministic breakage
   - high-probability drift
   - low-confidence inference

### Step 3: Re-retrieve candidate repairs

1. For missing raw sources, search `raw/` for renamed, moved, or updated files
2. For broken code refs, search the codebase for replacement paths and symbols
3. For suspicious capability drift, re-check linked feature pages and raw source material
4. For suspicious feature drift, re-check source material, entry points, and bounded code evidence
5. If a feature page is clearly too stale to repair incrementally, recommend `/devwiki-feature-doc`

### Step 4: Build the refresh proposal

The `refresh proposal` must include:

- which findings are deterministic breakage
- which findings are high-probability drift
- which findings remain low-confidence inference
- recommended actions:
  - update `sources.hash`
  - replace `sources.path`
  - replace `code_refs.path`
  - remove missing `symbol`
  - repair capability-feature links
  - tighten or simplify stale summaries

Separate by risk:

- low risk: `sources.hash` refresh, removing obviously broken auxiliary clues, index updates
- medium risk: replacing `code_refs`, `api_entries`, or `test_refs`; relinking an obvious capability-feature edge
- high risk: changing capability boundaries, creating or deleting a feature page, replacing primary code refs

### Step 5: Wait for confirmation

1. Medium- and high-risk repairs require confirmation
2. If multiple candidate paths or capability mappings remain plausible, do not decide silently
3. Ask 1 to 3 specific questions instead
4. Only apply changes after the proposal is confirmed

### Step 6: Apply repairs and record results

After confirmation:

1. update the relevant capabilities / features
2. update `wiki/index.md`
3. append `refresh | proposal-applied` to `wiki/log.md`
4. after writes succeed, run:

```bash
zatools qmd update
zatools qmd status
```

5. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

6. if additional deterministic breakage is exposed, recommend `/devwiki-check`

## Constraints

- **raw/ is read-only**: do not modify source documents
- **Proposal-first by default**: do not silently apply high-impact repairs
- **No fabricated symbols**: never rewrite a function, file, or symbol you could not verify
- **Medium/high risk requires confirmation**: especially capability boundary changes, feature re-linking, and primary `code_refs`
- **Separate facts from inference**: deterministic breakage, high-probability drift, and low-confidence inference must be reported separately
- **Admit uncertainty**: weak evidence is not enough to justify forced repairs

## Error Handling

- **wiki is mostly empty**: tell the user to run `/devwiki-init` or `/devwiki-ingest` first
- **code directory missing**: allow wiki/raw-layer refresh only, and say code was not verified
- **`zatools qmd ...` unavailable**: fall back to local search without blocking refresh
- **source target deleted**: report deterministic breakage and search `raw/` for replacements
- **multiple candidates remain plausible**: stop expanding and ask the user
