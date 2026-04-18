---
name: "devwiki-edit"
description: "Use when the user already knows what DevWiki content should be changed, especially for targeted page updates, metadata edits, adding new raw-source entries, or applying confirmed structured edits."
argument-hint: "[edit-request]"
---

# /devwiki-edit

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Make targeted DevWiki edits. `edit` is for cases where the user already knows what should change; use `/devwiki-refresh` instead when the issue is drift, broken refs, or misclassification.

## Inputs

- `request`: explicit edit request
- optional local file path, URL, or target wiki page path

## Outputs

- updated `wiki/` pages or newly added `raw/` source entries
- updated `wiki/index.md`
- updated `wiki/log.md`
- follow-up suggestions when relevant: `/devwiki-ingest` or `/devwiki-refresh`

## DevWiki Interaction

### Reads

- user-specified `wiki/` pages
- `wiki/index.md`
- related source material under `raw/`
- `config/project.yaml` when code-directory context is needed

### Writes

- CREATE / EDIT `wiki/**/*.md`
- add new files under `raw/` after confirmation when needed
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: Parse user intent

Classify the request into:

1. editing existing wiki pages
2. adding new raw-source entries
3. deleting or replacing incorrect content

If the request is really about drift, broken paths, or wrong classification, note that `/devwiki-refresh` is likely the better path.

### Step 2: Determine the edit boundary

1. lock the exact target page or target directory
2. if the target is ambiguous, ask for confirmation instead of guessing
3. if multiple pages are affected, list them before editing
4. if structural relations are affected, check whether reverse links or index entries also need updates

### Step 3: Apply the edits

1. **Edit wiki pages**
   - change only the requested fields or sections
   - preserve the page template instead of rewriting everything
   - if a forward link is added, maintain the reverse relation as well
2. **Add new raw-source entries**
   - adding files or source entries under `raw/` is allowed
   - newly added raw content does not enter the wiki automatically; recommend `/devwiki-ingest`
3. **Delete or replace content**
   - medium/high-risk deletions require confirmation
   - edits that affect primary capability or change ownership require confirmation first

### Step 4: Update navigation and logs

1. if title, slug, or classification changes, update `wiki/index.md`
2. append to `wiki/log.md`:
   - `edit | wiki-updated | <summary>`
   - or `edit | raw-added | <summary>`
3. if this edit changed `wiki/` pages or added files under `raw/`, run:

```bash
zatools qmd update
zatools qmd status
```

4. if the next task immediately depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still reports pending embeddings, ask whether to continue with:

```bash
zatools qmd embed
```

### Step 5: Recommend next steps

- new raw material added: recommend `/devwiki-ingest`
- existing knowledge disagrees with reality: recommend `/devwiki-refresh`
- broader capability documentation is needed: recommend `/devwiki-feature-doc`

## Constraints

- **Do not expand beyond the requested scope**
- **Existing raw files are not rewritten by default**: additions are allowed, replacement or deletion requires confirmation
- **Preserve structured page templates**
- **Edits affecting primary capability or change ownership require confirmation**
- **Link relationships must stay synchronized**

## Error Handling

- **target page missing**: tell the user first, then confirm whether to create it
- **request too vague**: ask for confirmation instead of guessing
- **cross-page high-impact edits required**: stop direct editing and recommend `/devwiki-refresh`
