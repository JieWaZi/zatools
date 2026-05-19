# DevWiki — 运行时 Schema

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，定义目录结构、页面 schema、链接规范、检索顺序与 workflow 约束。

## 目录结构

```text
./
├── README.md
├── {{RUNTIME_FILE}}
├── config/
│   ├── project.yaml
│   └── search.yaml
├── raw/
│   ├── requirements/
│   ├── designs/
│   ├── features/
│   └── tests/
└── wiki/
    ├── index.md
    ├── glossary.md
    ├── relations.yml
    ├── log.md
    ├── capabilities/
    ├── features/
    ├── workflows/
    ├── troubleshooting/
    └── outputs/
```

本目录就是 DevWiki 文档库根目录。代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录。Agent 在代码库内使用 `devwiki-query` 或 `devwiki-code-to-doc` 前，必须先阅读本文件；查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源，生成的新 Wiki 文件也必须写回本目录。

关键路径：

- `wiki/capabilities/{slug}.md`
- `wiki/features/{slug}.md`
- `wiki/workflows/{slug}.md`
- `wiki/troubleshooting/{slug}.md`
- `wiki/log.md`
- `wiki/relations.yml`
- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

## 页面类型

| 目录 | 文件名 | 职责 |
|------|--------|------|
| `wiki/capabilities/` | `{slug}.md` | 业务能力或系统能力中心页 |
| `wiki/features/` | `{slug}.md` | 具体功能设计页，说明参数、联动、边界和功能流转 |
| `wiki/workflows/` | `{slug}.md` | 工程定位页，合并流程与模块定位，说明入口、调用链、关键逻辑、代码引用和修改影响 |
| `wiki/troubleshooting/` | `{slug}.md` | 排障现象、诊断路径、证据、修复建议和适用版本 |

## 页面模板

### wiki/capabilities/{slug}.md

```yaml
---
title: ""
slug: ""
status: active
summary: ""
features: []
related_capabilities: []
sources: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
---
```

正文：`Overview`、`Business Value`、`Capability Boundary`、`Covered Features`、`Related Capabilities`、`Source Notes`。

Capability 只写能力边界、业务价值、覆盖功能和能力协作，不写代码引用、接口、函数或测试入口。

### wiki/features/{slug}.md

```yaml
---
title: ""
slug: ""
status: active
summary: ""
capabilities: []
workflow: ""
related_features: []
sources: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
---
```

正文：`Summary`、`User Scenarios`、`Functional Design`、`Parameters`、`Behavior Flow`、`Interactions`、`Constraints`、`Acceptance Notes`、`Engineering Entry`、`Source Notes`。

Feature 只写功能设计、参数取值、联动、边界和功能流转，不写代码路径、函数名、`code_refs`、`api_entries` 或 `test_refs`。需要指向实现时，只写 “实现定位见 [[workflow-slug]]”。

### wiki/workflows/{slug}.md

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

正文：`Summary`、`Entry Points`、`Call Chain`、`Key Logic`、`Data and State`、`Code References`、`Test References`、`Change Impact`、`Source Notes`。

Workflow 是面向编程的工程定位页。默认每个 feature 最多一个 workflow；只有用户确认，或两条调用链属于不同运行时服务且修改影响完全不同，才拆多个 workflow。

### wiki/troubleshooting/{slug}.md

```yaml
---
title: ""
slug: ""
status: active
summary: ""
features: []
workflows: []
sources: []
symptoms: []
visibility: internal
confidence: medium
last_verified_at: YYYY-MM-DD
---
```

正文：`Symptoms`、`Diagnosis`、`Evidence`、`Fix / Mitigation`、`Related Features`、`Related Workflows`、`Source Notes`。

## 来源和代码引用

来源信息内联到页面：

```yaml
sources:
  - path: "raw/designs/example.md"
    kind: design
    hash: ""
    title: ""
    confidence: medium
    notes: ""
```

`code_refs` 只写在 `wiki/workflows/` 或 `wiki/troubleshooting/`：

```yaml
code_refs:
  - path: "services/user/service.ts"
    kind: file
    symbol: ""
    note: "用户 CRUD 主服务入口"
    confidence: 0.9
```

## 检索顺序

- 能力问题：`capabilities → features`
- 功能问题：`features → capabilities`
- 代码问题：`workflows → features → rg`
- 排障问题：`troubleshooting → workflows → features`

检索通道按成本升档：本地搜索、`zatools qmd search`、`zatools qmd query`。`qmd` 只是召回工具，不是真相源。

## Workflow 约束

- `raw/` 只读
- 页面中的 `sources` 必须记录真实路径和 hash
- capability 和 feature 禁止写代码引用
- workflow 负责工程定位和修改影响
- 待确认问题优先通过对话和用户对齐
- 中高风险写入必须先给 proposal，再落盘
- 多个候选归属、拆分边界不清、页面合并/重命名、代码证据不足时必须先问用户

## 操作说明

- 使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库并安装 skills
- 使用 `zatools devwiki link` 将已有 DevWiki 文档库关联到代码库
- 使用 `devwiki-project-router` 作为项目知识任务的默认总入口
- 使用 `devwiki-ingest` 吸收 raw 文档并生成或更新三层 Wiki 页面、术语和关系
- 使用 `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失、关系错误和 query 污染
- 使用 `devwiki-query` 查询 Wiki、raw、代码线索、设计意图和排障知识
- 使用 `devwiki-code-to-doc` 从代码、接口、配置项、日志或路由反向生成或更新 workflow 页面

这份 schema 应保持稳定，只在 DevWiki 的数据模型或工作流发生真实变化时修改。
