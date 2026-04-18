# DevWiki QMD-First Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make DevWiki treat qmd as the primary retrieval engine when healthy, add a dedicated qmd sync skill for existing workspaces, and keep qmd indexes fresh after DevWiki write flows.

**Architecture:** Keep one-time collection registration in setup/sync flows, add explicit qmd helper commands in the DevWiki CLI, and update built-in skill templates so write flows run post-write qmd refresh while read flows use qmd-first retrieval with fallback guidance. Fix the current collection-add command shape so generated qmd commands match qmd's actual CLI contract.

**Tech Stack:** Go, Cobra, Go `testing`, embedded DevWiki template docs/skills, qmd CLI integration

---

### Task 1: Fix qmd collection registration command generation

**Files:**
- Modify: `internal/devwiki/tools.go`
- Modify: `internal/devwiki/tools_test.go`

- [ ] Add a failing test that asserts `BuildQMDCollectionCommands` emits `qmd collection add <path> --name <name>`.
- [ ] Run `go test ./internal/devwiki -run TestLoadSearchConfigAndBuildQMDCommands`.
- [ ] Update the command builder to match qmd's CLI ordering while preserving absolute-path normalization.
- [ ] Re-run `go test ./internal/devwiki -run TestLoadSearchConfigAndBuildQMDCommands`.

### Task 2: Add explicit DevWiki qmd helper commands

**Files:**
- Modify: `internal/cli/devwiki/command.go`
- Modify: `internal/cli/devwiki/command_test.go`
- Modify: `internal/devwiki/tools.go`
- Modify: `internal/devwiki/tools_test.go`

- [ ] Add failing CLI tests for `zatools devwiki tool qmd update`, `embed`, and `status`.
- [ ] Run `go test ./internal/cli/devwiki -run TestDevwikiToolCommands`.
- [ ] Add reusable qmd command execution helpers in `internal/devwiki/tools.go` and wire new Cobra subcommands in `internal/cli/devwiki/command.go`.
- [ ] Add focused helper tests in `internal/devwiki/tools_test.go` where pure command generation logic is covered.
- [ ] Re-run `go test ./internal/cli/devwiki ./internal/devwiki`.

### Task 3: Add the dedicated `devwiki-qmd-sync` built-in skill

**Files:**
- Create: `internal/devwiki/template/i18n/zh/skills/qmd-sync/SKILL.md`
- Create: `internal/devwiki/template/i18n/en/skills/qmd-sync/SKILL.md`
- Modify: `internal/devwiki/project_test.go`
- Modify: `internal/devwiki/template/docs/README.md`
- Modify: `internal/devwiki/template/docs/AGENTS.md`
- Modify: `internal/devwiki/template/docs/CLAUDE.md`

- [ ] Add or extend tests that verify extracted built-in skills include the new qmd sync skill and generated docs mention it.
- [ ] Run `go test ./internal/devwiki -run 'TestExtractBuiltinSkillsMaterializesSharedReferencesIntoEachSkill|TestGenerateProjectRendersReadmeAndRuntimeTemplates'`.
- [ ] Add the new skill templates with explicit collection registration, repair, refresh, and qmd status guidance.
- [ ] Update generated DevWiki docs/runtime schema docs to list and describe `devwiki-qmd-sync`.
- [ ] Re-run the targeted DevWiki project tests.

### Task 4: Make write skills refresh qmd and read skills use qmd-first retrieval

**Files:**
- Modify: `internal/devwiki/template/i18n/zh/skills/init/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/ingest/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/refresh/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/edit/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/ask/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/scope/SKILL.md`
- Modify: `internal/devwiki/template/i18n/zh/skills/feature-doc/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/init/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/ingest/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/refresh/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/edit/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/ask/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/scope/SKILL.md`
- Modify: `internal/devwiki/template/i18n/en/skills/feature-doc/SKILL.md`

- [ ] Update write-flow skills so successful writes trigger `zatools devwiki tool qmd update`, and only recommend `embed` when semantic freshness is required.
- [ ] Update read-flow skills so healthy qmd runs use `qmd query`/`qmd get` style retrieval first, with explicit degraded-mode fallback when qmd is unavailable.
- [ ] Mirror the behavior in both `zh` and `en` templates.

### Task 5: Verify end-to-end behavior and docs

**Files:**
- Modify: `README.md` (if the repo-level docs need the new qmd-first behavior surfaced)
- Modify: `AGENTS.md` (if repository guidance needs the new qmd skill/flow surfaced)

- [ ] Run `go test ./internal/cli/devwiki ./internal/devwiki ./internal/app/devwikiapp`.
- [ ] Run `go test ./...` if targeted tests pass.
- [ ] Run a real command dry-run: `go run ./cmd/zatools devwiki tool qmd sync --root <fixture-root>` and inspect emitted qmd commands.
- [ ] Run `git diff --check` and inspect final diffs for template parity.
