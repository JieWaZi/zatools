---
name: "devwiki-check"
description: "Use when running deterministic health checks on DevWiki documents, capabilities, changes, links, source_hash values, code refs, symbols, and index state, especially for validation, periodic audits, and before or after refresh."
argument-hint: "[check-scope]"
---

# /devwiki-check

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`
> - If the task traces code, attributes ownership, or verifies implementation behavior, also read `references/code-tracing.md`


> Run deterministic health checks on DevWiki and generate a tiered report. Report-only by default.

## Inputs

- `scope` (optional): check scope, such as “entire wiki” or “user-permission pages”
- `wiki/` directory
- optional `--fix` — auto-fix deterministic low-risk issues
- optional `--fix --dry-run` — preview repairs without applying them
- optional `--json` — emit JSON output

## Outputs

- health report
- optional repair preview or repair result
- optional write to `wiki/outputs/check-report-<date>.md`
- check summary appended to `wiki/log.md`

## DevWiki Interaction

### Reads

- `wiki/documents/**/*.md` — verify `source_path`, `source_hash`, and linked fields
- `wiki/capabilities/*.md` — verify `documents`, `changes`, and `code refs`
- `wiki/changes/*.md` — verify change links and classification
- `wiki/index.md` — verify catalog completeness
- `raw/*/*.md` — verify whether source material still exists
- local code directory — verify `code_refs.path` and `symbol`

### Writes

- no page writes by default
- only when `--fix` is explicitly requested and the issue is deterministic and low risk
- APPEND `wiki/log.md`


## Workflow

### Step 1: Run baseline checks

At minimum, check:

1. missing required fields
2. broken `source_path`
3. mismatched `source_hash`
4. missing `code_refs.path`
5. missing `symbol` in the referenced file
6. stale or missing index entries
7. missing reverse links
8. orphan pages
9. stale `qmd` index state

### Step 2: Classify findings

Group issues into:

- 🔴 fix immediately: deterministic broken links, paths, hashes, symbols
- 🟡 recommended fixes: classification anomalies, missing reverse links, stale index
- 🔵 optional improvements: orphan pages, low-value redundant clues

### Step 3: Handle fix mode

1. Default mode is report-only
2. If the user passes `--fix`:
   - repair only deterministic low-risk issues
   - examples: refresh `source_hash`, remove broken auxiliary code clues, patch simple index entries
3. If the user passes `--fix --dry-run`:
   - preview repair candidates without writing
4. Anything that requires re-judging capability or change ownership must not be auto-fixed; route it to `/devwiki-refresh`

### Step 4: Report results

The report must include:

- check scope
- issue counts
- details by severity
- if fix mode ran: what was repaired, what was previewed, and what still needs manual handling
- next-step suggestions:
  - use `--fix` for deterministic repairs
  - use `/devwiki-refresh` for classification drift
  - use `/devwiki-feature-doc` when structured documentation is missing

### Step 5: Log the run

Append to `wiki/log.md`:

- `check | report-only | <summary>`
- or `check | fix-applied | <summary>`

## Constraints

- **Report-only by default**: without `--fix`, do not modify wiki pages
- **`--fix` repairs only deterministic low-risk issues**: no automatic classification repairs
- **raw/ is read-only**: do not modify source material
- **No fabricated symbols**: if a symbol is missing, report it as missing
- **Results should be stable**: repeated runs on the same state should produce similar results

## Error Handling

- **wiki/ missing**: tell the user to run `/devwiki-init` first
- **code directory missing**: skip code checks and say coverage is reduced
- **`zatools qmd ...` unavailable**: report reduced index validation but continue other checks
- **fix mode hits medium/high-risk items**: stop auto-repair for those items and recommend `/devwiki-refresh`
