---
name: "devwiki-project-router"
description: "Use when the user asks about project features, design docs, code location, troubleshooting, Wiki construction, Wiki queries, Wiki health maintenance, code-to-doc work, or qmd retrieval health in a DevWiki workspace."
argument-hint: "<question, task, or source scope>"
---

# /devwiki-project-router

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`

DevWiki's front door. Decide the user's intent, evidence needs, retrieval boundary, and next Skill before answering or writing.

## Inputs

- `request`: the user's question, task, document range, failure symptom, or knowledge-capture request
- `context` (optional): pasted docs, conversation notes, logs, code anchors, or existing Wiki paths
- `role` (optional): developer, product, QA, support, ops, customer, public website, or external user

## Required First Output

Always start with:

```text
Judgment: this is [intent_type], route to [target Skill], qmd needed/not needed, code search needed/not needed.
```

Then run the target Skill workflow. If the target Skill is not available, state the target and the minimum input needed next.

## Routing Targets

| Intent | Target Skill | Use for |
|---|---|---|
| `ingest` | `devwiki-ingest` | Convert raw material into capability, feature, workflow, troubleshooting, terms, and relations |
| `maintain` | `devwiki-maintain` | Maintain evidence consistency, stale content, missing citations, relation errors, and query pollution in existing Wiki knowledge |
| `query` | `devwiki-query` | Answer capability, feature, engineering-location, code-location, and troubleshooting questions |
| `code_to_doc` | `devwiki-code-to-doc` | Generate or update Wiki pages from real code, APIs, routes, config keys, or log anchors |
| `qmd_sync` | `devwiki-qmd-sync` | Register or repair qmd collections, refresh indexes, and check retrieval readiness |

## Intent Detection

Route to `devwiki-ingest` when the user:

- provides requirements, design docs, API docs, deployment docs, tests, meeting notes, troubleshooting records, or pasted design conclusions
- asks to ingest, import, digest, build a Wiki, or create the first Wiki frame from documents
- wants raw material converted into structured Wiki pages

Route to `devwiki-maintain` when the user:

- asks to maintain Wiki, check Wiki health, audit consistency, or run a maintain pass
- says query answered with old rules, stale mechanisms, or an outdated page
- asks to compare raw/wiki/code consistency or check whether a Feature omitted key design points
- asks to fix conflicts, stale pages, broken links, orphan pages, missing citations, or relation/index/glossary errors
- wants old content marked historical, stale pages downgraded, or query pollution reduced

Route to `devwiki-query` when the user asks:

- what a feature means
- where a design, workflow, config, or troubleshooting note is documented
- which workflow/code path owns something
- how two concepts differ
- where code likely lives, when the answer can start from Wiki evidence

Route to `devwiki-code-to-doc` when the user:

- provides an API URL, file, function, route, config key, or log keyword and asks to document it
- says docs are missing or stale and current implementation should become Wiki knowledge
- wants workflow or troubleshooting pages created from code evidence, with feature updates only when needed

Route to `devwiki-qmd-sync` when the user:

- says qmd search/query is unavailable or stale
- asks to check `config/search.yaml`
- asks about `zatools qmd sync`, `download`, `update`, `status`, or `embed`
- needs existing DevWiki collections registered or repaired

## Priority For Mixed Requests

If one request matches several intents, choose in this order:

1. `devwiki-qmd-sync`
2. `devwiki-ingest`
3. `devwiki-maintain`
4. `devwiki-code-to-doc`
5. `devwiki-query`

If still unclear, ask what the desired output is: an answer, a Wiki write, Wiki maintenance, a code-derived page, or qmd repair.

## User Role

Default role:

```text
developer
```

Role boundaries:

| Role | Allowed knowledge |
|---|---|
| `developer` | `wiki/` + `raw/` + skills + code |
| `internal_non_developer` | `wiki/` + public or internally shareable raw summaries; avoid implementation detail by default |
| `external_user` | public only; do not expose internal paths, functions, private design, troubleshooting detail, or customer data |

Treat users as `external_user` when they explicitly say customer, external user, website, public page, or public-facing answer. Treat users as `internal_non_developer` when they say product, QA, presales, support, or ops.

## Retrieval Rules

Use the tiered recall rules in `references/zatools-qmd.md`:

1. If the request has clear anchors, search and read local files first.
2. If the request has only concepts or keywords, use `zatools qmd search`.
3. Use `zatools qmd query` only when concept-level recall is needed and the environment supports it.
4. If qmd cannot run, fall back to local search and say retrieval was degraded.
5. Stop once the top candidates are strong enough; do not expand without bounds.

Do not enter DevWiki retrieval for generic shell commands, language syntax, pure editing of fully provided text, or requests unrelated to project knowledge.

## Code Search Rules

Do not start with blind global code search. Use code search only when:

- the user asks for files, functions, interfaces, call chains, runtime behavior, config keys, or log origins
- Wiki/raw evidence is insufficient and code is required
- the target Skill will write or correct `code_refs`, `api_entries`, or `test_refs`

Search order:

1. Read related workflow `code_refs`, `api_entries`, and `test_refs`.
2. Search known anchors directly with `rg`.
3. If there is no code anchor, use Wiki/raw/qmd candidates to narrow the search first.
4. Read only the key top-K files, entries, and branches needed for the answer.
5. If code and docs conflict, report both as separate evidence.

## Workflow

1. Decide whether the request belongs to DevWiki. If not, say it is not a DevWiki project-knowledge task and answer normally.
2. Determine the user role and visibility boundary.
3. Classify the intent using the routing table.
4. Decide whether qmd, local Wiki/raw search, code search, and user confirmation are needed.
5. Output the required judgment line.
6. Continue with the routed Skill workflow.

## Edge Cases

- **Wiki is mostly empty**: route document import or first Wiki construction to `devwiki-ingest`.
- **Wiki has stale conclusions polluting query**: route knowledge-health correction to `devwiki-maintain`.
- **The user only pasted complete text and asks for polishing**: do not enter DevWiki unless they ask to persist knowledge.
- **External user asks for internals**: answer with the public boundary and avoid code paths/functions.
- **Several possible targets exist**: ask one to three concrete questions about desired output or anchors.
