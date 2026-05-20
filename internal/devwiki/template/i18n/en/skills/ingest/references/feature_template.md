# DevWiki Feature Writing Template v4 (Unified Chinese Version)

> **Applicable Location:** `wiki/features/<slug>.md`
> **Purpose:** A "Feature" page serves as a "Functional Contract / Functional Knowledge Archive." It is *not* a detailed design document, nor is it a technical explanation of the code implementation.
> **Goal:** To use clear language to document functional objectives, core behaviors, key rules, boundaries, interdependencies, and acceptance criteria. This enables future AI Agents to accurately answer questions such as: "What is this feature?", "When is it triggered?", "What rules apply?", "What is it related to?", and "Where can I find the implementation details?"

---

## I. Responsibilities of a Feature Page

A Feature page should answer:

- What problem does this feature solve?
- How does this feature manifest itself to users, the product team, QA, and operations?
- When is this feature triggered?
- What are the key rules governing this feature?
- What are the critical configurations, states, roles, and boundaries involved?
- How does this feature interact or interoperate with other features?
- What specific points should be focused on during testing and acceptance?
- If one needs to examine the implementation, which workflow document should be consulted?

A Feature page is *not* responsible for answering:

- Where the code entry point is located;
- Which specific class, module, or function handles the implementation;
- The exact details of the call chain/execution flow;
- Which specific variable or file is used to read/write state data;
- How the code needs to be modified within a specific branch;
- Detailed troubleshooting commands or repair steps.

These details should be placed in:

- `wiki/workflows/<slug>.md`: Engineering entry points, call chains, key implementation details, code references, and impact analysis for modifications;
- `wiki/troubleshooting/<slug>.md`: Symptoms of failure, relevant logs, diagnostic paths, and suggested fixes;
- `raw/`: The full text of the original design documents. ---

## II. The Boundary Between Features and Workflows

A one-sentence distinction:

```text
Feature = What are the functional behaviors and rules?
Workflow = What is the implementation path and code flow?
```

| Content                       | Feature (Functional View) | Workflow (Implementation View) |
|-------------------------------|---------------------------|--------------------------------|
| Functional Objective          | Detailed description      | Can reference Feature doc      |
| User Scenarios                | Detailed description      | Do not elaborate on code flow  |
| Functional Trigger Conditions | Detailed description      | Mappable to specific code      |
| Key Business Rules            | Detailed description      | Indicate where implemented     |
| Config Impact on Behavior     | Detailed description      | Indicate where read/applied    |
| State/Role Semantics          | Functional-level explanation | Explain how code determines this |
| Decision Tables/Policy Matrices | Retain functional rules   | Expand into implementation branches |
| Call Chain                    | Do not include            | Detailed description           |
| File Paths/Function Names     | Do not include            | Detailed description           |
| Impact of Changes             | List functional impact only | List code-level impact         |
| Troubleshooting Steps         | Link to guide only        | Detailed troubleshooting steps |

---

## III. Writing Principles

### 3.1 Do Not Over-Summarize

Do not simply write a single sentence like, "This feature is used for..." You should clearly specify:

- Who is involved;
- The trigger conditions;
- The key actions;
- The recovery or termination conditions;
- Which scenarios are *not* handled;
- Which configurations affect the behavior.

### 3.2 Do Not Replicate Original Documentation

Do not copy *every* section, interface field, database field, thread detail, and code structure from the original requirements or design documents into the Feature documentation.

Feature documentation should be distilled into "functional-level knowledge."

### 3.3 Retain Key Rules

If a specific rule impacts *any* of the following scenarios, it should be included in the Feature documentation:

- Explaining the product to others;
- Designing test cases;
- Developers understanding the impact *before* making code changes;
- Support agents answering functional questions;
- Operations/SRE teams diagnosing functional behavior;
- Locating relevant code via the Workflow documentation later on.

### 3.4 For Implementation Details, Provide Only Entry Links

In the Feature documentation, you may write:

```text
For implementation details/location, see: [[<workflow-slug>]]
```

Do *not* list specific files, functions, handlers, or call chains here; these belong in the Workflow documentation. ---

## IV. Recommended Template

```markdown

---
title: "<Feature Name>"
slug: "<feature-slug>"
status: draft
summary: "<One-sentence description of the core feature>"
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

- "<Chinese keywords>"

- "<English/code/configuration keywords>"

---

# <Feature Name>

## Feature Summary

Use 3 to 6 Overall Function Description:

- What problem does this function solve?

- Who is involved or affected?

- What is the core behavior?

- When is it triggered?

- What are the most important design principles or boundaries?

## Background and Objectives

Briefly describe the background and objectives of the function.

Suggested content includes:

- Why is this function needed?

- What problem does it solve?

- What risks are there if this function is not implemented?

- Its value to users, operations, product, or system stability.

## Function Scope

### Within Scope

This function is responsible for:

- ...
### Outside Scope

This function is not responsible for:

- ...
## Typical Scenarios

List typical scenarios; it is not required to cover all implementation branches.

- Scenario 1:

- Scenario 2:

- Scenario 3:

## Function Behavior

Describe in words how the function works, focusing on "behavior" rather than "implementation".

Recommended coverage:

- Normal paths;

- Abnormal paths;

- Behavior after enabling/disabling;

- Behavior after configuration changes;

- Interaction with related functions.

## Key Rules

Consolidate the core rules affecting function judgment. This is the most important section on the Feature page.

This can be done using text or tables.

| Rule | Condition | Action / Result | Description |
|---|---|---|---|
|  |  |  |  |

Writing Requirements:

- Avoid vague or generic descriptions;
- Do not write actual function implementations;
- Do not omit any rules that could impact testing, explanation, or modification risks;
- If there are too many rules, retain only the critical, feature-level rules; place the complete implementation mapping in the workflow documentation.

## Key Concepts

Document the concepts, states, roles, configurations, or terminology essential for understanding this feature.

| Concept | Meaning | Impact on Feature Behavior |
|---|---|---|
|  |  |  |

## Important Configurations

Document only the key configurations that influence the feature's behavior.

| Configuration | Default Value | Accepted Values ​​/ Constraints | Behavioral Impact |
|---|---|---|---|
|  |  |  |  |

If there are numerous configurations, retain only the critical items; the complete table of fields may be placed in the workflow or raw documentation.

## Relationships with Other Features

Describe the relationships between this feature and other features.

| Related Feature | Relationship | Impact |
|---|---|---|
|  |  |  |

## Edge Cases and Exceptions

Describe edge cases that are prone to misunderstanding or errors, or that require special handling logic.

| Scenario | Behavior | Description |
|---|---|---|
|  |  |  |

## Observability

Complete this section only if the feature involves logs, alerts, events, metrics, auditing, or user-observable states.

| Observability Item | Trigger Condition | Recovery / Cleanup Condition | Description |
|---|---|---|---|
|  |  |  |  |

If there are extensive observability details, the Feature document should retain only the trigger and recovery rules; place the complete troubleshooting details in the troubleshooting documentation.

## Acceptance Criteria

Synthesize the key points for testing and acceptance from a feature-centric perspective. - Normal Path:
- Exceptional Path:
- Boundary Conditions:
- Configuration/Feature Flags:
- Interdependencies/Side Effects:
- Observability/Alerts:

## Engineering Entry Point

For implementation details, see: `[[<workflow-slug>]]`

If a workflow has not yet been established:

- Workflow to be created: `<workflow-slug>`
- Issues requiring code verification:
- ...

## Related Troubleshooting

- `[[<troubleshooting-slug>]]`

## Source Attribution

Explain the sources used, any conflicts, uncertain content, and the status of code verification.

- Sources:
- Conflicts:
- Uncertainties:
- Details currently omitted:
- Pending confirmation:

## Keywords

Used for lexical search within QMD. Includes Chinese aliases, English terminology, configuration items, status names, and expressions users might search for.

- ...
```

The `sources` field within a Feature should *only* record raw data, existing Wiki entries, or non-code materials provided by users. Even if a Feature is reverse-engineered from code traces, do *not* list code file paths, function names, handlers, call chains, or `kind: code` within the Feature's `sources`. Code-based evidence *must* be recorded in the corresponding Workflow's `code_refs`; the Feature document should only point to the implementation details via the `workflow` field or links within the main body text.

---

## V. Optional Extended Sections

The following sections are not mandatory by default. Include them only if they are critical to understanding the functionality.

### State / Role Model

Applicable to features that are state-driven, role-driven, or involve complex lifecycles.

| State / Role | Meaning | Behavioral Impact |
|---|---|---|
|  |  |  |

### Decision Matrix

Applicable to features involving conditional matrices, policy tables, or combined state processing.

| Condition | Additional Condition | Behavior / Result | Description |
|---|---|---|---|
|  |  |  |  |

### Timing & Reliability

Applicable to features involving cycles, retries, timeouts, debouncing, confirmation windows, or protection windows. | Mechanism | Threshold / Default Value | Behavioral Impact |
|---|---|---|
|  |  |  |

### Data or Persistent State

Applicable to scenarios where data, configuration snapshots, feature flags, or state files influence functional behavior.

| Data / State | Purpose | Lifecycle |
|---|---|---|
|  |  |  |

### Interfaces and Integrations

Applicable to scenarios where interfaces or service interactions themselves constitute the functional contract.

| Interface / Integration Point | Triggering Behavior | Description |
|---|---|---|
|  |  |  |

### Compatibility and Upgrades

Applicable to scenarios where upgrades, rollbacks, default toggles, or version compatibility influence functional behavior.

- ...

---

## VI. Extracting Design Signals

Before generating the Feature document, the Agent should first extract "design signals" from the raw source materials:

```markdown
## Design Signals

### Functional Goals

### Core Behaviors

### Key Rules

### Key Concepts / States / Roles

### Key Configurations

### Key Processes / Flows

### Key Interactions

### Reliability / Recovery / Safeguards

### Observability / Alerting

### Boundaries / Constraints

### Acceptance Criteria

### Implementation Details for the Workflow Section

### Troubleshooting Details for the Troubleshooting Section

### Open Questions / Items for Clarification
```

The purpose of design signals is to prevent omissions; it does not imply that the generated Feature document must contain a corresponding section for every signal listed here.

---

## VII. Retention Policy

### 7.1 Must Be Included in the Feature Document

If present, the following content should typically be included in the Feature document:

- Rules that directly determine functional behavior;
- State changes that are observable by users or during testing;
- Rules governing feature toggles, configurations, interdependencies, or compatibility;
- Rules determining whether to trigger alerts, recovery actions, isolation, retries, or failures;
- Significant exception scenarios and boundary conditions;
- Functional constraints that impact the risk associated with future code modifications;
- Key coupling relationships with other features or components. ### 7.2 Place in Workflow

The following items should be placed in the Workflow section by default:

- Code entry points;
- Call chains;
- Classes, modules, and functions;
- Handlers;
- Thread implementations;
- Specific locations for state reads and writes;
- Specific code branches;
- Verification of discrepancies between implementation and design;
- Files affected by modifications, and corresponding test files.

### 7.3 Place in Troubleshooting

The following items should be placed in the Troubleshooting section by default:

- Specific log keywords;
- Error codes;
- Troubleshooting commands;
- Diagnostic paths;
- Remediation steps;
- Experience in handling field issues.

### 7.4 Optional Content

The following content is generally *not* included in a Feature:

- Empty sections from the original documentation template;
- Environmental descriptions unrelated to functional behavior;
- Explanations consisting solely of screenshots;
- Granular lists of fields that do not impact functional behavior;
- Outdated or unverified historical information.

---

## 8. Feature Quality Checklist

Perform a line-by-line check before finalization:

- Are the functional objectives and core behaviors clearly articulated?
- Are key rules—those critical for subsequent Q&A—retained?
- Has the direct, verbatim copying of detailed design specifications been avoided?
- Has over-summarization—which could lead to the omission of key rules—been avoided?
- If the feature is state- or role-dependent, are the key states or roles explained?
- If the feature involves critical decision-making rules, are the feature-level rules retained?
- If the feature is influenced by configuration settings, are the key configurations and their behavioral impacts retained?
- If the feature involves exceptions, recovery, self-healing, or rollback mechanisms, are the key conditions and actions retained?
- If the feature includes observability or alerting capabilities, are the triggering, recovery, and cleanup rules retained?
- Is the relationship with other features clearly defined?
- Have implementation details been delegated to the `workflow` section?
- Is every key fact supported by a corresponding source reference?
- Have code file paths or `kind: code` entries been excluded from the `sources` section?
- Has any uncertain content been explicitly noted within the source descriptions?