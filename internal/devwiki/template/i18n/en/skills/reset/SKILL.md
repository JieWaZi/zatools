---
name: "devwiki-reset"
description: "Use when resetting generated DevWiki content under controlled scopes such as wiki, raw, log, or checkpoints."
argument-hint: "[scope list]"
---

# /devwiki-reset

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`


> Reset generated DevWiki content in bounded scopes. Prefer preview-first; destructive actions require explicit confirmation.

## Inputs

- `scope` — comma-separated scope list: `wiki`, `raw`, `log`, `checkpoints`, or `all`
- optional `--dry-run` — preview only
- optional `--yes` — confirm the reset

## Outputs

- reset plan
- optional applied reset result
- appended reset log in `wiki/log.md`

## DevWiki Interaction

### Reads

- `wiki/capabilities/`
- `wiki/features/`
- `wiki/outputs/`
- `wiki/graph/`
- `wiki/index.md`
- `wiki/log.md`
- `raw/*/`
- `wiki/.checkpoints/`

### Writes

- DELETE generated files under `wiki/capabilities/`, `wiki/features/`, `wiki/outputs/`, and `wiki/graph/`
- DELETE files under the selected `raw/` subdirectories
- RESET `wiki/index.md`
- RESET `wiki/log.md` only when the `log` scope is selected


## Workflow

### Step 1: Build the reset plan

1. Expand `all` into `wiki,raw,log,checkpoints`
2. Collect candidate deletions
3. Never delete `.gitkeep`
4. Treat missing files as no-op, not failure

### Step 2: Show the plan

The preview must list:

- selected scopes
- delete targets
- reset targets
- whether the run is dry-run or executable

### Step 3: Confirm destructive execution

1. If `--dry-run`, do not write anything
2. If the run would delete files and `--yes` was not provided, stop at the plan
3. Only apply the reset after explicit confirmation

### Step 4: Apply the reset

After confirmation:

1. delete the planned files
2. rewrite `wiki/index.md` with the baseline template when the `wiki` scope is selected
3. rewrite `wiki/log.md` with the baseline template when the `log` scope is selected
4. append a dated reset summary to the surviving log when possible

## Constraints

- **Preview-first**: do not apply deletions implicitly
- **Do not delete `.gitkeep`**: keep the directory skeleton stable
- **Scope must stay bounded**: do not expand beyond the selected scopes
- **raw resets are destructive**: confirm before deleting source material

## Error Handling

- **unknown scope**: report valid scopes and stop
- **scope omitted**: ask for a concrete scope
- **path already missing**: treat as no-op
