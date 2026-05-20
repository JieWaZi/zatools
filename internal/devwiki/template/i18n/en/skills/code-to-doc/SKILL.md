---
name: "devwiki-code-to-doc"
description: "Use this skill when reverse-generating DevWiki documentation from real code, API URLs, key files, key functions, page paths, routes, configuration items, or log keywords. This Skill focuses on guiding the Agent to trace code step by step, understand call chains, identify key logic, state reads/writes, configuration handling, exception paths, and change impact. Final page structures reuse the Capability / Feature / Workflow / Troubleshooting templates defined by ingest; this Skill does not redefine those templates."
argument-hint: "<feature name, API URL, key file, key function, route, configuration item, or log keyword>"
---

# /devwiki-code-to-doc

## 1. Prerequisites

Before starting, read the general constraints:

- `references/evidence-grounding.md`
- `references/zatools-qmd.md`
- If writing, reclassifying, or performing destructive operations, also read `references/mutation-safety.md`
- If tracing code, attributing code, or verifying implementation, also read `references/code-tracing.md`

When generating or updating pages, do not redefine templates in this Skill. Reuse the templates from the ingest system:

- Capability: reuse `references/capability_template.md`
- Feature: reuse `references/feature_template.md`
- Workflow: reuse `references/workflow_template.md`
---

## 2. Core Positioning

Code-to-doc is not a template-writing Skill. It is a code understanding and evidence tracing Skill.

It answers:

- which code entry point to start from;
- how to trace downward step by step;
- when the trace is deep enough to stop;
- which code evidence can support the Wiki;
- which conclusions belong in Workflow;
- which functional semantics should be synchronized to Feature;
- which capability boundary changes require a Capability proposal;
- which exceptions, logs, or recovery paths should be synchronized to Troubleshooting;
- how to handle conflicts between code, raw, and wiki;
- when to ask the user.

The final page structure is handled by ingest templates. This Skill only turns code understanding into evidence and proposals ready for writing.

---

## 3. Default Output

Default output priority:

1. `wiki/workflows/<slug>.md`
    - code entry point;
    - call chain;
    - key logic;
    - state reads/writes;
    - configuration handling;
    - exception and recovery implementation;
    - code references;
    - test references;
    - change impact.

2. `wiki/features/<slug>.md`
    - Update only when code tracing confirms functional behavior, parameter semantics, linkage, boundaries, or acceptance concerns.
    - Feature must not contain code references; it only links to Workflow.
    - Feature `sources` must not contain code file paths or `kind: code`; code evidence belongs only in the corresponding Workflow `code_refs`.

3. `wiki/troubleshooting/<slug>.md`
    - Update only when the input anchor is a log, error code, symptom, or code tracing confirms a diagnosis/recovery path.

4. `wiki/capabilities/<slug>.md`
    - Output a proposal only when code tracing reveals capability boundary, covered feature, or capability relationship changes.
    - Do not expand Capability directly by default.

---

## 4. Input Anchors

The user should provide at least one anchor:

- feature name;
- API URL;
- route;
- page path;
- key file;
- key function;
- configuration item;
- log keyword;
- error code;
- known Feature / Workflow slug.

If only a feature name is provided, do not ask immediately. Search wiki, raw, and code first.  
Ask only after multiple rounds still fail to locate a stable implementation.

---

## 5. Source Priority

Investigate in this order by default:

1. `wiki/`
    - Check existing capability, feature, workflow, and troubleshooting pages first to avoid duplicate work.
2. `raw/`
    - Then check requirements, design docs, feature descriptions, API docs, and test materials to understand historical intent.
3. Local code
    - Use current code to confirm actual implementation and correct documentation drift.
4. User clarification
    - Ask only when confirmation cannot continue or a user decision is required.

Key principles:

- `wiki/` and `raw/` are clues and history, not final implementation truth.
- Current implementation conclusions must be supported by current code.
- If code conflicts with wiki/raw, explicitly record it in the proposal or the Workflow's “Code Verification Conclusion / Source Notes”.
- Do not treat historical design as current implementation by default.
- Do not present code reality as product design unless explicitly marked as “implementation status”.

---

## 6. Code Tracing Method

### 6.1 Start from Stable Entry Points

Preferred entry order:

1. API URL / route;
2. controller / handler;
3. main service method;
4. key file;
5. configuration item;
6. log keyword;
7. reverse lookup from caller;
8. global search.

Do not stop after finding the first same-name function.  
Confirm whether it is truly part of the current functional path.

---

### 6.2 Trace Downward Step by Step

Expand in this order:

1. Entry layer
    - where the request, command, task, thread, or event enters;
    - where parameters come from;
    - whether auth, validation, or feature switches exist.

2. Dispatch layer
    - how controller, router, dispatcher, handler, or registry dispatches;
    - whether dynamic registration, reflection, config mapping, or template generation exists.

3. Service layer
    - which service owns the main business logic;
    - where key branches are;
    - which helpers / managers / adapters are called.

4. State layer
    - where configuration is read;
    - where state is read;
    - where data is written;
    - whether cache, state files, database, memory variables, or external interfaces are involved.

5. Side-effect layer
    - whether configuration is delivered;
    - whether external services are called;
    - whether services are started/stopped;
    - whether logs, alarms, or events are written;
    - whether other features are affected.

6. Exception layer
    - how failure is handled;
    - whether retry, rollback, degradation, skip, or recovery exists;
    - whether timeout, concurrency, locks, or protection windows exist.

7. Test layer
    - whether unit tests, integration tests, scripts, or manual verification entry points exist;
    - which scenarios should regress for current changes.

---

### 6.3 Trace Depth

Trace at least far enough to explain:

- where the entry point is;
- how the call chain flows;
- what the core logic is;
- what the key branches are;
- where state is read;
- where state is written;
- how configuration is read, validated, synchronized, or delivered;
- how exceptions, failures, rollback, or recovery are handled;
- what changing this code may affect;
- what tests or verification entry points exist.

Stop when:

- Workflow can explain the engineering location without guessing;
- further tracing would become line-by-line code explanation;
- downstream belongs to an external system, sub-repository, or unavailable dependency;
- dynamic dispatch, reflection, or template generation prevents further static confirmation;
- multiple candidate entry points remain after several searches and require user confirmation.

---

## 7. Evidence Recording

Code evidence must be recorded clearly.

Recommended proposal output:

```markdown
## Code Evidence Summary

| Evidence | Type | Notes | Confidence |
|---|---|---|---|
| `<path>` / `<symbol>` | file/function/config/test |  | high/medium/low |
```

Mark:

- confirmed entry points;
- confirmed call chain;
- confirmed key logic;
- confirmed state reads/writes;
- confirmed configuration handling;
- confirmed exception paths;
- confirmed test entry points;
- unconfirmed dynamic branches;
- conflicts with wiki/raw;
- anchors requiring user confirmation.

Evidence requirements:

- Do not fabricate code paths, function names, module names, or interface names.
- Uncertain code clues must not be marked high confidence.
- Put only verified code in `code_refs`.
- `code_refs` belongs only in Workflow, never in Feature or Capability.
- Feature `sources` may record only raw material, existing Wiki pages, or user-provided non-code context. Even when a Feature conclusion comes from code verification, do not write code file paths, function names, or `kind: code` into Feature.

---

## 8. Ownership Decision

| Code tracing discovery | Destination |
|---|---|
| Entry points, call chain, classes, modules, functions, handlers | Workflow |
| State reads/writes, configuration reads, validation, synchronization, delivery | Workflow |
| Exception, failure, retry, rollback, recovery implementation | Workflow |
| Implementation differences versus raw/wiki | Workflow |
| Change impact, test references, verification suggestions | Workflow |
| Functional behavior, parameter semantics, boundary rules confirmed from code | Feature |
| Functional linkage, edge cases, acceptance concerns confirmed from code | Feature |
| Logs, error codes, diagnostic path, fix/recovery path | Troubleshooting |
| Capability boundary, covered feature, capability relationship changes | Capability proposal |

Do not redefine these page section templates in code-to-doc. Read and follow the corresponding ingest templates when writing pages.

---

## 9. Conflict Handling

If code conflicts with wiki/raw:

1. do not silently choose one side;
2. do not directly modify raw;
3. do not write historical design as current implementation;
4. do not write code reality as product design unless explicitly marked;
5. output the difference in the proposal;
6. after confirmation:
    - functional semantic differences go to Feature;
    - implementation differences go to Workflow;
    - troubleshooting differences go to Troubleshooting;
    - capability boundary differences go to Capability proposal.

If it is unclear which source is newer, ask for confirmation.

---

## 10. Question Rules

The Agent may attempt several independent searches first. If confirmation is still impossible, stop and ask the user.

Must ask when:

- only a feature name is provided, no stable entry point is found, and multiple candidate implementations remain after multiple searches;
- API URL, page path, or key function cannot be found in the codebase;
- code takes effect indirectly through dynamic registration, reflection, configuration delivery, or template generation, and cannot be statically confirmed;
- key external systems, gateways, or third-party interfaces have no local implementation or documentation;
- a same-name file already exists and a decision is needed between “update” and “create”;
- code clearly conflicts with existing wiki/raw and it is unclear which is newer;
- pages need to be created, split, merged, or renamed.

Question style:

- keep questions short;
- ask only for the missing anchor;
- do not ask a long list of questions at once;
- do not ask vaguely, “Can you provide more information?”

Example:

```text
I found two candidate entry points in the code: `a` and `b`. Which chain do you want me to start from?
```

```text
I could not find interface `<URL>` in the local code. Is this interface in a gateway, sub-repository, or external service?
```

---

## 11. Write Proposal

Before writing, output a proposal.

```markdown
# Code-to-Doc Proposal

## Input Anchors

## Materials Checked

| Source | Result |
|---|---|

## Code Tracing Summary

| Layer | Discovery | Evidence |
|---|---|---|
| Entry layer |  |  |
| Dispatch layer |  |  |
| Service layer |  |  |
| State layer |  |  |
| Side-effect layer |  |  |
| Exception layer |  |  |
| Test layer |  |  |

## Proposed Writes

| Page | Type | Action | Reason | Confidence |
|---|---|---|---|---|

## Differences from Wiki / Raw

## Questions Requiring Confirmation

## Content Not Written
```

High-risk changes must wait for user confirmation:

- creating pages;
- splitting pages;
- merging pages;
- renaming pages;
- changing the primary Feature / Workflow relationship;
- writing current implementation conclusions that conflict with raw/wiki;
- deleting or downgrading old content.

---

## 12. Apply After Confirmation

After user confirmation:

1. read the ingest template for the target page type;
2. create or update `wiki/workflows/<slug>.md`;
3. update `wiki/features/<slug>.md` if necessary;
4. update `wiki/troubleshooting/<slug>.md` if necessary;
5. output a Capability adjustment proposal if necessary;
6. update `wiki/index.md`;
7. update `wiki/glossary.md`;
8. append `wiki/log.md`;
9. run or prompt:

```bash
zatools qmd update
zatools qmd status
```

---

## 13. Prohibited Actions

### 13.1 Investigation Prohibitions

- Do not guess behavior from names only.
- Do not read code directly without checking wiki/raw first.
- Do not stop after finding the first same-name function.
- Do not skip key branches, helper calls, or data flow.
- Do not interrupt the user frequently while independent investigation can continue.
- Do not ask vague questions like “Can you provide more information?”

### 13.2 Writing Prohibitions

- Do not redefine Capability / Feature / Workflow templates in this Skill.
- Do not write unverified inference as fact.
- Do not fabricate code paths, functions, interfaces, or module names.
- Do not put code references in Capability or Feature.
- Do not put code file paths, function names, or `kind: code` in Feature `sources`.
- Do not copy the full Feature business description into Workflow.
- Do not turn Workflow into line-by-line code explanation.
- Do not write historical raw design as current implementation by default.
- Do not silently handle code-documentation conflicts.
- Do not create duplicate pages without confirmation.
- Do not skip proposal and write directly.

### 13.3 Granularity Prohibitions

- Do not split into multiple workflows just because multiple APIs, helpers, or branches appear.
- Do not split one complete functional chain into scattered code-fragment pages.
- Do not force missing external-system implementation into local implementation.
- Do not write definite conclusions when dynamic dispatch cannot be confirmed.

---