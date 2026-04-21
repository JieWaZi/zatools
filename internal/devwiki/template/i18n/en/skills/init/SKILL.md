---
name: "devwiki-init"
description: "Use when bootstrapping the first DevWiki knowledge skeleton for a single-product repo from existing raw documents, especially after setup or when structured capability and feature pages do not exist yet."
argument-hint: "[scope]"
---

# /devwiki-init

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Bootstrap the first DevWiki from `raw/`. The goal is not to fully digest everything at once, but to create a credible first-pass capabilities / features skeleton, then present an `init proposal` for confirmation.

## Inputs

- `scope` (optional): initialization scope, such as “user management docs only” or “all raw documents”
- `raw/*/*.md` — raw requirements, designs, feature notes, and test plans
- `config/project.yaml` — primary code directory, language, and code repo configuration

## Outputs

- `wiki/capabilities/*.md` — capability pages identified during initialization
- `wiki/features/*.md` — feature pages identified during initialization
- `wiki/index.md` — updated catalog
- `wiki/log.md` — init proposal and apply logs
- one `init proposal` for user confirmation

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory and language configuration
- `raw/*/*.md` — source of truth
- `wiki/index.md` — avoid duplicate initialization when content already exists
- `wiki/capabilities/*.md` — avoid duplicate capability creation
- `wiki/features/*.md` — avoid duplicate feature creation
- local code directory — optional first-pass code clue scan for feature candidates

### Writes

- CREATE `wiki/capabilities/*.md`
- CREATE `wiki/features/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: Check initialization preconditions

1. Confirm the current directory is the DevWiki root and setup has already been run
2. Confirm `raw/` contains at least some Markdown source files
3. If `wiki/capabilities/` and `wiki/features/` already contain substantial content, decide whether this is truly a first initialization
4. If the repo already looks initialized, do not overwrite blindly; explain the state and prefer `/devwiki-ingest` or `/devwiki-refresh`

### Step 2: Scan raw documents and build feature candidates

1. Traverse `raw/*/*.md`
2. Extract for each file:
   - title
   - likely feature topic
   - source kind inferred from the parent directory
   - `source_path`
   - `source_hash`
3. Recompute `source_hash` from current file contents
4. If a document has no level-1 heading, fall back to its filename and mark it as a lower-quality source in the proposal
5. Group related raw sources into feature candidates targeting `wiki/features/<feature-slug>.md`
6. Do not create document mirror pages; raw files remain the source layer

### Step 3: Derive capability candidates

1. Extract candidate capabilities from titles, headings, repeated business terms, and recurring product workflows
2. Keep the capability view business-first: capability pages should answer what the system can do, not how it is implemented
3. If multiple raw sources point to the same capability, merge them into one candidate
4. Map each capability candidate to one or more feature candidates
5. If capability boundaries are unclear, preserve that uncertainty instead of forcing a decision

### Step 4: Run the first code clue scan

Follow the tiered recall rules in `references/zatools-qmd.md`, **local-first by default**:

1. Read the primary code directory from `config/project.yaml`
2. If a code directory is configured:
   - start with local `grep` / file search for known anchors
   - escalate to `zatools qmd search` only when local hits are insufficient
   - avoid `zatools qmd query` during initialization; only consider it when concept-level recall is genuinely needed AND local acceleration is available
3. Keep this scan lightweight; the goal is initial `code_refs`, `api_entries`, and `test_refs` for feature candidates, not full call-chain tracing
4. Deep tracing belongs to `/devwiki-ask` and `/devwiki-feature-doc`
5. Initial code clues must carry confidence; low-confidence hits must not be treated as primary refs

### Step 5: Build the init proposal

The `init proposal` must include:

- which capabilities will be created
- which features will be created
- how raw documents map to feature pages
- how features map to capabilities
- which code refs are only clues versus higher-confidence refs
- which points need user confirmation

Also separate actions by risk:

- low risk: logs, index refresh, source hash capture
- medium risk: attaching a feature to an existing capability, adding auxiliary code clues
- high risk: creating capabilities, creating features, writing primary code refs

### Step 6: Wait for confirmation

1. All medium- and high-risk actions must wait for confirmation
2. If multiple capability boundaries are unclear or code hits are too scattered, do not write yet
3. In those cases, ask 1 to 3 concrete questions, then revise the proposal
4. Do not create capability or feature pages before confirmation

### Step 7: Apply and update navigation

After confirmation:

1. write `wiki/capabilities/`
2. write `wiki/features/`
3. update `wiki/index.md`
4. append `init | proposal-applied` to `wiki/log.md`
5. after writes succeed, run:

```bash
zatools qmd update
zatools qmd status
```

6. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

## Constraints

- **raw/ is read-only**: do not modify files under `raw/`
- **No fabrication**: never force capabilities or features without raw or code evidence
- **source_hash must be real**: it must reflect current source content
- **Medium/high risk requires confirmation**: especially new capabilities, new features, and primary code refs
- **Initialization is not full digestion**: `/devwiki-init` creates the first skeleton; it does not replace later `/devwiki-ingest`
- **Keep code scanning lightweight**: first pass only, no deep call-chain tracing here

## Error Handling

- **raw is empty**: tell the user to add raw documents first
- **wiki already has substantial content**: explain that initialization likely already happened and suggest `/devwiki-ingest` or `/devwiki-refresh`
- **code directory not configured**: allow raw-only initialization but state clearly that code linkage was not verified
- **`zatools qmd ...` unavailable**: fall back to local keyword search without blocking init
- **multiple capability candidates conflict**: stop expanding the search and ask the user
