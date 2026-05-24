# DevWiki Topic 编写模板

> 适用位置：`wiki/topics/<slug>.md`  
> 定位：Topic 是主题页，合并原能力边界和功能规则，不写代码实现细节。  
> 目标：让 Agent 能先判断主题是否命中，再读取核心功能规则，再按需跳转 Workflow。

```markdown
---
title: "<主题名>"
slug: "<topic-slug>"
kind: topic
status: draft
summary: "<一句话说明该主题解决的问题>"
aliases: []
workflows: []
related_topics: []
troubleshooting: []
sources: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
search_terms: []
---

# <主题名>

<!-- devwiki:section id=card -->
## 导航卡

- 主题定位：
- 适合回答：
  -
- 不适合回答：
  -
- 核心规则摘要：
  -
- 关联 Workflow：
  -
<!-- /devwiki:section -->

<!-- devwiki:section id=core -->
## 核心内容

### 主题摘要

用 3 到 6 条说明：

- 该主题解决什么问题；
- 涉及哪些用户、系统或模块；
- 核心行为是什么；
- 最重要的边界是什么；
- 需要看实现时进入哪个 Workflow。

### 功能边界

#### 范围内

-

#### 范围外

-

### 核心行为

说明功能如何表现，重点写行为，不写实现函数。

### 关键规则

| 规则 | 条件 | 行为 / 结果 | 说明 |
|---|---|---|---|
|  |  |  |  |

### 关键配置与状态

| 配置 / 状态 | 取值 / 条件 | 行为影响 | 说明 |
|---|---|---|---|
|  |  |  |  |

### 关联 Workflow

- `[[<workflow-slug>]]`
<!-- /devwiki:section -->

<!-- devwiki:section id=explain -->
## 详细说明

### 背景与目标

### 典型场景

### 边界场景

| 场景 | 行为 | 说明 |
|---|---|---|
|  |  |  |

### 相关主题

| 相关 Topic | 说明 |
|---|---|
|  |  |

### 可观测性

仅当存在日志、告警、事件、指标、审计或用户可观察状态时填写。

### 验收关注点

- 正常路径：
- 异常路径：
- 边界条件：
- 配置/开关：
- 联动影响：
- 观测/告警：

### 来源说明

- 来源：
- 冲突：
- 不确定：
- 暂未写入的细节：
- 待确认：
<!-- /devwiki:section -->
```

## 质量检查

- Topic 不写代码路径、函数名、handler、调用链。
- 实现相关内容只写关联 Workflow 链接和一句话入口说明，不写代码路径。
- 高频高价值内容进入 card/core。
- 低频背景、验收、冲突、来源进入 explain。
- 不复制 raw 原文。
