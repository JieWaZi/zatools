# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go CLI project. The executable entrypoint lives in `cmd/zatools/main.go`.

- `internal/cli/`: Cobra command wiring for `skill`, `rule`, `devwiki`, and shell completion.
- `internal/app/skillapp/`: application-layer orchestration for `skill add`, `list`, `remove`, `check`, `update`, `init`, and builtin-library installs such as `zatools/devwiki`.
- `internal/app/ruleapp/`: application-layer orchestration for `rule add`, `list`, `remove`, `check`, and `update`.
- `internal/app/devwikiapp/`: application-layer orchestration for `devwiki init`, `devwiki update`, and runtime skill installation.
- `internal/skills/`: core domain logic for source parsing, installation, lock files, and workspace resolution.
- `internal/rules/`: rule discovery and metadata parsing.
- `internal/devwiki/`: DevWiki project generation, code-repo linking, and reset/log tooling.
- `internal/qmd/`: `zatools qmd` config parsing, environment injection, command execution, and collection sync helpers.
- `internal/cli/qmd/`: top-level `zatools qmd` Cobra wiring and passthrough argument parsing.
- `internal/platform/agents/`: agent-specific installation path rules.
- `internal/ui/`: terminal output, localized copy, selectors, and styling.

Keep new code inside `internal/` unless it must be imported by another module.

## Build, Test, and Development Commands

Use standard Go tooling from the repository root:

- `go build ./cmd/zatools`: build the CLI binary.
- `go run ./cmd/zatools --help`: run the tool locally.
- `go run ./cmd/zatools devwiki --help`: inspect the DevWiki command group locally.
- `go test ./...`: run all unit tests.
- `go test -race ./...`: run tests with the race detector.
- `go vet ./...`: catch common correctness issues.
- `gofmt -w .`: format Go files before submitting changes.

If `golangci-lint` is installed and compatible with your Go version, run `golangci-lint run` as an extra check.

## Coding Style & Naming Conventions

Follow idiomatic Go:

- Format with `gofmt`; do not hand-align spacing.
- Keep package names short, lowercase, and singular where possible, for example `skills`, `cli`, `ui`.
- Exported identifiers require Go doc comments.
- New exported struct fields should also include concise field comments; do not leave newly added fields undocumented.
- Prefer small functions and clear layer boundaries: `cmd -> internal/cli -> internal/app -> internal/skills`.
- Use `context.Context` for blocking or cancelable operations.

## Important Notes

- All user-facing Chinese and English copy must be defined centrally in `internal/ui/i18n.go`; do not duplicate localized strings inside feature packages such as `internal/app/*` or `internal/cli/*`.
- When adding a new command or asset type, wire its display text, prompts, flags, statuses, and count text through the shared i18n catalog first, then reference that catalog from the implementation.
- Keep root docs and embedded DevWiki template docs aligned with user-visible `devwiki` CLI changes when behavior or command surface changes.
- For DevWiki qmd workflows, keep the repo-level docs and embedded template docs aligned on command semantics: `sync` handles collection registration, `download` prewarms required qmd models, `devwiki init` triggers that warmup automatically, `update` refreshes the index after writes, and `status` verifies qmd-first readiness.
- Treat the two rules above as required review items for future changes.

## Testing Guidelines

Tests use Go’s built-in `testing` package. Place tests next to the code they cover, using `*_test.go` naming, as in `internal/skills/source_test.go`.

- Prefer table-driven tests for parsing and validation logic.
- Cover error paths as well as success cases.
- Run `go test ./...` and `go test -race ./...` before opening a PR.

## Commit & Pull Request Guidelines

The current history is minimal (`init`), so use short, imperative commit messages such as `refactor cli layering` or `add source traversal tests`.

For pull requests:

- Explain the user-visible impact and the internal design change.
- List verification commands you ran.
- Link any related issue if one exists.
- Include terminal output or screenshots only when UI behavior changes.
