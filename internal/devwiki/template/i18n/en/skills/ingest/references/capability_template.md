# DevWiki Capability Writing Template

> Target path: `wiki/capabilities/<slug>.md`  
> Positioning: a Capability is the capability-boundary and capability-map page. It is not a feature detail page and not an implementation page.  
> Goal: explain what the system can do, why the capability exists, which Features it covers, where the boundary is, and how it collaborates with other capabilities.

---

## 1. Responsibility

A Capability page should answer:

- What is this system capability?
- Which class of business or technical problem does it solve?
- What value does it provide to users, product, QA, ops, or system stability?
- Which Features does it cover?
- What is outside the capability boundary?
- Which capabilities does it depend on or collaborate with?
- Where should a reader go for concrete feature rules?

A Capability page should not answer:

- full rules for a concrete feature;
- detailed state machines or decision tables;
- code entries, call chains, or function names;
- troubleshooting runbooks;
- detailed API fields;
- full design details for a Feature.

Put those details in:

- `wiki/features/<slug>.md` for functional behavior, rules, config, boundaries, and acceptance concerns;
- `wiki/workflows/<slug>.md` for engineering entries, call chains, code references, and change impact;
- `wiki/troubleshooting/<slug>.md` for symptoms, logs, diagnosis paths, and fixes;
- `raw/` for full original design documents.

---

## 2. Capability / Feature Boundary

Short distinction:

```text
Capability = what the system can do and where the capability boundary is
Feature = how a concrete function under that capability works
```

| Content | Capability | Feature |
|---|---|---|
| Capability definition | detailed | may link |
| Business value | detailed | brief |
| Capability boundary | detailed | only this feature's boundary |
| Covered functions | list and summary | detailed |
| Trigger conditions | do not expand | detailed |
| Key functional rules | summary only | detailed |
| State machine / decision table | do not write | keep if important |
| Code entry | do not write | do not write; link Workflow |
| Troubleshooting steps | do not write | link Troubleshooting |

---

## 3. Writing Rules

### 3.1 Capability Is The Upper Map

It should help a reader decide:

- which capability domain a question belongs to;
- which Feature to read next;
- which Features relate to each other;
- which topics are out of scope.

### 3.2 Do Not Copy Feature Content

Correct:

```markdown
- VIP failover: maintains the business access entry during active/standby changes. Detailed rules: [[feature-vip-failover]].
```

Incorrect:

```markdown
Copy every trigger condition, decision table, and state transition from the VIP failover Feature.
```

### 3.3 State The Boundary Clearly

The most important value of a Capability page is scope narrowing:

- what this capability owns;
- what it does not own;
- which capabilities it collaborates with;
- which questions should be routed elsewhere.

---

## 4. Recommended Template

```markdown
---
title: "<Capability Name>"
slug: "<capability-slug>"
status: draft
summary: "<one sentence explaining the problem this capability solves>"
features:
  - "<feature-slug>"
related_capabilities: []
sources:
  - path: "raw/designs/<source-file>.md"
    kind: design
    hash: ""
    title: ""
    confidence: medium
    notes: ""
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
search_terms:
  - "<business keyword>"
  - "<English/code/config keyword>"
---

# <Capability Name>

## Capability Summary

Use 3 to 6 bullets:

- which class of problem this capability solves;
- main scenarios;
- core covered Features;
- related capabilities;
- most important boundaries.

## Background And Value

Explain why this capability exists:

- business background;
- technical background;
- user/product/QA/ops/stability value;
- risk if the capability does not exist.

## Capability Boundary

### In Scope

This capability owns:

- ...

### Out Of Scope

This capability does not own:

- ...

## Covered Features

List Features with one-sentence summaries. Do not expand rules.

| Feature | Summary | Status |
|---|---|---|
| `[[<feature-slug>]]` |  | draft / active |

## Capability Relations

| Related Capability | Relation | Notes |
|---|---|---|
| `[[<capability-slug>]]` | depends on / depended on by / collaborates with / affects |  |

## Core Concepts

Record only capability-level concepts.

| Concept | Meaning | Notes |
|---|---|---|
|  |  |  |

## Typical Scenarios

- Scenario 1:
- Scenario 2:
- Scenario 3:

## Capability-Level Constraints

- ...

## Navigation

| User question type | Read first |
|---|---|
| Understand the capability | this Capability |
| Understand a concrete function | `wiki/features/<slug>.md` |
| Inspect implementation path | `wiki/workflows/<slug>.md` |
| Troubleshoot a failure | `wiki/troubleshooting/<slug>.md` |

## Acceptance Concerns

- Does the capability cover core scenarios?
- Are boundaries clear?
- Are covered Features linked?
- Are cross-capability dependencies explicit?

## Source Notes

Explain source coverage, uncertainty, conflicts, and version/applicability notes.

## Search Terms

List terms users and qmd are likely to search.
```
