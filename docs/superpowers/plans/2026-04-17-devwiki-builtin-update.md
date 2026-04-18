# DevWiki Builtin Source And Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `zatools/devwiki` as a builtin skill library source and add `zatools devwiki update` to update changed DevWiki skills in the active scope.

**Architecture:** Treat `zatools/devwiki` as a first-class source in `internal/skills`, so `skill add`, `check`, and `update` can reuse the existing discovery and install flow. Add a DevWiki-specific update command in `internal/app/devwikiapp` that resolves scope from the current project lock, filters DevWiki builtin entries, presents changed skills with default-all selection, and then delegates the actual reinstall work to shared `skillapp` update helpers.

**Tech Stack:** Go, Cobra, Bubble Tea, existing `skills` source/install helpers

---

### Task 1: Add builtin `zatools/devwiki` source support

**Files:**
- Modify: `internal/skills/source.go`
- Modify: `internal/app/common/helpers.go`
- Test: `internal/skills/source_test.go`

- [ ] Add source metadata for builtin libraries, plus a helper for constructing stable builtin source strings with a language variant.
- [ ] Parse `zatools/devwiki` before generic GitHub shorthand parsing, defaulting the variant to the current UI language.
- [ ] Resolve builtin DevWiki sources by extracting embedded builtin skills instead of cloning or reading a local directory.
- [ ] Add tests covering builtin parse and resolve behavior.

### Task 2: Persist DevWiki installs as builtin sources

**Files:**
- Modify: `internal/app/devwikiapp/actions.go`
- Test: `internal/app/devwikiapp/actions_test.go`

- [ ] Change DevWiki runtime skill installation to write lock entries with builtin source metadata instead of temporary extraction paths.
- [ ] Preserve the selected DevWiki language in the stored source string so later updates can resolve the same language variant.
- [ ] Add tests proving `devwiki init` or direct DevWiki skill install stores `zatools/devwiki` in the lock file.

### Task 3: Expose shared skill update helpers

**Files:**
- Modify: `internal/app/skillapp/service.go`
- Modify: `internal/app/skillapp/actions.go`
- Test: `internal/app/skillapp/actions_test.go`

- [ ] Export a shared “check installed skills” helper for reuse outside `skillapp`.
- [ ] Export a shared “apply updates for selected check results” helper that reuses the existing reinstall logic.
- [ ] Refactor `skill update` to call the shared helper without changing current behavior.

### Task 4: Implement `devwiki update`

**Files:**
- Modify: `internal/app/devwikiapp/service.go`
- Modify: `internal/app/devwikiapp/actions.go`
- Modify: `internal/cli/devwiki/command.go`
- Modify: `internal/cli/devwiki/command_test.go`
- Modify: `internal/ui/i18n.go`
- Test: `internal/app/devwikiapp/actions_test.go`

- [ ] Add `devwiki update` to the CLI.
- [ ] Resolve scope from the current project root lock file, falling back to global when the project lock does not exist.
- [ ] Filter check results down to DevWiki builtin entries only.
- [ ] If nothing changed, print the existing “all up to date” message.
- [ ] In TTY mode, show a default-all multiselect of changed DevWiki skills; in non-TTY mode, update all changed DevWiki skills.
- [ ] Reuse shared `skillapp` update helpers to apply selected updates.

### Task 5: Backward compatibility and docs

**Files:**
- Modify: `internal/app/devwikiapp/actions.go`
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `internal/devwiki/template/docs/README.md`
- Test: `internal/app/devwikiapp/actions_test.go`

- [ ] Migrate legacy DevWiki lock entries that still point at extracted temp paths to the stable builtin source before update checks.
- [ ] Document `skill add zatools/devwiki` and `devwiki update`.
- [ ] Keep root docs and embedded DevWiki docs aligned with the new command surface.

### Task 6: Final verification

**Files:**
- Test: `internal/skills/source_test.go`
- Test: `internal/app/skillapp/actions_test.go`
- Test: `internal/app/devwikiapp/actions_test.go`
- Test: `internal/cli/devwiki/command_test.go`

- [ ] Run targeted tests for the changed packages.
- [ ] Run `go test ./...`.
- [ ] Run `git diff --check`.
