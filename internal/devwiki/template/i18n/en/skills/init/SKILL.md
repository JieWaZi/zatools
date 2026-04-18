---
name: "devwiki-init"
description: "Use when bootstrapping the first DevWiki knowledge skeleton for a single-product repo from existing raw documents, especially after setup or when structured documents, capabilities, and changes do not exist yet."
argument-hint: "[scope]"
---

# /devwiki-init

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Bootstrap the first DevWiki from `raw/`. The goal is not to fully digest everything at once, but to create a credible first-pass skeleton for documents, capabilities, and changes, then present an `init proposal` for confirmation.

## Inputs

- `scope` (optional): initialization scope, such as “user management docs only” or “all raw documents”
- `raw/*/*.md` — raw requirements, designs, feature notes, code summaries, postmortems, API docs, test plans
- `config/project.yaml` — primary code directory, language, and code repo configuration

## Outputs

- `wiki/documents/**/*.md` — structured mirror pages for raw documents
- `wiki/capabilities/*.md` — capability pages identified during initialization
- `wiki/changes/*.md` — change pages identified during initialization
- `wiki/index.md` — updated catalog
- `wiki/log.md` — init proposal and apply logs
- one `init proposal` for user confirmation

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory and language configuration
- `raw/*/*.md` — source of truth
- `wiki/index.md` — avoid duplicate initialization when content already exists
- `wiki/documents/**/*.md` — skip already mirrored documents
- `wiki/capabilities/*.md` — avoid duplicate capability creation
- `wiki/changes/*.md` — avoid duplicate change creation
- local code directory — optional first-pass code clue scan

### Writes

- CREATE `wiki/documents/**/*.md`
- CREATE `wiki/capabilities/*.md`
- CREATE `wiki/changes/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: Check initialization preconditions

1. Confirm the current directory is the DevWiki root and setup has already been run
2. Confirm `raw/` contains at least some Markdown source files
3. If `wiki/documents/`, `wiki/capabilities/`, and `wiki/changes/` already contain substantial content, decide whether this is truly a first initialization
4. If the repo already looks initialized, do not overwrite blindly; explain the state and prefer `/devwiki-ingest` or `/devwiki-refresh`

### Step 2: Scan raw documents and build document candidates

1. Traverse `raw/*/*.md`
2. Extract for each file:
   - title
   - `doc_type`
   - `source_path`
   - `source_hash`
3. Recompute `source_hash` from current file contents
4. If a document has no level-1 heading, fall back to its filename and mark it as a lower-quality source in the proposal
5. Build document candidates targeting `wiki/documents/<type>/`

### Step 3: Derive capability candidates

1. Extract candidate capabilities from titles, headings, repeated product terms, and recurring technical phrases
2. Cover both user-visible and system capabilities where evidence exists
3. Stay conservative: do not over-split capabilities just to look complete
4. If multiple documents point to the same capability, merge them into one candidate
5. If capability boundaries are unclear, preserve that uncertainty instead of forcing a decision

### Step 4: Derive change candidates

1. Extract change candidates from language such as add, modify, migrate, refactor, replace, fix
2. Decide whether each candidate looks more like:
   - first-time introduction
   - modification of an existing capability
   - repair of a known problem
3. If a change is only weakly implied by one document, do not overstate it as a confirmed fact
4. Early change pages are provisional records and will be corrected later by `/devwiki-ingest` and `/devwiki-refresh`

### Step 5: Run the first code clue scan

1. Read the primary code directory from `config/project.yaml`
2. If a code directory is configured:
   - run `zatools qmd status` first
   - if `zatools qmd status` is healthy, prefer `zatools qmd query` across `raw / wiki / code`, then inspect top-K hits with `zatools qmd get` / `zatools qmd multi-get`
   - otherwise fall back to local keyword search
3. Keep this scan lightweight; the goal is initial `code_refs`, not full call-chain tracing
4. Deep tracing belongs to `/devwiki-scope`, `/devwiki-ask`, and `/devwiki-feature-doc`
5. Initial code clues must carry confidence; low-confidence hits must not be treated as primary refs

### Step 6: Build the init proposal

The `init proposal` must include:

- which document mirror pages will be created
- which capabilities will be created
- which changes will be created
- how raw documents map to capabilities and changes
- which code refs are only clues versus higher-confidence refs
- which points need user confirmation

Also separate actions by risk:

- low risk: new document mirror pages, logs, index refresh
- medium risk: attaching docs to existing capabilities, adding auxiliary code clues
- high risk: creating capabilities, creating changes, writing primary code refs

### Step 7: Wait for confirmation

1. All medium- and high-risk actions must wait for confirmation
2. If multiple capability boundaries are unclear, multiple change candidates conflict, or code hits are too scattered, do not write yet
3. In those cases, ask 1 to 3 concrete questions, then revise the proposal
4. Do not create capability or change pages before confirmation

### Step 8: Apply and update navigation

After confirmation:

1. write `wiki/documents/`
2. write or update `wiki/capabilities/`
3. write or update `wiki/changes/`
4. update `wiki/index.md`
5. append `init | proposal-applied` to `wiki/log.md`
6. after writes succeed, run:

```bash
zatools qmd update
zatools qmd status
```

7. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

## Constraints

- **raw/ is read-only**: do not modify files under `raw/`
- **No fabrication**: never force capabilities or changes without raw or code evidence
- **source_hash must be real**: it must reflect current source content
- **Medium/high risk requires confirmation**: especially new capabilities, new changes, and primary code refs
- **Initialization is not full digestion**: `/devwiki-init` creates the first skeleton; it does not replace later `/devwiki-ingest`
- **Keep code scanning lightweight**: first pass only, no deep call-chain tracing here

## Error Handling

- **raw is empty**: tell the user to add raw documents first
- **wiki already has substantial content**: explain that initialization likely already happened and suggest `/devwiki-ingest` or `/devwiki-refresh`
- **code directory not configured**: allow document-only initialization but state clearly that code linkage was not verified
- **`zatools qmd ...` unavailable**: fall back to local keyword search without blocking init
- **multiple capability candidates conflict**: stop expanding the search and ask the user
