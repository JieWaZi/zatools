---
name: "devwiki-ingest"
description: "Use when bringing one or more new raw documents into DevWiki and deciding how they should update capabilities, features, and feature-level code clues."
argument-hint: "<document path or scope>"
---

# /devwiki-ingest

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Bring new raw documents into the DevWiki proposal flow. `ingest` incrementally absorbs source material, matches it against existing knowledge, adds initial code clues, and updates the wiki after confirmation.

## Inputs

- `source`: one document path, one directory, or a batch of raw documents
- `raw/*/*.md` — added or updated source documents
- `config/project.yaml` — primary code directory, language, and code repo configuration

## Outputs

- created or updated `wiki/capabilities/*.md`
- created or updated `wiki/features/*.md`
- updated feature-level `sources`, `code_refs`, `api_entries`, and `test_refs`
- updated `wiki/index.md`
- ingest record appended to `wiki/log.md`
- one ingest proposal for user confirmation

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory
- `raw/*/*.md` — source documents to ingest
- `wiki/features/*.md` — dedup by `sources.path` and `sources.hash`
- `wiki/capabilities/*.md` — match existing capabilities
- `wiki/index.md` — locate existing pages
- local code directory — supplement candidate code locations and bounded second-pass verification

### Writes

- CREATE / EDIT `wiki/capabilities/*.md`
- CREATE / EDIT `wiki/features/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: Expand the source set and deduplicate

1. Expand `source` into the concrete document list
2. For each file, extract:
   - title
   - likely feature topic
   - source kind inferred from the parent directory
   - `source_path`
   - `source_hash`
3. Check existing feature pages by three signals:
   - matching `sources.path`
   - matching `sources.hash`
   - close title / feature-name similarity
4. If `sources.path` matches but the hash changed, treat it as an update to an existing feature page
5. If a raw file is new but clearly belongs to an existing feature, update that feature instead of forcing a new page

### Step 2: Match existing capabilities and features

1. First match the most likely feature page
2. Then match the most likely supporting capability pages
3. Match using:
   - titles and aliases
   - existing capability-feature links
   - existing feature `code_refs`, `api_entries`, and `test_refs`
   - repeated business terminology in the raw content
4. If multiple capabilities look plausible, keep multiple candidates instead of forcing one
5. If the raw source suggests a genuinely new feature, keep that as a proposal instead of writing immediately

### Step 3: Retrieve candidate code clues

Follow the tiered recall rules in `references/zatools-qmd.md`, **local-first by default**:

1. If a code directory is configured:
   - start with local `grep` / file search for known anchors (symbols, files, directories, API URLs)
   - escalate to `zatools qmd search` only when local hits are insufficient
   - only consider `zatools qmd query` when concept-level recall is genuinely needed; apply the hard fallback when no GPU/accelerator is available
2. Re-check top-K candidate files locally
3. Confirm at least:
   - why the file is relevant
   - whether the key symbol exists
   - whether it is a primary clue or a supporting clue
4. Do not write high-confidence `code_refs` when hits are still scattered

### Step 4: Form the ingest proposal

The proposal must answer:

- should this update an existing feature
- should this create a new feature
- should it link to existing capabilities
- should it create a new capability
- should it update feature-level code clues

Separate actions by risk:

- low risk: source hash refresh, logs, index refresh
- medium risk: attaching a feature to an existing capability, adding auxiliary code clues
- high risk: creating a capability, creating a feature, rewriting primary `code_refs`

### Step 5: Ask follow-up questions when confidence stays low

After a few bounded searches, ask the user when:

- one raw document plausibly maps to multiple capabilities
- several feature pages still compete
- the document mentions an endpoint, module, or symbol that cannot be found
- code search returns many scattered hits without a coherent entry point

Question rules:

- ask only 1 to 3 narrow questions
- prefer anchor questions: API URL, key file, key function, page route, requirement ticket
- do not ask vague questions like "please provide more context"

### Step 6: Wait for confirmation

1. All medium- and high-risk writes must wait for confirmation
2. Especially do not default to:
   - creating a new capability
   - creating a new feature
   - reattaching a feature to a different capability
   - writing primary `code_refs`
3. Only apply the proposal after explicit confirmation

### Step 7: Apply writes and refresh navigation

After confirmation:

1. write or update `wiki/capabilities/`
2. write or update `wiki/features/`
3. update `wiki/index.md`
4. append `ingest | proposal-applied` to `wiki/log.md`
5. after writes succeed, run:

```bash
zatools qmd update
zatools qmd status
```

6. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

Useful next-step suggestions:

- use `/devwiki-ask` for change classification or follow-up Q&A
- use `/devwiki-feature-doc` when raw docs are weak but code clearly contains the feature
- use `/devwiki-refresh` when existing wiki knowledge conflicts with the new evidence

## Constraints

- **raw/ is read-only**: do not modify source documents
- **No fabricated code refs**: files and symbols must not be claimed with high confidence unless verified
- **Do not silently re-scope capabilities**: preserve ambiguity until the user confirms
- **Medium/high risk requires confirmation**: especially new capabilities, new features, and primary code refs
- **source hashes must be refreshed**: content changes require new hashes in the feature page
- **ingest is not a full design pass**: it adds knowledge and bounded code clues; it does not replace `/devwiki-feature-doc`

## Error Handling

- **source is empty or path is missing**: ask the user for a valid raw document path
- **document type cannot be inferred**: ask the user to place it under the correct raw directory
- **`zatools qmd ...` unavailable**: fall back to local search without blocking ingest
- **code directory missing**: allow raw-only updates, but state clearly that code linkage was not verified
- **multiple capability candidates remain plausible**: stop expanding and ask the user
