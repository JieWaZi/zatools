---
name: "devwiki-ask"
description: "Use when Codex needs to answer questions over existing DevWiki knowledge, historical changes, raw documents, and related code clues, especially for questions like “how was this designed before”, “which docs are related”, “which files are involved”, or “what does the current implementation do”."
argument-hint: "<question>"
---

# /devwiki-ask

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> General DevWiki question answering. Default mode is answer-only; write back to `wiki/outputs/` only when the user explicitly asks to save the result.

## Inputs

- `question`: natural-language question
- `--format` (optional): output format, default `markdown`, options: `table` / `bullets` / `timeline`
- `--save-output` (optional): use only when the user explicitly wants the answer persisted under `wiki/outputs/`

## Outputs

- **Always**: one synthesized answer with citations
- **Optional**: if explicitly requested, write `wiki/outputs/<query-slug>.md`
- **Included**: related capabilities, changes, documents, code locations, knowledge gaps, and next-step suggestions

## DevWiki Interaction

### Reads

- `config/project.yaml` — get the primary code directory, language, and code repo configuration
- `wiki/index.md` — locate candidate pages
- `wiki/documents/**/*.md` — requirements, designs, feature notes, code summaries, API docs, test plans, postmortems
- `wiki/capabilities/*.md` — capability aggregation pages
- `wiki/changes/*.md` — change history pages
- `wiki/outputs/*.md` — previously saved answers, if any
- `raw/*/*.md` — original source documents when wiki summaries are insufficient
- local code directory — required when the question is about current behavior, file ownership, endpoints, or function responsibilities

### Writes

- No writes by default
- Only when the user explicitly requests persistence:
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: Establish search scope

1. Read `config/project.yaml` to determine the primary code directory
2. Classify the question:
   - historical design
   - document linkage
   - capability ownership
   - change impact
   - code location
   - implementation behavior
3. If the question is really a pre-change scoping task, note that `/devwiki-scope` may be more appropriate

### Step 2: Retrieve candidate evidence

1. Read `wiki/index.md` and match candidate slugs by keywords and capability names
2. Prioritize results from `wiki/documents/`, `wiki/capabilities/`, and `wiki/changes/`
3. Retrieve candidates according to `references/zatools-qmd.md`
4. Keep the candidate set bounded and read top-K first (K ≤ 12)
5. If a wiki page only contains a summary, go back to the underlying `raw/` source when needed

### Step 3: Verify code when needed

The following question types require code verification and must not be answered from wiki/raw alone:

- “Is the current implementation doing this?”
- “Which file or function owns this?”
- “How does this endpoint work now?”
- “Which code paths are likely impacted by this change?”

When verifying code:

1. Start from `code_refs` already attached to capabilities or changes
2. If `code_refs` are insufficient, run targeted code searches in the local repo
3. Confirm at least one concrete layer of evidence: entry file, key function, or a critical call edge
4. If code and docs disagree, state both explicitly

### Step 4: Compose the answer

The answer must include:

1. **Direct answer first** — do not lead with the retrieval process
2. **Evidence citations** for every key conclusion
   - wiki citations: `wiki/capabilities/...`, `wiki/changes/...`, `wiki/documents/...`
   - raw citations: `raw/...`
   - code citations: file paths and symbols when relevant
3. **Layer separation**
   - confirmed facts
   - reasonable inference
   - unresolved gaps
4. **Related items**
   - related capabilities
   - related changes
   - related documents
   - related code refs
5. **Next-step suggestions**
   - use `/devwiki-scope` for change classification
   - use `/devwiki-feature-doc` when code exists but documentation is missing
   - use `/devwiki-refresh` when wiki knowledge has drifted
   - use `/devwiki-ingest` when raw source material is missing from the knowledge base
   - use `/devwiki-check` when deterministic validation is needed

### Step 5: Ask the user when confidence is still low

After a few independent searches, ask the user instead of forcing an answer when:

- multiple candidate capabilities or changes remain plausible
- the referenced endpoint, file, function, or route cannot be found locally
- dynamic registration, config-driven wiring, gateways, or external services block reliable confirmation
- the question is too broad and spans multiple sub-capabilities

Question rules:

- ask only 1 to 3 narrow questions
- do not ask vague questions like “can you provide more context?”
- prefer anchor questions: API URL, key file, key function, page route, or recent change identifier

### Step 6: Persist the result only on explicit request

Only when the user explicitly asks to save the answer:

1. write `wiki/outputs/<query-slug>.md`
2. include the original question, cited sources, conclusions, and unresolved items
3. append `ask | <question-summary> | saved-output` to `wiki/log.md`
4. if the answer reveals stale or incorrect wiki facts, recommend or run `/devwiki-refresh` instead of silently rewriting pages

## Constraints

- **No fabrication**: all key conclusions must come from actual DevWiki content or local code
- **raw/ is read-only**: do not modify files under `raw/`
- **No wiki writes by default**: persist only with explicit user intent
- **Code questions require code checks**: implementation-behavior answers must verify code
- **Separate facts from inference**: do not present inference as confirmed truth
- **Admit uncertainty**: say “insufficient evidence in DevWiki” when support is weak
- **Citations must exist**: do not cite pages, files, or symbols that do not exist
- **Keep retrieval bounded**: prefer top-K evidence instead of unlimited searching

## Error Handling

- **wiki is mostly empty**: tell the user to run `/devwiki-init` or `/devwiki-ingest` first
- **code directory config missing**: answer from `wiki/` and `raw/` only, and say that code was not verified
- **`zatools qmd ...` unavailable**: fall back to local text search and manual inspection
- **no relevant evidence found**: say so honestly and suggest `/devwiki-ingest`, `/devwiki-feature-doc`, or a more specific anchor
- **save requested but evidence is weak**: ask the user for clarification or recommend `/devwiki-refresh` / `/devwiki-feature-doc` before persisting
