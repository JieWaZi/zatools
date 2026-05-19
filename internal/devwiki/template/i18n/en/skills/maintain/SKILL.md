---
name: "devwiki-maintain"
description: "Performs evidence consistency and knowledge health maintenance for DevWiki. Used to discover and fix conflicts, omissions, outdated content, over-compression, missing references, historical mechanisms that no longer apply, query hits on stale content, and other issues between Wiki and raw/source/code."
argument-hint: "<maintenance scope, for example the entire wiki, a specific capability/feature/workflow, a raw file, or a failed query case>"
---

# /devwiki-maintain

## 1. Design Sources and Adoption Principles

```text
raw/source/code = evidence layer, read-only, not directly modified
current wiki page = current understanding layer, maintainable and rewritable
outputs/report = maintenance process report, not a factual entry point
relations/index/glossary = query entry-control layer
```

---

## 2. Prerequisites

Read before starting:

- `references/evidence-grounding.md`
- `references/zatools-qmd.md`
- When writing, reclassifying, or performing destructive operations, also read `references/mutation-safety.md`
- When verifying code, also read `references/code-tracing.md`

---

## 3. Maintenance Goal

The goal of Maintain is not to “clean up formatting”, but to ensure that subsequent query usage relies on the current correct knowledge.

This Skill must discover and fix:

- Wiki is inconsistent with raw/source/code;
- Wiki uses old mechanisms, old conclusions, or old configurations;
- Key design points in raw are missing from Wiki;
- Wiki is over-compressed, causing key rules, boundaries, states, configurations, and exception scenarios to be lost;
- Conflicts exist between multiple Wiki pages;
- Multiple raw/source/code items describe the same conclusion inconsistently;
- Wiki conclusions lack sources;
- Differences exist between the design draft and code implementation but have not been properly placed;
- Outdated pages may still be retrieved by qmd/query and used for answers;
- relations/index/glossary are not synchronized, causing the Agent to enter the wrong entry point.

---

## 4. Core Hard Rule: Do Not Create Active Difference Pages

Maintain may generate difference audits, but must not write difference audits as long-lived active Wiki pages.

It is forbidden to create or retain active pages of this kind:

```text
wiki/sources/*implementation-errata*.md
wiki/features/*errata*.md
wiki/workflows/*errata*.md
wiki/*design-draft-and-implementation-difference*.md
```

Unless the user explicitly requests saving the audit report, the difference audit should only be output in the current response or proposal.

If the user requests saving the report, it may only be written to:

```text
wiki/outputs/<topic>-maintain-report-YYYY-MM-DD.md
```

And it must use:

```yaml
status: report
exclude_from_query: true
visibility: internal
```

The top of the report must write:

```markdown
> This is a maintenance process report, not a functional factual source.
> Current functional rules are based on the corresponding Feature.
> Current implementation paths are based on the corresponding Workflow.
```

Valid conclusions must be merged back into authoritative pages:

| Difference type | Correct placement |
|---|---|
| Capability boundary difference | Capability |
| Functional behavior, rule, configuration, or boundary difference | Feature |
| Code entry point, call chain, or implementation difference | Workflow |
| Symptom, log, or repair-path difference | Troubleshooting |
| Maintenance process comparison table or audit table | outputs/report, not entering active query |


This Skill operates in `discussion_only` mode by default; it is prohibited from modifying any Wiki files unless the user explicitly authorizes write access.

### Write Modes

| Mode              | Description                                       | Allowed Actions                                     |
|-------------------|---------------------------------------------------|-----------------------------------------------------|
| `discussion_only` | Discussion, analysis, and proposal generation only | Cannot create, modify, or delete any files          |
| `dry_run`         | Simulate write operations; preview intended changes | Cannot write changes to disk                        |
| `confirmed_write` | User explicitly authorizes write access           | Can modify Wiki files in accordance with proposals |

Default Mode:

```text
discussion_only
```
---

## 5. Issue Types

Maintain must classify issues according to the following types.

| Type | Meaning | Typical handling |
|---|---|---|
| Missing coverage | raw/source has key facts that Wiki does not cover | Supplement the corresponding authoritative page |
| Over-compression | Wiki is too thin, causing key rules, boundaries, and exceptions to be lost | Rewrite the relevant sections according to the template |
| Unsupported conclusion | Wiki states a conclusion, but no raw/source/code support can be found | Mark as pending confirmation; do not fabricate a source |
| Evidence conflict | Multiple sources or Wiki pages describe inconsistently | Output a conflict table; modify only after confirmation |
| Historical invalidity | Wiki content once applied, but no longer applies to the current version/implementation | Mark historical scope and update the current conclusion |
| Implementation divergence | Design document and code implementation are inconsistent | Feature writes the functional conclusion, Workflow writes the implementation difference |
| Difference report mistakenly landed | maintain wrote errata/report as active Wiki | Move to outputs/report or delete, and merge conclusions back into authoritative pages |
| Relationship error | relations/index/glossary points incorrectly or is missing | Fix relationships and entry points |
| Query pollution | Old pages or report pages are hit by query and mislead answers | Downgrade, exclude, change entry points, and update the index |
| Template non-compliance | Title, frontmatter, source, or status fields do not conform to the specification | Low-risk direct fix |

---

## 7. Maintenance Levels

### 7.1 Directly Fixable

The following conditions may be fixed directly:

- Markdown section titles are inconsistent;
- frontmatter fields are missing;
- source path format is incorrect;
- index/relations/glossary updates are missing;
- obvious broken links;
- search_terms are missing;
- page status fields are missing;
- log is not recorded;
- A maintenance report was mistakenly placed as active and is not referenced by other pages; it may be moved into outputs and given `exclude_from_query: true`;
- Low-risk wording normalization.

A maintenance report is still required after direct fixes.

### 7.2 Requires Proposal Before Fixing

The following situations must first output a proposal:

- Key rules in raw are missing and Feature needs to be supplemented;
- Wiki content is over-compressed and needs a large rewrite;
- Old mechanisms need to be marked as historical;
- Conclusions conflict between multiple pages;
- Feature / Workflow ownership needs adjustment;
- Pages need to be merged, split, or renamed;
- An active errata page needs to be decomposed and merged back into authoritative pages;
- The query entry point needs to be changed to avoid continuing to hit outdated content.

### 7.3 Requires Human Confirmation

The following situations must not be applied automatically:

- Multiple raw/source/code items conflict with each other on the same rule;
- Design and code implementation are inconsistent, and it is unclear which should be authoritative;
- Deleting pages;
- Deleting business rules;
- Changing active to deprecated;
- Changing the main capability/feature boundary;
- Conclusions affecting external customers, upgrade compatibility, alarms, or interface behavior;
- Writing something as a definite fact without evidence.

---

## 8. Workflow

### Phase 1: Determine Maintenance Scope

First clarify the maintenance scope: the entire Wiki, a specific Capability / Feature / Workflow / Troubleshooting, a raw/source file, a failed query answer, recently changed pages, or an incorrectly created errata/report page.

If the user does not specify a scope, check in the following priority order by default:

1. Pages related to the issue the user just pointed out;
2. active pages;
3. pages frequently hit by query;
4. pages used as entry points in relations/index/glossary;
5. recently updated pages.

### Phase 2: Read Context

Do not modify after reading only one page. At minimum, read: the target Wiki page, frontmatter sources, the corresponding raw/source, related Capability / Feature / Workflow / Troubleshooting, `relations.yml`, `index.md`, and `glossary.md`. When implementation differences are involved, verify code; when errata/report is involved, read the raw/code it references and determine where the conclusions should be placed.

### Phase 3: Evidence Audit

Check whether each key conclusion can be traced to a source; whether the source exists, is outdated, or conflicts; whether the page writes design intent as current implementation; whether it writes historical mechanisms as current mechanisms; and whether it treats a maintain report as a factual source.

```markdown
## Evidence Audit

| Wiki conclusion | Current source | Evidence status | Issue type | Recommendation |
|---|---|---|---|---|
|  |  | Supported / Missing evidence / Conflict / Outdated / Incorrectly references report |  |  |
```

### Phase 4: Coverage Audit

Extract key knowledge signals from raw/source/code, then check whether Wiki covers them.

| Page type | Key checks |
|---|---|
| Capability | Capability goal, capability boundary, covered Feature, capability relationship, capability-level constraints |
| Feature | Functional goal, core behavior, key rules, key concepts, important configuration, boundary exceptions, acceptance concerns |
| Workflow | Code entry point, call chain, key logic, state reads/writes, configuration handling, implementation differences, test references, change impact |
| Troubleshooting | Symptoms, logs, diagnostic path, possible causes, repair/recovery, related functions |

### Phase 5: Difference Placement Audit

If “design draft and implementation difference”, “errata”, or “implementation difference” type content is found, the placement audit must be performed.

```markdown
## Difference Placement Audit

| Difference item | Current page | Correct placement | Already merged | Follow-up action |
|---|---|---|---|---|
|  | errata/report | Feature / Workflow / Capability / Troubleshooting / outputs | Yes / No |  |
```

Handling rules: Difference items must not exist only in errata/report for the long term; functional-level differences are merged into Feature; implementation-level differences are merged into Workflow; troubleshooting-level differences are merged into Troubleshooting; capability-boundary differences are merged into Capability; after merging, errata/report must be moved into outputs or deleted, and excluded from query.

### Phase 6: Conflict and Historical Invalidity Audit

Check conflicts between pages, conflicts between raw/source/code, historical mechanisms still being written as current mechanisms, and conflicts between errata/report and authoritative pages.

| Situation | Handling |
|---|---|
| New evidence clearly supersedes old evidence | Update the current conclusion and move old content into “Historical Notes” |
| It is uncertain which of old/new evidence is valid | Mark the conflict and wait for confirmation |
| The old mechanism still has version value | Keep it, but mark the applicable version/time/condition |
| The old mechanism no longer applies at all | Recommend deprecating or moving it into a historical page in the proposal |
| Code implementation differs from design | Feature writes the functional conclusion; Workflow writes the current implementation difference |
| errata/report is hit by query | Move it into outputs, set `exclude_from_query: true`, and merge conclusions back into authoritative pages |

### Phase 7: Query Pollution Check

Focus on preventing query from continuing to answer with old content or maintenance reports. Check whether outdated pages and errata/report are active, whether report lacks `exclude_from_query: true`, whether index/relations/glossary still point to old pages or reports, and whether qmd search ranks old pages/reports ahead of authoritative pages.

### Phase 8: Output Maintain Proposal

High-risk changes must first output a proposal.

```markdown
# Maintain Proposal

## Maintenance Scope

## Discovered Issues

| Page | Issue type | Severity | Description | Recommended action |
|---|---|---|---|---|

## Difference Placement Plan

| Difference item | Current location | Target page | Modification method | Needs confirmation |
|---|---|---|---|---|

## Evidence and Sources

## Query Pollution Risk

## Questions Requiring Confirmation

## Content Not Recommended for Automatic Modification
```

### Phase 9: Apply After Confirmation

After user confirmation, apply high-risk changes: update authoritative pages, merge valid conclusions from errata/report back into authoritative pages; write the maintenance report into `wiki/outputs/` and set `exclude_from_query: true`; update `relations.yml`, `index.md`, `glossary.md`, and `log.md`; finally execute or prompt execution of:

```bash
zatools qmd update
zatools qmd status
```

It is forbidden to write the maintenance report as an active Wiki page.

### Phase 10: Post-Maintenance Verification

Search the original question keywords again; check whether qmd prioritizes authoritative pages; whether errata/report will not become a query entry point; whether old pages are clearly marked historical/deprecated; whether relations/index/glossary point to the current mechanism; and use 2 to 5 seed questions to test whether query will still answer with old content.

---

## 10. Prohibited Actions

### 10.1 General Prohibitions

- Do not modify raw/source original materials.
- Do not modify Wiki without reading the source.
- Do not rewrite conclusions after reading only a single page of context.
- Do not write uncertain content as a definite fact.
- Do not handle conflicts silently.
- Do not skip proposal and directly make high-risk changes.
- Do not delete pages unless the user explicitly confirms.
- Do not finish maintaining Wiki without updating index/relations/glossary.
- Do not finish maintaining Wiki without updating the qmd index.

### 10.2 Difference Report Prohibitions

- Do not write differences between the design draft and implementation as long-lived active Wiki pages.
- Do not place errata/report under `wiki/sources/` as factual sources.
- Do not let errata/report become the main query entry point.
- Do not cite maintain report in Feature/Workflow sources to support facts.
- Do not only write a difference report without merging back into authoritative pages.
- Do not point index/relations/glossary to errata/report as the current conclusion.
- Do not describe a report as “this page + code is authoritative”; the current conclusion should be based on Feature/Workflow.

### 10.3 Evidence Prohibitions

- Do not fabricate citation/source.
- Do not force a source onto an unsupported fact.
- Do not delete unsupported facts to hide the issue; flag them first.
- Do not write plans in design documents as current implementation.
- Do not write current code behavior as product design unless explicitly marked as “implementation status”.

### 10.4 Conflict Handling Prohibitions

- Do not automatically choose the side that seems more reasonable.
- Do not mix rules from multiple versions into one rule.
- Do not leave old-version rules in an active page without marking their applicable scope.
- Do not change active to deprecated without confirmation.
- Do not merge or split main pages without confirmation.

### 10.5 Query Pollution Prevention Prohibitions

- Do not only modify the body without modifying summary/status/search_terms.
- Do not only modify pages without modifying index/relations/glossary.
- Do not keep multiple conflicting active entry points.
- Do not let old pages continue as the main entry point.
- Do not let glossary point old terms preferentially to old mechanisms.
- Do not let outputs/report participate in normal query.

---