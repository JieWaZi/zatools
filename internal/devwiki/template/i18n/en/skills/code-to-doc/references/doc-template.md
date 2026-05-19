# Workflow Page Template

Default target path: `wiki/workflows/<feature-slug>.md`

Use this structure when drafting or refreshing an engineering-location page from code:

```markdown
# <Workflow Title>

## Summary

## Entry Points

## Call Chain

## Key Logic

## Data and State

## Code References

## Test References

## Change Impact

## Source Notes
```

Frontmatter should at least cover:

```yaml
---
title: ""
slug: ""
status: active
summary: ""
features: []
sources: []
code_refs: []
api_entries: []
test_refs: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
---
```

If a feature design page must also be updated, keep it limited to behavior, parameters, interactions, and functional flow. Do not put code references there.

Every conclusion must be backed by code or raw/wiki evidence.
