# DevWiki — 运行时 Schema

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，定义目录结构、页面 schema、链接规范、检索顺序与 workflow 约束。

---

## 目录结构

```text
{{WORKSPACE_DIR}}/
├── README.md                    ← 项目说明、使用方式与工作流入口
├── {{RUNTIME_FILE}}             ← 当前 agent 对应的运行时规则副本
├── config/
│   ├── project.yaml             ← 当前项目默认配置，如项目 slug、代码仓列表、语言与 agent
│   └── search.yaml              ← qmd collection 配置
├── raw/
│   ├── requirements/            ← 原始需求文档
│   ├── designs/                 ← 原始设计文档
│   ├── features/                ← 原始特性说明
│   ├── code-summaries/          ← 原始代码总结
│   ├── postmortems/             ← 缺陷复盘与事故复盘
│   ├── api/                     ← 接口文档与接口说明
│   └── tests/                   ← 测试方案与测试记录
└── wiki/
    ├── documents/
    │   ├── requirements/        ← 需求文档镜像页
    │   ├── designs/             ← 设计文档镜像页
    │   ├── features/            ← 特性说明镜像页
    │   ├── code-summaries/      ← 代码总结镜像页
    │   ├── postmortems/         ← 复盘文档镜像页
    │   ├── api/                 ← 接口文档镜像页
    │   └── tests/               ← 测试文档镜像页
    ├── capabilities/            ← 业务能力与系统能力中心页
    ├── changes/                 ← 用于判断 new / modify / unclear 的变更页
    ├── index.md                 ← Wiki 总目录与导航入口
    └── log.md                   ← 追加式操作日志
```

当前项目根的桥接运行时文件会要求 agent 在处理 DevWiki 任务前先阅读 `./{{WORKSPACE_DIR}}/{{RUNTIME_FILE}}`。项目级 skills 与 `.zatools-lock.json` 位于当前项目根，不在本目录内。

`raw/` 是事实来源层，`wiki/` 是结构化知识层。`config/search.yaml` 只保存检索配置，不替代事实内容。

---

## 页面类型

DevWiki 在 v1 中只有三类一等 wiki 页面。代码不是一等页面，而是通过 `code_refs` 被结构化引用。

| 目录 | 文件名 | 职责 |
|------|--------|------|
| `wiki/documents/<doc-type>/` | `{slug}.md` | 一份 raw 文档的结构化镜像页 |
| `wiki/capabilities/` | `{slug}.md` | 某个业务能力或系统能力的中心页 |
| `wiki/changes/` | `{slug}.md` | 用于判断 `new`、`modify`、`unclear` 的变更页 |

支持的 document 子类型：

- `requirements`
- `designs`
- `features`
- `code-summaries`
- `postmortems`
- `api`
- `tests`

关键显式路径：

- `wiki/documents/requirements/`
- `wiki/documents/designs/`
- `wiki/documents/features/`
- `wiki/capabilities/`
- `wiki/changes/`

---

## 链接语法

所有内部链接使用 Obsidian wikilink：

```markdown
[[slug]]
[[user-management]]
[[permission-assignment]]
[[user-group-permission-refactor]]
```

命名规范：

- 全小写
- 使用连字符分隔
- 不带空格
- slug 一旦发布应尽量保持稳定

`raw` 文件与外部代码文件不使用 wikilink，而是通过 `source_path`、`source_hash`、`api_refs`、`code_refs` 追踪。

---

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| document/D 写入 `capabilities: [[capability-C]]` | capability/C 的 `documents` 追加 D |
| document/D 写入 `changes: [[change-X]]` | change/X 的 `source_documents` 追加 D |
| capability/A 写入 `child_capabilities: [[capability-B]]` | capability/B 的 `parent_capabilities` 追加 A |
| capability/A 写入 `related_capabilities: [[capability-B]]` | capability/B 的 `related_capabilities` 追加 A |
| change/X 写入 `capabilities: [[capability-C]]` | capability/C 的 `changes` 追加 X |
| document/D 写入 `related_documents: [[document-E]]` | document/E 的 `related_documents` 追加 D |

以下外部对象没有反向页面要求：

- `raw/*` 属于外部来源，只通过 `source_path` 与 `source_hash` 追踪
- 代码文件和 symbol 属于外部对象，只通过 `code_refs` 追踪
- 接口可通过 `api_refs` 或 URL 记录，不要求 wiki 反向链接

---

## 页面模板

### wiki/documents/{doc-type}/{slug}.md

```yaml
---
title: ""
slug: ""
doc_type: requirement
source_path: "raw/requirements/example.md"
source_hash: ""
status: active
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
authors: []
tags: []
capabilities: []
changes: []
related_documents: []
code_refs: []
summary_confidence: high
---
```

正文建议分区：

- `## Summary`
- `## Intent`
- `## Scope`
- `## Key Decisions`
- `## Linked Capabilities`
- `## Code Clues`
- `## Risks / Open Questions`
- `## Source Trace`

### wiki/capabilities/{slug}.md

```yaml
---
title: ""
slug: ""
type: business
status: active
summary: ""
aliases: []
parent_capabilities: []
child_capabilities: []
related_capabilities: []
owner: ""
tags: []
documents: []
changes: []
code_refs: []
api_refs: []
confidence: 0.8
last_verified_at: YYYY-MM-DD
verification_status: pending
---
```

正文建议分区：

- `## Overview`
- `## Classification`
- `## Behaviors`
- `## Design History`
- `## Related Documents`
- `## Code Map`
- `## Known Constraints`
- `## Open Questions`
- `## Notes`

`type` 只能是：

- `business`
- `system`

### wiki/changes/{slug}.md

```yaml
---
title: ""
slug: ""
change_type: feature
status: proposed
change_classification: modify
source_documents: []
capabilities: []
related_changes: []
tags: []
created_at: YYYY-MM-DD
updated_at: YYYY-MM-DD
impact_summary: ""
planned_code_refs: []
actual_code_refs: []
decision_confidence: 0.7
verification_status: pending
---
```

正文建议分区：

- `## Why This Change`
- `## Capability Impact`
- `## Proposed Design Delta`
- `## Code Impact`
- `## Actual Implementation Notes`
- `## Risks`
- `## Follow-ups`
- `## Verification`

`change_classification` 只能是：

- `new`
- `modify`
- `unclear`

---

## Code Ref 结构

推荐的 `code_refs` 结构：

```yaml
code_refs:
  - path: "services/user/service.ts"
    kind: file
    symbol: ""
    note: "用户 CRUD 主服务"
    confidence: 0.9
  - path: "services/user/group.ts"
    kind: function
    symbol: "createUserGroup"
    note: "用户组创建逻辑"
    confidence: 0.8
```

使用规则：

- 如果整个文件都相关，直接存文件列表
- 如果文件很大但只涉及局部逻辑，使用 `path + symbol`
- 接口场景允许使用 `api_refs` URL，前提是能帮助 agent 快速跳转
- 没核对过的路径和 symbol 不得虚构

---

## 检索顺序

DevWiki 的检索建立在三个 collection 之上：

- `wiki`
- `raw`
- `code`

建议顺序：

1. 先查 `wiki`
2. 再查 `raw`
3. 再查 `code`
4. 最后对 top-K 代码候选做一次定向本地排查

`qmd` 只负责加速召回，不是事实存储层，不能覆盖 `raw/`、`wiki/` 与真实代码内容。

---

## Workflow 约束

### 事实约束

- `raw/` 是只读原始资料
- 每个 document 镜像页都必须记录真实的 `source_path` 与 `source_hash`
- documents、capabilities、changes 必须建立在原始文档证据或已核对代码证据之上
- 事实与推断必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据

### 变更风险策略

- 低风险：
  - 新建 document 镜像页
  - 追加确定性日志
- 中风险：
  - 将文档挂到已有 capability
  - 追加辅助性 `code_refs`
- 高风险：
  - 新建 capability
  - 合并或拆分 capability
  - 新建或重分类 change
  - 替换主 `code_refs`
  - 修改 `change_classification`

任何中高风险写入都必须先显式确认。

### 搜索纪律

- 禁止没有边界地盲搜
- 若经过几轮检索后仍然低置信，应停止继续扩散，并向用户提 1 到 3 个具体问题
- 优先询问接口 URL、关键文件、关键函数、路由或 capability 名称

### 编辑纪律

- 写正向关系时，同步写反向关系
- 不要静默删除坏链；应报告问题，或分流到 `/devwiki-refresh`、`/devwiki-check`
- 只需改一个字段或章节时，不要整页重写

### 范围纪律

- capability 页面是业务能力与系统能力的共同中心页
- `business` 与 `system` 用字段区分，不用目录区分
- change 页面是判断某需求属于 `new`、`modify`、`unclear` 的主载体

---

## 操作说明

- 使用 `zatools devwiki init` 初始化 DevWiki 工作区、安装 skills、预热 qmd 模型，并生成项目根桥接运行时文件
- 使用 `devwiki-setup` 解释初始化约束、安装范围与运行时使用方式
- 对已有工作区补做或修复 qmd collection 注册、索引刷新与状态检查时，使用 `devwiki-qmd-sync`
- 使用 `devwiki-init` 基于现有 `raw/` 建立第一版知识骨架
- 使用 `devwiki-ingest` 吸收增量 raw 文档
- 在真正动手开发前使用 `devwiki-scope`
- 当 wiki 与 raw 或代码漂移时使用 `devwiki-refresh`
- 使用 `devwiki-check` 做确定性健康检查

这份 schema 应保持稳定，只在 DevWiki 的数据模型或工作流发生真实变化时修改。
