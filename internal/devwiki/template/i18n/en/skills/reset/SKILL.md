---
name: "devwiki-reset"
description: "Use when DevWiki generated content must be cleared by scope, failed initialization left residue behind, or a clean workspace is needed before running init or ingest again."
argument-hint: "--scope wiki|raw|log|checkpoints|all [--project-root <devwiki-root>]"
---

# /devwiki-reset

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/mutation-safety.md`

> Reset DevWiki by scope. `/devwiki-reset` is destructive and must always show a dry-run plan before execution.

## Inputs

- `--scope`: required, one of `wiki`, `raw`, `log`, `checkpoints`, `all`
- `--project-root`: optional; use `.` when the current directory is already the DevWiki workspace
- the current `raw/` and `wiki/` trees

## Outputs

- a dry-run deletion and reset plan
- an execution summary after user confirmation
- optional reset entry in `wiki/log.md`

## DevWiki Interaction

### Reads

- `raw/`
- `wiki/documents/`
- `wiki/capabilities/`
- `wiki/changes/`
- `wiki/outputs/`
- `wiki/graph/`
- `wiki/.checkpoints/`

### Writes

- DELETE files selected under `raw/`
- DELETE generated files under `wiki/documents/`, `wiki/capabilities/`, `wiki/changes/`, `wiki/outputs/`, and `wiki/graph/`
- RESET `wiki/index.md`
- optionally RESET `wiki/log.md`

## Workflow

### Step 1: Normalize the scope and run dry-run

Normalize the requested scope, then run:

```bash
zatools devwiki tool reset --scope <scope> --project-root <devwiki-root>
```

This prints a plan only. It must not delete anything yet. Show `delete` and `reset` separately.

### Step 2: Explain the risk

Explain what each scope means:

- `wiki`: clear generated knowledge pages and outputs while preserving the scaffold
- `raw`: clear raw source material and is the highest-risk option
- `log`: reset `wiki/log.md`
- `checkpoints`: clear intermediate state
- `all`: everything above

If the scope includes `raw` or `all`, explicitly warn that deleting `raw/` is usually irreversible.

### Step 3: Wait for explicit confirmation

Never execute before confirmation. Use wording like:

```text
About to delete N files and reset M files for scope=<scope>. Confirm before continuing.
```

Only proceed after the user explicitly confirms.

### Step 4: Execute the reset

After confirmation, run:

```bash
zatools devwiki tool reset --scope <scope> --project-root <devwiki-root> --yes
```

Read the result and verify:

- which files were actually deleted
- which files were actually reset
- whether the counts match the plan

### Step 5: Append a reset log

If the scope did not include `log`, append a low-risk log entry:

```bash
zatools devwiki tool log --wiki-root <devwiki-root>/wiki --message "reset | scope=<scope>"
```

### Step 6: Suggest next steps

Tailor the next step to the scope:

- if `raw/` was cleared: ask the user to restore source documents first
- if only `wiki` was cleared: suggest `/devwiki-init` or `/devwiki-ingest`
- if only `checkpoints` was cleared: suggest resuming the previous ingest or refresh flow

Common follow-ups:

- `/devwiki-init`
- `/devwiki-ingest`
- `/devwiki-setup`

## Constraints

- **Dry-run, then confirm, then execute**: never skip the plan stage
- **Do not call `--yes` without explicit confirmation**
- **Preserve scaffold markers**: do not delete `.gitkeep`
- **Do not touch installation state**: never modify project-root `.agents/` or `.zatools-lock.json`
- **`raw/` is high risk**: any scope including `raw` requires an extra warning
- **Results must be explainable**: report deletes and resets separately instead of saying “cleared”

## Error Handling

- **Missing or invalid scope**: print valid values and stop instead of guessing
- **`zatools devwiki tool reset` fails**: report the failure and do not replace it with ad-hoc bulk deletes
- **Missing `wiki/` or `raw/`**: produce an empty or partial plan and explain why
- **Log append failure**: report a warning without hiding the main reset result
- **User cancels confirmation**: clearly state that only a dry-run happened and nothing was modified
