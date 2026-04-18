# Mutation Safety Reference

> Shared reference for any skill that may write into `wiki/`, propose structural changes, or mutate code-related metadata.

---

## Core Rule

Medium-risk and high-risk mutations require **explicit confirmation** before write.

Low-risk deterministic maintenance may proceed without extra approval when the workflow already implies it.

---

## Risk Levels

### Low Risk

Examples:

- create a new document mirror
- refresh derived indexes
- append deterministic log entries
- update clearly stale generated output

These actions are usually reversible and do not redefine knowledge ownership.

### Medium Risk

Examples:

- append a document to an existing capability
- add supporting `code_refs`
- attach a change to an existing capability
- add secondary code clues or API references

These actions modify structure but usually do not redefine the system center.

### High Risk

Examples:

- create a new capability
- merge or split capabilities
- create or reclassify a change
- replace primary `code_refs`
- rewrite `change_classification`

These actions affect future retrieval, ownership, and planning behavior. They must never be auto-written silently.

---

## Confirmation Protocol

Before a medium-risk or high-risk write:

1. show the proposal
2. separate facts from interpretation
3. call out the risk level
4. ask for explicit confirmation

Good confirmation language:

- "This is a medium-risk attachment to an existing capability. Confirm before write."
- "This is a high-risk reclassification from `new` to `modify`. Confirm before write."

---

## When To Escalate To `/devwiki-refresh` Or `/devwiki-check`

Route work to `/devwiki-refresh` when:

- wiki knowledge drifts from raw or code
- path, symbol, or classification mismatches appear
- a user is correcting earlier mistakes

Route work to `/devwiki-check` when:

- the need is deterministic health validation
- the user wants a report-first scan
- the issue may be broken links, stale `source_hash`, missing reverse links, or stale `code_refs`

---

## Proposal Content Requirements

A good mutation proposal should include:

- what will change
- why it is justified
- what evidence supports it
- what remains uncertain
- what will happen if the proposal is accepted

If multiple candidate capabilities or changes still compete, do not choose one quietly. Ask the user instead.

---

## Destructive Operations

Special handling is required for destructive actions such as `/devwiki-reset`.

Always:

- produce a dry-run plan first
- list the affected files or scopes
- warn clearly when `raw/` is involved
- wait for explicit confirmation before execution

---

## What NOT To Do

- Do not classify a high-risk mutation as low risk just because it is convenient
- Do not hide a structural rewrite inside a small wording change
- Do not overwrite ownership or classification silently
- Do not treat user silence as confirmation
- Do not use `/devwiki-check --fix` to perform capability or change reclassification
