---
name: "devwiki-setup"
description: "Use when DevWiki must be initialized through zatools, a runtime must be selected, one or more code directories must be registered, and DevWiki skills must be installed."
argument-hint: "[<project-name>] [--agent codex|cursor|claude --lang zh|en --code-dir <path>] [--global]"
---

# /devwiki-setup

> Read the shared constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - `references/mutation-safety.md`

> Initialize DevWiki through zatools and always use `zatools devwiki init`.

## Inputs

- `project-name`: optional; prompt for it in interactive mode, but require it in non-interactive mode
- `--agent codex|cursor|claude`
- `--lang zh|en`
- `--code-dir <path>`: repeatable, or passed once as a comma-separated list
- `--global`: optional; default is project-scoped installation
- the current working directory and its detected project root

## Outputs

- a new `devwiki-<project-name>/` directory
- `devwiki-<project-name>/README.md`
- `devwiki-<project-name>/raw/`
- `devwiki-<project-name>/wiki/`
- `devwiki-<project-name>/config/project.yaml`
- `devwiki-<project-name>/config/search.yaml`
- the selected DevWiki skills
- for project-scoped installs: project-root `.agents/` and `.zatools-lock.json`
- for global installs: home-scoped skill installation and lock file
- optional `zatools qmd` registration results or generated manual commands
- optional manual reminder to run `zatools qmd download --root .` after init
- optional: hand off to `devwiki-qmd-sync` when the workspace already exists

## DevWiki Interaction

### Reads

- current working directory, to detect the project root
- user-provided code directories
- built-in DevWiki skills and shared references
- `devwiki-<project-name>/config/search.yaml` for `zatools qmd sync`

### Writes

- CREATE `devwiki-<project-name>/`
- CREATE `devwiki-<project-name>/raw/`
- CREATE `devwiki-<project-name>/wiki/`
- CREATE `devwiki-<project-name>/config/project.yaml`
- CREATE `devwiki-<project-name>/config/search.yaml`
- CREATE / UPDATE the selected-scope DevWiki skills
- CREATE / UPDATE `.zatools-lock.json`


## Workflow

### Step 1: Collect initialization parameters

If the user did not provide a full command line, fill in the missing fields interactively:

- project name
- runtime: `codex`, `cursor`, or `claude`
- language: `zh` or `en`
- one or more code directories
- install scope: project or global

If the arguments are already present, do not ask again.

### Step 2: Create the DevWiki project and install skills

Standard initialization command:

```bash
zatools devwiki init <project-name> --agent <agent> --lang <lang> --code-dir <dir1> --code-dir <dir2>
```

For non-interactive usage, prefer:

```bash
zatools devwiki init <project-name> --agent <agent> --lang <lang> --code-dir <dir1> --yes
```

If the user explicitly wants global skill installation, add:

```bash
--global
```

For project-scoped installs, remember:

- `.agents/` and `.zatools-lock.json` are written to the **current detected project root**
- not inside `devwiki-<project-name>/`

### Step 3: Sync the `zatools qmd` retrieval layer

If the user wants `zatools qmd` collections, the dry-run command is:

```bash
zatools qmd sync --root <devwiki-root>
```

Only apply it when appropriate:

```bash
zatools qmd sync --root <devwiki-root> --apply
```

After collection registration, continue with:

```bash
zatools qmd update
zatools qmd status
```

If the user wants to download qmd models manually, run this inside the DevWiki workspace:

```bash
zatools qmd download --root .
```

If `status` still shows significant pending embeddings and the next task explicitly depends on higher-quality semantic retrieval, ask whether to continue with:

```bash
zatools qmd embed
```

If `zatools qmd ...` is unavailable, do not pretend setup succeeded. Print the generated commands and explicitly state that the workspace is in fallback mode.

### Step 4: Print the setup report

The report should include at least:

- the created `devwiki-<project-name>` path
- selected runtime and language
- registered code directories
- whether skills were installed project-wide or globally
- where `.agents/` and `.zatools-lock.json` landed for project-scoped installs
- whether `zatools qmd` was applied or only printed
- whether `zatools qmd update` / `zatools qmd status` were executed
- whether the user should continue with `devwiki-qmd-sync`

## Constraints

- **Do not rely on the old script-based bootstrap chain**
- **`zatools devwiki init` already installs skills**
- **Do not generate internal template artifacts**: no `i18n/`, `tools/`, `setup.*`, `requirements.txt`, or `config/*.example`
- **Project-scoped installation state belongs at the current project root**: not inside `devwiki-<project-name>/`
- **Code directories must be real and accessible**
- **Do not fake `zatools qmd` success**: either apply it or clearly report fallback mode

## Error Handling

- **Missing project name with no interaction possible**: stop and report it instead of guessing
- **Code directory missing or not a directory**: stop and ask the user to fix `--code-dir`
- **Target directory already exists**: stop instead of overwriting `devwiki-<project-name>`
- **Skill installation fails**: report the failure and do not claim setup completed
- **`zatools qmd sync --apply` fails**: fall back to printing commands and explicitly state fallback mode
- **User cancels the prompts**: clearly say setup did not complete
