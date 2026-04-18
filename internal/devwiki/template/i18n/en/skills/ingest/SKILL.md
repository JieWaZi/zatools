---
name: "devwiki-ingest"
description: "Use when bringing one or more new raw documents into DevWiki and deciding how they should map to documents, capabilities, changes, and code refs, especially for incremental knowledge maintenance."
argument-hint: "<document-path-or-scope>"
---

# /devwiki-ingest

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Bring new raw documents into the DevWiki proposal flow. `ingest` incrementally absorbs source material, matches it against existing knowledge, adds initial code clues, and updates the wiki after confirmation.

## Inputs

- `source`: a document path, a directory, or a batch of raw documents
- `raw/*/*.md` — new or updated source material
- `config/project.yaml` — primary code directory, language, and code repo configuration

## Outputs

- created or updated `wiki/documents/**/*.md`
- created or updated `wiki/capabilities/*.md`
- created or updated `wiki/changes/*.md`
- written or corrected `code refs`
- updated `wiki/index.md`
- appended ingest entries in `wiki/log.md`
- one ingest proposal for user confirmation

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory
- `raw/*/*.md` — source documents to ingest
- `wiki/documents/**/*.md` — dedup by `source_path` and `source_hash`
- `wiki/capabilities/*.md` — match existing capabilities
- `wiki/changes/*.md` — match existing changes
- `wiki/index.md` — locate existing pages
- local code directory — add candidate code locations and perform second-pass verification

### Writes

- CREATE / EDIT `wiki/documents/**/*.md`
- CREATE / EDIT `wiki/capabilities/*.md`
- CREATE / EDIT `wiki/changes/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: Parse sources and check deduplication

1. Expand the input `source` into a concrete document list
2. Extract for each document:
   - title
   - `doc_type`
   - `source_path`
   - `source_hash`
3. Check for existing records using `source_path`, `source_hash`, and title
4. If the source is unchanged, mark it in the proposal as “no update needed”
5. If `source_path` matches but `source_hash` changed, treat it as an update to an existing document mirror

### Step 2: Match existing documents, capabilities, and changes

1. Match the document itself against existing mirrors
2. Match the most likely capability
3. Match the most likely change record
4. Prioritize:
   - titles and aliases
   - already linked documents
   - capability `code_refs`
   - change classification and linked scope
5. If several capabilities look plausible, keep multiple candidates instead of forcing one

### Step 3: Retrieve candidate code clues

1. If a code directory is configured:
   - run `zatools qmd status` first
   - if `zatools qmd status` is healthy, prefer `zatools qmd query` across `raw / wiki / code`, then inspect top-K hits with `zatools qmd get` / `zatools qmd multi-get`
   - then run local keyword search in code
2. Perform a second-pass inspection on top-K candidate files
3. Confirm at least:
   - why the file is relevant
   - whether a key function or symbol actually exists
   - whether the hit is a primary ref or only an auxiliary clue
4. If hits are too scattered and cannot be localized, do not write high-confidence `code refs`

### Step 4: Build the ingest proposal

The proposal must answer:

- should this create a new document mirror
- should it update an existing document mirror
- should it attach to an existing capability
- should it create a new capability
- should it update an existing change
- should it create a new change
- should it write or adjust code refs

Risk separation is mandatory:

- low risk: new document mirror, `source_hash` refresh, logs
- medium risk: attaching to an existing capability, adding auxiliary code refs, linking to an existing change
- high risk: creating capabilities, creating changes, changing primary code refs, changing change classification

### Step 5: Ask the user when confidence remains low

After a few bounded searches, ask the user when:

- one document plausibly maps to multiple capabilities
- several change records look equally plausible
- the referenced endpoint, module, or function cannot be found in code
- code hits are only scattered clues with no interpretable cluster

Question rules:

- ask only 1 to 3 narrow questions
- prefer anchor questions: API URL, key file, key function, page route, or requirement ticket
- do not ask vague questions like “please provide more context”

### Step 6: Wait for confirmation

1. All medium- and high-risk writes require confirmation
2. These actions must never be applied silently:
   - creating a capability
   - creating a change
   - remapping a document to a different capability
   - writing primary code refs
3. Only apply changes after the proposal is confirmed

### Step 7: Apply and refresh navigation

After confirmation:

1. write or update `wiki/documents/`
2. write or update `wiki/capabilities/`
3. write or update `wiki/changes/`
4. update `wiki/index.md`
5. append `ingest | proposal-applied` to `wiki/log.md`
6. after writes succeed, run:

```bash
zatools qmd update
zatools qmd status
```

7. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

When useful, recommend follow-up steps:

- use `/devwiki-scope` for change classification
- use `/devwiki-feature-doc` when code exists but documentation is still missing
- use `/devwiki-refresh` when wiki knowledge and current code disagree

## Constraints

- **raw/ is read-only**: do not modify source documents
- **No fabricated code refs**: never write high-confidence file or function refs without checking code
- **Do not silently merge capabilities**: preserve ambiguity until confirmed
- **Medium/high risk requires confirmation**: especially capability, change, and primary code-ref writes
- **source_hash must be refreshed**: document content changes require a new hash
- **Code clues are not full design docs**: `ingest` adds local evidence but does not replace `/devwiki-feature-doc`

## Error Handling

- **source missing or empty**: ask the user for a valid raw document path
- **document type cannot be inferred**: ask the user to place it in the right raw directory or convert it first
- **`zatools qmd ...` unavailable**: fall back to local search without blocking ingest
- **code directory missing**: allow document-only ingest, but state clearly that code was not verified
- **multiple capability / change matches remain**: stop expanding and ask the user
