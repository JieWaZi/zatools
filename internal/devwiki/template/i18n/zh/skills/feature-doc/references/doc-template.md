# Feature 页面模板

目标文件默认写入：`wiki/features/<feature-slug>.md`

整理或刷新 feature 页面时，默认使用以下结构：

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

Frontmatter 至少应包含：

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

所有结论都必须能回溯到代码证据或 raw/wiki 证据。
