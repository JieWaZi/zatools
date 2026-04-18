---
name: "devwiki-qmd-sync"
description: "Use when a DevWiki workspace already exists but `zatools qmd` collections have not been registered, the registration looks suspicious, the index is stale, or the `zatools qmd` retrieval mode must be restored to a healthy state."
argument-hint: "[--root <devwiki-root>]"
---

# /devwiki-qmd-sync

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If the task writes, reclassifies, or performs destructive actions, also read `references/mutation-safety.md`


> Register or repair `zatools qmd` collections for an existing DevWiki workspace, refresh the index, and report whether the `zatools qmd` retrieval mode is actually ready. Prefer dry-run and status checks before applying changes.

## Inputs

- `--root <devwiki-root>`: optional; default to the current directory
- `config/search.yaml`
- local access to `zatools qmd ...`

## Outputs

- `zatools qmd` collection dry-run commands
- optional applied collection-registration results
- the latest `zatools qmd` index refresh result
- the latest `zatools qmd status` summary
- a recommendation on whether `embed` is still needed for higher-quality semantic retrieval

## DevWiki Interaction

### Reads

- `config/search.yaml`
- `config/project.yaml`
- local `zatools qmd` collection and index state

### Writes

- no wiki-page writes
- may update `zatools qmd` collection, index, and embedding state


## Workflow

### Step 1: Validate prerequisites

1. Confirm that the current directory or `--root` points at a real DevWiki root
2. Confirm that `config/search.yaml` exists
3. Confirm that `zatools qmd status` can run in the current environment
4. If `zatools qmd ...` cannot run, enter fallback mode immediately: print the required commands and explicitly say the `zatools qmd` retrieval mode cannot be enabled yet

### Step 2: Inspect the dry-run commands first

Run:

```bash
zatools qmd sync --root <devwiki-root>
```

Review the generated collection-add commands before deciding whether to apply them.

### Step 3: Register or repair collections when needed

If collections are missing, paths are wrong, or the user explicitly wants a repair run, execute:

```bash
zatools qmd sync --root <devwiki-root> --apply
```

Do not apply blindly when the task only requires inspection.

### Step 4: Refresh the index and inspect status

Once collection registration looks correct, execute:

```bash
zatools qmd update
zatools qmd status
```

Confirm at least:

- the DevWiki collections exist
- the `raw / wiki / code` collections do not report obviously wrong zero-file states
- `Pending`, `Updated`, and related status lines match the current workspace state

### Step 5: Run embed only when it is justified

If the current task depends on higher-quality semantic retrieval through `zatools qmd query`, and `status` still shows pending vectors, ask whether to continue with:

```bash
zatools qmd embed
```

Do not force `embed` after every sync by default.

### Step 6: Report readiness

The conclusion should include:

- whether `zatools qmd` collections are correctly registered
- whether the index has been refreshed
- whether embeddings are still pending
- whether downstream skills can now run in the `zatools qmd` retrieval mode
- if not healthy yet, whether the next action is fixing config, re-running sync, or staying in fallback mode

## Constraints

- **Dry-run before apply**
- **Do not treat `zatools qmd` retrieval hits as facts**
- **Keep collection registration separate from index refresh**
- **`embed` is opt-in by need, not mandatory every time**
- **If `zatools qmd ...` is unavailable, explicitly report fallback mode**

## Error Handling

- **missing `config/search.yaml`**: tell the user to run `/devwiki-setup` or `zatools devwiki init` first
- **`zatools qmd ...` cannot run**: stop before apply, print the required commands, and state that the `zatools qmd` retrieval mode is not active
- **sync apply fails**: report the failure, keep the dry-run output, and do not pretend registration succeeded
- **status shows missing collections or suspicious zero-file counts**: recommend checking collection paths or re-running sync
