# Mutation Safety Reference

> Shared reference for any skill that may write into `wiki/`, propose structural changes, or mutate code-related metadata.

## Core Rule

Medium-risk and high-risk mutations require **explicit confirmation** before writing.

Low-risk deterministic maintenance may proceed without extra approval when the workflow already implies it.

Questions to align should be handled through conversation first, not written to a default file. Save them to `wiki/outputs/` only when the user explicitly asks for a report.

## Risk Levels

### Low Risk

- append deterministic log entries
- refresh derived indexes
- refresh inline `sources.hash`
- update clearly stale generated output

### Medium Risk

- attach a feature to an existing capability
- add supporting workflow `code_refs`
- add secondary API or test entry points
- tighten a stale feature summary without changing scope
- update a workflow call-chain summary

### High Risk

- create a new capability
- merge or split capabilities
- create a new feature
- re-scope a feature to a different capability
- split or merge workflows
- replace primary `code_refs`

## Confirmation Protocol

Before a medium-risk or high-risk write:

1. show the proposal
2. separate facts from interpretation
3. call out the risk level
4. list the decisions needed from the user
5. wait for explicit confirmation

If multiple candidate capabilities, features, or workflows still compete, ask the user instead of choosing quietly.

## Proposal Content Requirements

A good mutation proposal includes what will change, why, evidence, uncertainty, and what will happen if accepted.

## Destructive Operations

Special handling is required for destructive actions such as `zatools devwiki tool reset`.

Always produce a dry-run plan first, list affected files or scopes, warn clearly when `raw/` is involved, and wait for explicit confirmation before execution.

## What NOT To Do

- Do not classify high-risk mutation as low risk for convenience.
- Do not hide a structural rewrite inside a small wording change.
- Do not overwrite capability or feature ownership silently.
- Do not treat user silence as confirmation.
- Do not hide capability-boundary, feature-ownership, or workflow-split changes inside automatic fixes.
