---
name: "devwiki-ingest"
description: "Used when a user provides design documents, requirements specs, interface specs, configuration guides, deployment guides, test plans, meeting minutes, troubleshooting logs, code snippets, or discussion conclusions, and requests that these materials be digested, imported, used to generate Wiki pages, or used to build a knowledge base. This Skill transforms raw source materials into capabilities, features, workflows, troubleshooting guides, terminology, relationships, and questions requiring conversational clarification. This version exclusively uses Chinese titles and clearly defines the boundaries between the three layers: Capability, Feature, and Workflow—where Capability represents the capability map, Feature represents the functional contract, and Workflow represents the implementation path."
argument-hint: "<Document Path, Directory, Text Snippet, or Scope for Ingestion>"
---

# /devwiki-ingest

> Read the general constraints first:
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - If writing, reclassifying, or destructive operations are involved, also read `references/mutation-safety.md`
> - If code tracing, code attribution, or implementation verification is involved, also read `references/code-tracing.md`
> - Before generating a Capability page, prioritize reading `references/capability_template.md`
> - Before generating a Feature page, prioritize reading `references/feature_template.md`
> - Before generating a Workflow page, prioritize reading `references/workflow_template.md`

Transform a single source document or a batch of raw materials into maintainable knowledge. Do not simply write a free-form prose summary; you must first categorize the content, control the granularity, and propose a structure. Only commit the content to storage after the user has confirmed any medium-to-high-risk details.

The core objectives of this Skill are:

1. To extract reusable knowledge from the raw source materials, rather than merely copying the original text verbatim;
2. To preserve key design signals that will be relied upon during subsequent Q&A sessions, development, testing, and troubleshooting;
3. To avoid redundant maintenance across the Capability, Feature, Workflow, and Troubleshooting categories;
4. To avoid creating "thin" Feature entries that prevent subsequent Agent interactions from answering questions regarding rules, boundaries, or interdependencies;
5. To use Chinese exclusively for Markdown headings, avoiding any mixing of Chinese and English section titles. ---

## I. The Three-Tiered Core Model

DevWiki's core knowledge is structured around three distinct tiers and a troubleshooting section:

| Tier | Path | Purpose |
|---|---|---|---|
| Capability | `wiki/capabilities/<slug>.md` | Business/System Capabilities: Capability boundaries, intended effects, covered features, and inter-capability collaboration. |
| Feature | `wiki/features/<slug>.md` | Functional Contracts/Knowledge: Functional objectives, core behaviors, key rules, key concepts, important configurations, interdependencies, boundaries, and acceptance criteria. |
| Workflow | `wiki/workflows/<slug>.md` | Engineering Perspective (Code-Oriented): Entry points, call chains, key logic, code references, impact analysis for modifications, and implementation variance checks. |
| Troubleshooting | `wiki/troubleshooting/<slug>.md` | Symptoms, logs, error codes, diagnostic paths, evidence, and suggested fixes. |
---

## II. Boundaries of the Capability / Feature / Workflow Tiers

A one-sentence distinction:

```text
Capability = What capabilities does the system possess, and what are their boundaries?
Feature = What are the behaviors and rules of a specific function?
Workflow = How is the function implemented within the codebase?
```

Content must not be duplicated across the three-tier pages. Each tier maintains only the "authoritative facts" for which it is responsible; other tiers should provide only summaries and links.

| Tier | Core Questions | Authoritative Content | What *Not* to Include |
|---|---|---|---|
| Capability | What capabilities does the system possess? What are their boundaries? | Capability definitions, business value, scope, covered Features, inter-capability relationships, capability-level constraints. | Specific functional rules, state machines, decision tables, code paths. |
| Feature | How does this function behave? What are its key rules? | Functional objectives, user scenarios, trigger conditions, core behaviors, key concepts, important configurations, boundary exceptions, acceptance criteria. | Code entry points, function names, call chains, implementation branches, complete troubleshooting steps. |
| Workflow | How is this function implemented in the code? | | Code Entry Point, Call Chain, Classes/Modules/Functions, State Read/Write, Configuration Handling, Exception Implementation, Test References, Impact Analysis | Comprehensive Business Context, Detailed Restatement of Feature Rules, Explanation of Capability Value |

---

### 2.1 Referencing Conventions Between the Three Layers

Correct Relationships:

```text
Capability → List and link to Features
Feature → Describe functional rules and link to Workflows
Workflow → Map Feature rules to code implementations
Troubleshooting → Link to Features/Workflows and provide troubleshooting paths
```

Recommended Syntax:

* [[feature-vip-failover]]: Responsible for VIP takeover behavior; detailed rules can be found on the Feature page. * For implementation details regarding positioning, see: [[workflow-vip-failover]]

---

### 2.2 Three-Layer Classification Table

| Information Type | Writing Location | Description |
|---|---|---|
| Capability Definition, Business Value, Capability Boundaries | Capability | Serves as the authoritative source of truth for the Capability page |
| Covered Functionality | Capability | Lists only Feature summaries and links |
| Functional Objectives, Functional Behaviors, Functional Rules | Feature | Serves as the authoritative source of truth for the Feature page |
| User Scenarios, Trigger Conditions, Boundary Exceptions | Feature | Intended for understanding and testing purposes |
| Functional Implications of States/Roles | Feature | Explains only the functional impact |
| Specific State Determination Code | Workflow | Records the code entry point and determination location |
| Functional Outcomes of Decision Rules | Feature | Retains rule summaries or key tables |
| Code Branches for Decision Rules | Workflow | Maps to the specific implementation location |
| Impact of Configuration on Behavior | Feature | Describes the functional impact |
| Code for Configuration Reading/Validation/Distribution | Workflow | Records the implementation path |
| Code Paths, Function Names, Call Chains | Workflow | *Forbidden* on the Feature page |
| Failure Symptoms, Logs, Remediation Steps | Troubleshooting | Feature page provides only links |
| Files Affected by Changes, Test Files | Workflow | Describes the engineering impact |

---

## III. Inputs
- `raw/**/*.md`: Raw requirements, designs, functional specifications, test plans, etc., that have already been placed in the DevWiki;
- `config/project.yaml`: Project name, language, agent, and code repository configuration;
- `config/search.yaml`: QMD collection and model configuration.

Supported Document Types:

- Design documents, requirements documents, interface specifications, configuration guides, deployment guides, test plans;
- Meeting minutes, troubleshooting records, change review records;
- Code logic pasted by users, runtime logs, or conclusions drawn from user-AI discussions. ---

## IV. Outputs

Each ingest operation is permitted to create or update only the following files:

- `wiki/capabilities/<slug>.md`
- `wiki/features/<slug>.md`
- `wiki/workflows/<slug>.md`
- `wiki/troubleshooting/<slug>.md`
- `wiki/relations.yml`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `wiki/outputs/<slug>.md`: Only when the user explicitly requests that the report be saved.

---

## V. Granularity Rules

By default, a single functional topic corresponds to at most:

- 1 Capability
- 1 Feature
- 1 Workflow

Do not automatically split a topic into multiple workflows simply because it involves multiple APIs, modules, or code branches. Splitting a topic may be proposed within a proposal *only if* one of the following conditions is met:

- The user explicitly requests a split;
- The source documentation describes multiple distinct, independent functions;
- The two call chains belong to different runtime services, and modifications to one have no impact on the other;
- A single page would exceed a maintainable length, and the boundaries of the split can be clearly articulated using natural language;
- The source materials contain multiple functional objectives that are mutually independent and cannot be classified under a single Feature.

Any decision to split a topic must be confirmed through dialogue prior to execution.

---

## VI. Source Rules

Source information must be embedded inline within the target page, using the following format:

```yaml
sources:
- path: "raw/designs/example.md"
kind: design
hash: ""
title: ""
confidence: medium
notes: ""
```

Requirements:

- Every significant factual statement must be traceable back to a file in the `raw/` directory, an existing Wiki page, or verified code evidence.
- For content pasted directly by the user, use `path: "pasted context"` and specify the actual source within the `notes` field.
- Speculative or unconfirmed information must not be presented as established fact.
- Historical design documents, meeting minutes, and troubleshooting logs must clearly indicate the relevant date and applicable version.
- Any conflicts or discrepancies found within the documentation must be addressed via a formal proposal; one version must not be silently selected over another.
- If a discrepancy exists between the design specifications and the actual code implementation, describe the "Design Intent/Functional Rules" within the Feature section, and describe the "Implementation Discrepancies/Code Evidence" within the Workflow section. ---

## VII. Page Templates

### 7.1 Capability Template

```text
references/feature_template.md
```
### 7.2 Feature Template
```text
references/feature_template.md
```
### 7.3 Workflow Template

```text
references/workflow_template.md
```

### 7.4 Troubleshooting Template

```markdown
---
title: ""
slug: ""
status: active
summary: ""
features: []
sources: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
search_terms: []
---

# <Issue Name>

## Symptoms

## Diagnosis Path

## Logs and Error Keywords

## Possible Causes

## Fix / Recovery

## Related Features

## Related Engineering Workflows

## Source Details

## Search Terms
```

---

## V. Workflows

### Phase 1: Parsing Inputs

1. Expand the `source` scope to obtain the list of materials to be processed.
2. Extract the title, type, version, date, owner, hash, and key terms for each piece of material.
3. Perform a preliminary identification of candidate Capabilities, Features, Workflows, and Troubleshooting topics.
4. Before batch processing, sample 3 to 5 items to verify extraction quality, naming conventions, and page granularity.

Do not jump directly from the raw source material to the final Wiki page.

---

### Phase 2: Extracting Knowledge Signals

For each piece of material, first output "Knowledge Signals" to prevent omissions and facilitate subsequent categorization.

```markdown
## Knowledge Signals

### Capability Signals

### Feature Signals

### Engineering Signals

### Troubleshooting Signals

### Key Terms

### Potential Relationships with Existing Knowledge

### Conflicting or Uncertain Content

### Issues Requiring User Confirmation
```

Notes:

- Knowledge Signals do not constitute the final Wiki page.
- It is not required that every category be present.
- If the signals for a specific category are very weak, do not force the generation of a corresponding page.