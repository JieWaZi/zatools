# zatools qmd Usage

> Shared reference for DevWiki skills that retrieve, sync, or refresh through `zatools qmd ...`.

Always use `zatools qmd ...` for retrieval and maintenance commands. Do not call other commands directly.

In sandboxed agents such as Codex or Claude Code, confirm the agent can run `zatools qmd ...` successfully and that the project-root `.cache` directory is writable; otherwise `zatools qmd ...` status checks may fail and should be reported as reduced validation coverage.

Rule: if the task is running inside a DevWiki workspace, read `embed_model`, `rerank_model`, and `generate_model` from `config/search.yaml` first, then pass them explicitly on every `zatools qmd ...` command with `--embed-model`, `--rerank-model`, and `--generate-model`. If any value is missing, fall back to the CLI built-in defaults.

For retrieval-type tasks, do not run `zatools qmd status` as a prerequisite probe. Execute the target `zatools qmd ...` retrieval command directly. If that command fails, treat the `zatools qmd` path as unavailable for the current run, report degraded retrieval, and fall back to local index, text, or file search.

For retrieval-type tasks, prefer dual-path recall: one path through `zatools qmd ...`, one path through local index or direct file/code search. If the runtime supports delegation or parallel agents, run those two paths in parallel and merge the results in the main answer. If delegation is unavailable, run them sequentially with the same fallback behavior.
