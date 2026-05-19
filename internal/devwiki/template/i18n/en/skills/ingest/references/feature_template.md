# DevWiki Feature Writing Template

> Target path: `wiki/features/<slug>.md`  
> Positioning: a Feature is a functional contract / functional knowledge archive. It is not the raw design document and not the implementation page.  
> Goal: preserve functional goals, behavior, rules, boundaries, interactions, and acceptance concerns so later agents can answer what the function is, when it triggers, which rules matter, what it relates to, and where to inspect implementation.

---

## 1. Responsibility

A Feature page should answer:

- What problem does this function solve?
- What does it look like to users, product, QA, ops, or system behavior?
- When does it trigger?
- What are the key rules?
- Which configs, status values, roles, and boundaries matter?
- How does it interact with other Features?
- What should testing and acceptance focus on?
- Which Workflow should be read for implementation?

A Feature page should not answer:

- where the code entry is;
- which class, module, or function implements it;
- how the call chain works;
- where state is physically read or written;
- how to change a branch in code;
- detailed troubleshooting commands and recovery steps.

Put those details in:

- `wiki/workflows/<slug>.md` for engineering entry, call chain, key implementation, code references, and change impact;
- `wiki/troubleshooting/<slug>.md` for symptoms, logs, diagnosis paths, and fixes;
- `raw/` for the full source document.

---

## 2. Feature / Workflow Boundary

Short distinction:

```text
Feature = functional behavior and rules
Workflow = implementation path and code
```

| Content | Feature | Workflow |
|---|---|---|
| Functional goal | detailed | may link |
| User scenario | detailed | do not expand |
| Trigger condition | detailed | map to code |
| Key business rule | detailed | explain implementation location |
| Config impact | detailed | explain read/validation/write path |
| Status/role meaning | functional meaning | code-level judgment/update |
| Decision table | functional rules | implementation branches |
| Call chain | do not write | detailed |
| File path / function name | do not write | detailed |
| Change impact | functional impact | code impact and regression scope |
| Test strategy | acceptance concerns | test entry and validation steps |

---

## 3. Writing Rules

### 3.1 Do Not Over-Summarize

Do not stop at "this feature is used for ...". Preserve:

- participants;
- trigger conditions;
- key actions;
- recovery/end conditions;
- scenarios that are not handled;
- configs that affect behavior.

### 3.2 Do Not Reproduce Raw Documents

Do not copy every original section, API field, database field, thread detail, or code structure into Feature. Extract functional knowledge.

### 3.3 Preserve Key Rules

A rule belongs in Feature if it affects:

- product explanation;
- test case design;
- developer impact analysis;
- agent answers about function behavior;
- ops judgment;
- later Workflow code mapping.

### 3.4 Implementation Details Are Links Only

Feature may say:

```text
Implementation location: [[<workflow-slug>]]
```

Do not write concrete files, functions, handlers, or call chains.

---

## 4. Recommended Template

```markdown
---
title: "<Feature Name>"
slug: "<feature-slug>"
status: draft
summary: "<one sentence explaining the core function>"
capabilities:
  - "[[<capability-slug>]]"
workflow: "<workflow-slug>"
related_features: []
troubleshooting: []
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
  - "<functional keyword>"
  - "<English/code/config keyword>"
---

# <Feature Name>

## Functional Summary

Use 3 to 6 bullets:

- what problem this function solves;
- who participates or is affected;
- core behavior;
- trigger condition;
- most important design rule or boundary.

## Background And Goal

Explain:

- why the function is needed;
- what problem it solves;
- risk if it does not exist;
- value for user, ops, product, or stability.

## Functional Scope

### In Scope

This Feature owns:

- ...

### Out Of Scope

This Feature does not own:

- ...

## Typical Scenarios

- Scenario 1:
- Scenario 2:
- Scenario 3:

## Functional Behavior

Describe behavior, not implementation:

- normal path;
- exceptional path;
- enabled/disabled behavior;
- behavior after config changes;
- interactions with related Features.

## Key Rules

| Rule | Condition | Behavior / Result | Notes |
|---|---|---|---|
|  |  |  |  |

## Key Concepts

| Concept | Meaning | Functional impact |
|---|---|---|
|  |  |  |

## Important Configs

| Config | Default | Values / Constraints | Behavior Impact |
|---|---|---|---|
|  |  |  |  |

## Relations To Other Features

| Related Feature | Relation | Impact |
|---|---|---|
|  |  |  |

## Boundary And Exceptions

| Scenario | Behavior | Notes |
|---|---|---|
|  |  |  |

## Observability

Fill only when logs, alarms, events, metrics, audit, or visible status matter.

| Observation | Trigger | Recovery / Cleanup | Notes |
|---|---|---|---|
|  |  |  |  |

## Acceptance Concerns

- Normal path:
- Exceptional path:
- Boundary conditions:
- Config/switch:
- Interaction impact:

## Engineering Entry

Implementation location: `[[<workflow-slug>]]`

Do not write concrete code paths here.

## Source Notes

Explain source coverage, uncertainty, conflicts, and version/applicability notes.

## Search Terms

List terms users and qmd are likely to search.
```
