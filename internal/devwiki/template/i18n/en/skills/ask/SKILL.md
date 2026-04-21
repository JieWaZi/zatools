---
name: "devwiki-ask"
description: "Use when Codex needs to answer questions over existing DevWiki capabilities, feature pages, raw documents, and related code clues, AND when a change must be scoped before implementation to decide whether it is new or modify."
argument-hint: "<question or change description>"
---

# /devwiki-ask

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Unified DevWiki entry point for both general Q&A and pre-implementation change scoping. Read-only by default; attach the change-classification block based on the user's intent.

## Inputs

- `question`: free-form input — may be a question OR a change description
- `--format` (optional): output format, default `markdown`, options: `table` / `bullets` / `timeline`
- `--save-output` (optional): use only when the user explicitly wants the answer persisted under `wiki/outputs/`

## Outputs

- **Always**: one synthesized answer with citations
- **When the intent is development / change**: also emit `classification: new / modify / unclear` + top-K candidate features/code + next-step suggestion
- **Optional**: if explicitly requested, write `wiki/outputs/<query-slug>.md`
- **Included**: related capabilities, related features, code locations, and knowledge gaps

## DevWiki Interaction

### Reads

- `config/project.yaml` — primary code directory, language, and repo configuration
- `wiki/index.md` — locate candidate pages
- `wiki/capabilities/*.md` — capability aggregation pages
- `wiki/features/*.md` — feature pages with source trace, entry points, code refs, and test refs
- `wiki/outputs/*.md` — previously saved answers
- `raw/*/*.md` — original source documents when wiki summaries are insufficient
- local code directory — inspect only when the question is about implementation reality, ownership, endpoint flow, or when wiki/raw evidence is still insufficient

### Writes

- No writes by default
- Only when the user explicitly requests persistence:
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: Detect intent and bound the scope

1. Read `config/project.yaml` to determine the primary code directory
2. Classify intent from the user input:
   - **Query / Q&A**
     - signal phrases: "how was...", "what capability owns...", "which feature supports...", "which file owns...", "what does ... do"
     - run Step 2 → if wiki/raw already settle it, Step 4; otherwise Step 3 → Step 4 → done
   - **Development / change**
     - signal phrases: "we need to change...", "add a new...", "how should this be built", "refactor...", "modify..."
     - run Step 2 → Step 3 only when implementation verification is needed → Step 5 → Step 4 → done
3. When the boundary is ambiguous, default to the query path and note at the end that the user can ask for change classification explicitly

### Step 2: Retrieve candidate evidence

Follow the tiered recall rules in `references/zatools-qmd.md`:

1. **Start with local `grep` / file search** for known anchors (symbols, files, API URLs, capability slugs, feature slugs, ticket ids)
2. If the results are thin, escalate to `zatools qmd search` for keyword recall across `raw / wiki / code`
3. Only when tiers 1 and 2 are still insufficient AND the question is a concept/design/intent question, consider `zatools qmd query`; apply the hard fallback when no GPU/accelerator is available
4. Keep the candidate set bounded: top-K (K ≤ 12), prioritize highly relevant capability pages and feature pages, fall back to `raw/` only when wiki summaries are not enough
5. Decide whether the wiki/raw layer already closes the loop:
   - If the documents already answer the question, do not expand into code just to feel safer
   - If the pages already answer the question, do not expand into code just to feel safer
   - Only enter Step 3 when the pages do not settle the question, the user explicitly asks about implementation reality, or the task is a development / change request

### Step 3: Verify code only when needed

Only enter Step 3 when the wiki/raw layer does not settle the question, the user explicitly asks about implementation reality, or the task is a development / change request.

Only enter Step 3 when the documents do not settle the question, the user explicitly asks about implementation reality, or the task is a development / change request.

These cases require code verification and must not be answered from wiki/raw alone:

- "Is the current implementation doing this?"
- "Which file or function owns this?"
- "How does this endpoint work now?"
- "Which code paths are likely impacted by this change?"
- A development / change request where the related feature pages and existing `code_refs` are still not enough to support the classification or next-step advice

When verifying code:

1. Start from `code_refs`, `api_entries`, and `test_refs` already attached to relevant feature pages
2. If those are insufficient, run targeted code searches in the local repo (tier 1 / tier 2 only)
3. Confirm at least one concrete layer of evidence: entry file, key function, endpoint registration, or a critical call edge
4. If code and docs disagree, state both explicitly
5. If you only need extra change-scoping clues, prefer stopping at feature-level entry anchors instead of deep-reading every file

### Step 4: Compose the answer

The answer must include:

1. **Direct answer first** — do not lead with the retrieval process
2. **Evidence citations** for every key conclusion
   - wiki citations: `wiki/capabilities/...`, `wiki/features/...`
   - raw citations: `raw/...`
   - code citations: file paths and symbols when relevant
3. **Layer separation**
   - confirmed facts
   - reasonable inference
   - unresolved gaps
4. **Related items**
   - related capabilities
   - related features
   - related code refs (only when code was actually verified this round, or when existing `code_refs` themselves are part of the evidence)
5. **Next-step suggestions**
   - use `/devwiki-feature-doc` when code exists but the feature page is missing or stale
   - use `/devwiki-refresh` when wiki knowledge has drifted
   - use `/devwiki-ingest` when raw source material is missing from the knowledge base
   - use `/devwiki-check` when deterministic validation is needed
6. **If the answer is primarily wiki/raw-grounded**: end with "If useful, I can run a second pass against the code and give you an implementation-verified summary"
7. **If Step 5 was executed**: place the change classification block right after the direct answer

### Step 5: Change classification (development / change intent only)

Consolidate Step 2 / Step 3 evidence into three buckets, then classify:

1. **Three evidence buckets**
   - existing capabilities
   - existing features
   - raw sources
2. **Top-K candidate feature/code set**
   - start from related capability pages, related feature pages, existing `code_refs`, `api_entries`, and `test_refs`
   - open file contents only when implementation reality or entry ownership must be verified
   - explain why the feature or file is relevant
   - when code was verified, state whether the key function / class / symbol exists
3. **Emit classification**
   - `modify`: existing capability or feature strongly matches
   - `new`: no meaningful match to current capability or feature, but the target is clear
   - `unclear`: evidence conflicts, is too scattered, or the entry anchor is missing
4. **Next-step suggestion**
   - `new / modify` with enough confidence: proceed to design or `/devwiki-feature-doc`
   - `unclear`: either ask follow-up questions (Step 6) or recommend `/devwiki-refresh` / `/devwiki-ingest` to fill the gap

### Step 6: Ask the user when confidence stays low

After a few independent searches, ask the user instead of forcing an answer when:

- multiple candidate capabilities or features remain plausible
- the referenced endpoint, file, function, or route cannot be found locally
- dynamic registration, config-driven wiring, gateways, or external services block reliable confirmation
- the question is too broad and spans multiple sub-capabilities
- a development / change request has no usable entry anchor

Question rules:

- ask only 1 to 3 narrow questions
- do not ask vague questions like "can you provide more context?"
- prefer anchor questions: API URL, key file, key function, page route, known feature name, known capability name

### Step 7: Persist the result only on explicit request

Only when the user explicitly asks to save the answer:

1. write `wiki/outputs/<query-slug>.md`
2. include the original question, cited sources, conclusions, and unresolved items; include the classification block if Step 5 was executed
3. append `ask | <question-summary> | saved-output` to `wiki/log.md`
4. if the answer reveals stale or incorrect wiki facts, recommend or run `/devwiki-refresh` instead of silently rewriting pages

## Constraints

- **No fabrication**: all key conclusions must come from actual DevWiki content or local code
- **raw/ is read-only**: do not modify files under `raw/`
- **No wiki writes by default**: persist only with explicit user intent
- **Stop when wiki/raw are enough**: if `wiki/` and `raw/` already support the answer and the question is not about implementation reality, do not expand into code by default
- **Code questions require code checks**: implementation-behavior answers must verify code
- **Separate facts from inference**: the classification is inference; page / file / symbol existence is fact
- **Retrieval must stay bounded**: stop and ask after a few unsuccessful rounds
- **Citations must exist**: do not cite pages, files, or symbols that do not exist
- **Do not create change pages**: `new / modify / unclear` is an answer-time judgment only

## Error Handling

- **wiki is mostly empty**: tell the user to run `/devwiki-init` or `/devwiki-ingest` first
- **code directory config missing**: answer from `wiki/` and `raw/` only, and say that code was not verified
- **`zatools qmd ...` unavailable**: apply the fallback rules in `references/zatools-qmd.md`; use local search only
- **`zatools qmd query` not supported by environment**: stop at `zatools qmd search` + local search and state the degradation; do not silently block
- **no relevant evidence found**: say so honestly and suggest `/devwiki-ingest`, `/devwiki-feature-doc`, or a more specific anchor
- **save requested but evidence is weak**: ask the user for clarification or recommend `/devwiki-refresh` / `/devwiki-feature-doc` before persisting
