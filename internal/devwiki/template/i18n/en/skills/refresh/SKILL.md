---
name: "devwiki-refresh"
description: "Use when existing DevWiki knowledge has drifted from current raw documents, code paths, symbols, or capability/change classification, especially after user corrections, source_hash mismatches, broken code refs, missing symbols, or outdated change classification."
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

- `scope` (optional): drift scope or issue description, such as “user-permission capability drift” or “all broken code refs”
- `wiki/documents/**/*.md`
- `wiki/capabilities/*.md`
- `wiki/changes/*.md`
- `raw/*/*.md`
- `config/project.yaml`

## Outputs

- one `refresh proposal`
- lists of documents / capabilities / changes needing repair
- lists of broken `code refs`, stale `source_hash`, missing `symbol`, and invalid paths
- confirmed repair results after user approval
- updated `wiki/index.md`
- appended refresh logs in `wiki/log.md`

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory
- `wiki/documents/**/*.md` — verify `source_path` and `source_hash`
- `wiki/capabilities/*.md` — verify capability classification, linked docs, and code refs
- `wiki/changes/*.md` — verify change classification and linked entities
- `wiki/index.md` — cross-check page existence
- `raw/*/*.md` — compare against current raw sources
- local code directory — verify code-ref paths and whether a `symbol` still exists

### Writes

- proposal-first by default, no immediate writes
- only after confirmation:
  - EDIT `wiki/documents/**/*.md`
  - EDIT `wiki/capabilities/*.md`
  - EDIT `wiki/changes/*.md`
  - EDIT `wiki/index.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: Identify the refresh scope

1. Decide whether this is a global refresh or a targeted repair for one capability, change, or source path
2. Classify the issue:
   - `source_hash` mismatch
   - broken `source_path`
   - broken `code refs` path
   - missing `symbol`
   - incorrect capability classification
   - incorrect `new / modify / unclear` change classification
3. If the user already reported a concrete mistake, stay tightly scoped to that mistake instead of scanning everything blindly

### Step 2: Run deterministic drift checks

1. Re-verify `source_path` and `source_hash` for document mirrors
2. Check whether capability and change `code refs` paths still exist
3. If a code directory is configured, verify whether each referenced `symbol` still exists in its file
4. Run `zatools qmd status`
5. If `zatools qmd status` is healthy, use `zatools qmd query` across `wiki / raw / code`, then inspect top-K hits with `zatools qmd get` / `zatools qmd multi-get` to narrow the repair scope
6. Separate findings into:
   - deterministic breakage
   - high-probability drift
   - low-confidence inference

### Step 3: Re-retrieve candidate repairs

1. For broken documents, search `raw/` for renamed, moved, or updated source files
2. For broken code refs, search the codebase for replacement paths and symbols
3. For suspicious capability classification, re-check linked documents, changes, and code refs
4. For suspicious change classification, re-evaluate whether it is `new`, `modify`, or `unclear`
5. If documents are clearly lagging behind code behavior, recommend `/devwiki-feature-doc` where appropriate

### Step 4: Build the refresh proposal

The `refresh proposal` must include:

- which findings are deterministic breakage
- which findings are high-probability drift
- which findings remain low-confidence inference
- recommended actions:
  - update `source_hash`
  - replace `source_path`
  - replace `code refs.path`
  - remove missing `symbol`
  - repair capability classification
  - repair change classification

Separate by risk:

- low risk: `source_hash` refresh, removing obviously broken auxiliary clues, index updates
- medium risk: replacing `code refs`, adding better code candidates
- high risk: changing capability classification, changing change classification, replacing primary code refs

### Step 5: Wait for confirmation

1. Medium- and high-risk repairs require confirmation
2. If multiple candidate paths, capability mappings, or change classifications remain plausible, do not decide silently
3. Ask 1 to 3 specific questions instead
4. Only apply changes after the proposal is confirmed

### Step 6: Apply repairs and record results

After confirmation:

1. update the relevant documents / capabilities / changes
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
- **Medium/high risk requires confirmation**: especially capability classification, change classification, and primary `code refs`
- **Separate facts from inference**: deterministic breakage, high-probability drift, and low-confidence inference must be reported separately
- **Admit uncertainty**: weak evidence is not enough to justify forced repairs

## Error Handling

- **wiki is mostly empty**: tell the user to run `/devwiki-init` or `/devwiki-ingest` first
- **code directory missing**: allow document-layer refresh only, and say code was not verified
- **`zatools qmd ...` unavailable**: fall back to local search without blocking refresh
- **source_path target deleted**: report deterministic breakage and search `raw/` for replacements
- **multiple candidates remain plausible**: stop expanding and ask the user
