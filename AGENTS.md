# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go CLI project. The executable entrypoint lives in `cmd/zatools/main.go`.

- `internal/cli/`: Cobra command wiring and shell completion.
- `internal/app/skillapp/`: application-layer orchestration for `add`, `list`, `remove`, `check`, `update`, and `init`.
- `internal/skills/`: core domain logic for source parsing, installation, lock files, and workspace resolution.
- `internal/platform/agents/`: agent-specific installation path rules.
- `internal/ui/`: terminal output, localized copy, selectors, and styling.

Keep new code inside `internal/` unless it must be imported by another module.

## Build, Test, and Development Commands

Use standard Go tooling from the repository root:

- `go build ./cmd/zatools`: build the CLI binary.
- `go run ./cmd/zatools --help`: run the tool locally.
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
- Prefer small functions and clear layer boundaries: `cmd -> internal/cli -> internal/app -> internal/skills`.
- Use `context.Context` for blocking or cancelable operations.

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
