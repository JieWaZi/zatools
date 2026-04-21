# Code Tracing Discipline

> Shared reference for `/devwiki-ask`, `/devwiki-feature-doc`, and any workflow that must move from an entry anchor into concrete implementation understanding.

> Use this only after the workflow has already decided that code inspection is necessary. If documents already answer the question, `/devwiki-ask` should answer first and offer code verification as an optional follow-up instead of tracing by default.

---

## Core Rule

Do not stop at the first plausible file.

A correct DevWiki trace starts from a clear **entry anchor**, then walks the relevant call chain until the implementation boundary is understood well enough to explain the behavior.

---

## Acceptable Entry Anchors

A good trace should start from at least one of:

- API URL
- key file
- key function
- page route
- explicit feature name
- known capability name

If none exists, ask the user for one before broad code search explodes.

---

## Standard Tracing Flow

### Step 1: Normalize the Entry Anchor

Translate the user input into searchable terms.

Examples:

- API URL → controller / router / handler candidates
- key function → direct symbol lookup
- page route → frontend route + backend endpoint candidates
- feature name → capability candidates + top-K code hits

### Step 2: Retrieve Top-K Candidates

Use `qmd` and local code search to produce top-K candidates, then reduce to the most likely entry files.

Do not confuse retrieval with understanding.

### Step 3: Verify The Entry File

Confirm:

- the file really participates in the feature
- the target symbol exists
- the file is an entry point, not a random helper

### Step 4: Walk The Call Chain

Trace the implementation deeper:

- caller
- callee
- key branching logic
- persistence or external API boundaries
- important side effects

If the behavior crosses multiple layers, follow the chain until the key design decisions are visible.

### Step 5: Stop At A Clear Boundary

A trace is sufficiently grounded when you can answer:

- where the request enters
- which modules own the main behavior
- which functions or classes do the critical work
- where side effects happen
- what remains unresolved

---

## When To Ask The User

Ask the user instead of guessing when:

- multiple entry files look equally plausible
- the claimed interface cannot be found
- symbol lookup fails repeatedly
- the call chain breaks due to dynamic dispatch or unclear routing
- code search returns too many scattered matches

Ask 1 to 3 focused questions only.

---

## Output Expectations

A good trace summary should include:

- entry anchor used
- verified top-K candidates
- main call chain
- important implementation nodes
- unresolved gaps

When the output is a document, prefer `path + symbol` for large files.

---

## What NOT To Do

- Do not stop after reading one top-level function
- Do not treat filename similarity as proof
- Do not skip internal calls just because the first layer looks obvious
- Do not invent a call chain when the trace is incomplete
- Do not avoid asking the user when the trace remains ambiguous
