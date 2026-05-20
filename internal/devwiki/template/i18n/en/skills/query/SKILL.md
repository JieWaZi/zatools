---
name: "devwiki-query"
description: "Use when the user asks about project features, design details, capability boundaries, code locations, flows, configuration, troubleshooting, public-facing explanations, or existing knowledge. This skill answers from DevWiki, glossary, zatools qmd retrieval, and necessary rg code search."
argument-hint: "<question>"
---

# /devwiki-query

> Read shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - Read `references/code-tracing.md` for code tracing, attribution, or implementation checks.
> - Read `references/mutation-safety.md` before saving answers or crystallizing conclusions.

Answer from Project Brain. Do not invent project facts; check DevWiki first, verify code only when needed, and keep important conclusions traceable.

## Outputs

- sourced answer
- matching capability / feature / workflow / troubleshooting pages
- code-location clues when needed
- knowledge gaps, conflicts, and questions to align
- change impact and test suggestions for development questions
- public-facing answer when the role requires it
- crystallization suggestion: worth saving / not needed

## Reads

- `config/project.yaml`
- `config/search.yaml`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/capabilities/*.md`
- `wiki/features/*.md`
- `wiki/workflows/*.md`
- `wiki/troubleshooting/*.md`
- local code directories only when implementation truth, code location, config location, log location, change impact, or troubleshooting requires it

## Writes

Default: write no files. Only when the user explicitly asks to save or crystallize the answer:

- CREATE `wiki/outputs/<query-slug>.md`
- APPEND `wiki/log.md`

## Directory Selection

| Intent | User is asking | Primary directory | Support |
|---|---|---|---|
| capability explanation | why it exists, what it can do, capability boundary | `wiki/capabilities/` | `wiki/features/` |
| feature explanation | behavior, parameters, values, interactions, design flow | `wiki/features/` | `wiki/capabilities/`, `wiki/workflows/` |
| code location / change impact | where code lives, call chain, how to change it, what is impacted | `wiki/workflows/` | `wiki/features/`, then `rg` |
| troubleshooting | error, not working, diagnosis, fix | `wiki/troubleshooting/` | `wiki/workflows/`, `wiki/features/` |

Default order:

- capability question: `capabilities -> features`
- feature question: `features -> capabilities`
- code question: `workflows -> features -> rg`
- troubleshooting question: `troubleshooting -> workflows -> features`

## Authority Rules

- Capability definitions come from capability pages.
- Feature rules, parameters, interactions, and design flow come from feature pages.
- Call chains, code references, change impact, and test entry points come from workflow pages.
- Logs, error codes, diagnosis, and fixes come from troubleshooting pages.

Other pages may be used only as summaries or navigation.

## Workflow

### Step 1: Narrow Intent

Read `config/project.yaml`, `wiki/index.md`, and `wiki/glossary.md`. Classify the question as:

```text
explain_feature
locate_code
troubleshoot
compare
public_answer
design_detail
change_impact
```

If `wiki/index.md` or `wiki/glossary.md` is missing, answer:

```text
The current Project Brain does not have enough information to support that conclusion.
```

and suggest `devwiki-ingest`.

### Step 2: Retrieve Evidence

Use query terms from the user question, `wiki/glossary.md`, `wiki/index.md`, relationships in candidate page frontmatter/body links, and optional `config/aliases.yml`.

Search the primary directory first. If local matches are insufficient, use `zatools qmd search` according to `references/zatools-qmd.md`. Use `zatools qmd query` only for concept/design/intent retrieval when cheaper steps are insufficient.

If the documents already answer the question, do not expand into code just to feel safer.

### Step 3: Verify Code When Required

Verify code when the user asks where, which file, which function, which interface, whether current implementation matches, change impact, config location, log location, or troubleshooting current behavior.

Prefer workflow `code_refs`, `api_entries`, and `test_refs`; then use targeted `rg`; then expand to configured code roots.

### Step 4: Answer

Use:

```markdown
## Conclusion

## Evidence

## Code Location Clues

## Conflicts / Questions To Align

## Change Impact / Tests

## Crystallization Suggestion
```

Only include code-location clues for code, change, or troubleshooting questions.

### Step 7: Save Only When Asked

Only write `wiki/outputs/<query-slug>.md` and append `wiki/log.md` when the user explicitly asks to save the answer. Show the target path and summary first.

## Error Handling

- **Not enough Wiki evidence**: state the gap and suggest `devwiki-ingest` or `devwiki-code-to-doc`.
- **Low retrieval confidence**: stop expanding and ask 1 to 3 concrete questions.
- **Document/code conflict**: show document description and code reality separately.
- **Current implementation required but code not checked**: do not give a definite conclusion.
