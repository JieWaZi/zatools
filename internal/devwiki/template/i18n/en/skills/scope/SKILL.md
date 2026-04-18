---
name: "devwiki-scope"
description: "Use when an engineering change must be scoped before implementation, especially to decide whether it is new or modify, retrieve related history, locate candidate code, and ask focused follow-up questions when confidence stays low."
argument-hint: "<change-query>"
---

# /devwiki-scope

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Scope an engineering change before writing design docs, editing code, or generating new wiki content. `/devwiki-scope` is read-first and should not rewrite the wiki by default.

## Inputs

- `query`: the change description, with at least one concrete target
- Optional anchors:
  - capability or module name
  - API URL
  - key file
  - key function or class
  - page path or route
  - ticket or requirement id
- `config/project.yaml`: used to read the primary code directory

## Outputs

- A scope report
- `classification`: `new`, `modify`, or `unclear`
- Related documents, capabilities, and changes
- Top-K candidate code files and symbols
- 1 to 3 concrete user questions when confidence is still low

## DevWiki Interaction

### Reads

- `config/project.yaml`
- `wiki/documents/`
- `wiki/capabilities/`
- `wiki/changes/`
- `wiki/index.md`
- `raw/`
- The local code directory

### Writes

- No page writes by default
- Only when the user explicitly wants an artifact, allow a low-risk output under `wiki/outputs/`


## Workflow

### Step 1: Narrow the problem boundary

Compress the request into a searchable scope. Confirm:

- what is changing
- whether it looks more like a business capability or a system capability
- whether there is at least one entry anchor

If the user only says something vague like “check how this requirement should be done,” ask 1 to 3 focused questions first instead of searching blindly.

### Step 2: Retrieve across the three collections

Retrieve candidates according to `references/zatools-qmd.md` across:

- `raw`
- `wiki`
- `code`

Look for:

- relevant requirements, designs, feature docs, code summaries, and postmortems under `wiki/documents/`
- related pages in `wiki/capabilities/`
- historical records in `wiki/changes/`
- useful but not-yet-structured material under `raw/`
- top-K candidate files in the code directory

### Step 3: Merge structured evidence

Group the first-pass results into three evidence buckets:

- existing capabilities
- historical changes
- documents and raw sources

Explain why each item is relevant instead of dumping filenames only.

### Step 4: Infer `new / modify / unclear`

At minimum, consider:

- whether an existing capability matches
- whether historical design or requirement docs match
- whether an existing change matches
- whether current code contains an explainable implementation

Suggested interpretation:

- `modify`: existing capability, docs, or code strongly match
- `new`: no meaningful match to current capability or implementation, but the target is clear
- `unclear`: evidence conflicts, is too scattered, or the entry anchor is missing

### Step 5: Perform local code verification on the top-K candidates

Do not stop at filename similarity. Verify:

- why the file is relevant
- whether the key function, class, or symbol exists
- whether it is a primary implementation, entry point, or just a clue
- whether there is a better entry file closer to the real flow

`/devwiki-scope` should inspect code carefully, but it does not need the full call-chain depth of `/devwiki-feature-doc`.

### Step 6: Produce the scope report

The scope report should include:

- whether the change looks like `new`, `modify`, or `unclear`
- the related capability, change, document, and raw evidence
- top-K code candidates with rationale
- the current evidence gaps
- next-step suggestions:
  - use `/devwiki-ask` for general knowledge lookup
  - use `/devwiki-feature-doc` for a full implementation trace
  - use `/devwiki-refresh` when the wiki is clearly drifting from reality

### Step 7: Ask the user when confidence stays low

After a few bounded search rounds, stop expanding and ask the user 1 to 3 concrete questions. Prefer asking for:

- API URL
- key file
- key function
- page path
- known capability name

## Constraints

- **`zatools qmd ...` is a retrieval accelerator, not the source of truth**
- **Separate facts from inference**: page existence, file existence, and symbol existence are facts; `new / modify / unclear` is inference
- **Read-only by default**: do not casually edit the wiki during `/devwiki-scope`
- **Search must stay bounded**: stop and ask instead of searching forever
- **Code candidates must be verified**: keyword hits are not conclusions
- **No fabrication of capabilities or implementation relationships**

## Error Handling

- **Wiki is mostly empty**: suggest `/devwiki-init` or `/devwiki-ingest` first
- **Code directory missing or not configured**: still produce a document-only scope result, but say code verification was skipped
- **`zatools qmd ...` unavailable**: fall back to local search and explain the quality drop
- **Too many scattered hits**: stop broadening the search and ask the user for an anchor
- **No usable evidence found**: report the empty result and request at least one entry anchor
