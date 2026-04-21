# Feature Page Template

Default target path: `wiki/features/<feature-slug>.md`

Use this structure when drafting or refreshing a feature page:

```markdown
# <Feature Title>

## Summary

## Supported Capabilities

## Business Flow

## Constraints

## Entry Points

## Code Clues

## Test Entry Points

## Source Trace

## Open Questions
```

Frontmatter should at least cover:

```yaml
---
title: ""
slug: ""
status: active
summary: ""
capabilities: []
sources: []
api_entries: []
code_refs: []
test_refs: []
related_features: []
owner: ""
tags: []
confidence: 0.8
last_verified_at: YYYY-MM-DD
verification_status: pending
---
```

Every conclusion must be backed by code or raw/wiki evidence.
