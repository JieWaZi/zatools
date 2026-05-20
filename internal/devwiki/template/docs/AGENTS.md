# DevWiki — 运行时 Schema

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，定义目录结构、页面 schema、链接规范、检索顺序与 workflow 约束。

---

## 目录结构

```text
./
├── README.md                    ← 项目说明、使用方式与工作流入口
├── {{RUNTIME_FILE}}             ← 当前 agent 对应的运行时规则副本
├── config/
│   ├── project.yaml             ← 当前项目默认配置，如项目 slug、代码仓列表、语言与 agent
│   └── search.yaml              ← qmd collection 配置
├── raw/
│   ├── requirements/            ← 原始需求文档
│   ├── designs/                 ← 原始设计文档
│   ├── features/                ← 原始功能说明
│   └── tests/                   ← 测试方案与测试记录
└── wiki/
    ├── index.md                 ← Wiki 总目录与导航入口
    ├── glossary.md              ← 术语表
    ├── log.md                   ← 追加式操作日志
    ├── capabilities/            ← 系统能力
    ├── features/                ← 具体功能设计
    ├── workflows/               ← 工程定位：调用链、关键逻辑与代码引用
    ├── troubleshooting/         ← 排障知识
    └── outputs/                 ← ingest / maintain / query / code-to-doc / qmd-sync 报告
```

本目录就是 DevWiki 文档库根目录。代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录。Agent 在代码库内使用 `devwiki-query` 或 `devwiki-code-to-doc` 前，必须先阅读本文件；查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源，生成的新 Wiki 文件也必须写回本目录。

`raw/` 是事实来源层，`wiki/` 是结构化知识层。`config/search.yaml` 只保存检索配置，不替代事实内容。

---

## 页面类型

DevWiki 的人工维护知识页围绕能力、功能、工程定位和排障知识组织。代码不是独立目录，而是通过 `wiki/workflows/{slug}.md` 中的 `code_refs`、`api_entries`、`test_refs` 被结构化引用。

| 目录 | 文件名 | 职责 |
|------|--------|------|
| `wiki/capabilities/` | `{slug}.md` | 某个业务能力或系统能力的中心页 |
| `wiki/features/` | `{slug}.md` | 某个具体功能的设计页，负责参数、联动、边界和功能流转 |
| `wiki/workflows/` | `{slug}.md` | 面向编程的工程定位页，负责入口、调用链、关键逻辑、代码引用和修改影响 |
| `wiki/troubleshooting/` | `{slug}.md` | 排障现象、诊断路径、证据、修复建议和适用版本 |

生成目录不是人工知识实体：

- `wiki/outputs/`

关键显式路径：

- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `wiki/capabilities/`
- `wiki/features/`
- `wiki/workflows/`
- `wiki/troubleshooting/`
- `wiki/outputs/`
- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

---

## 链接语法

所有内部链接使用 Obsidian wikilink：

```markdown
[[slug]]
[[user-management]]
[[permission-assignment]]
[[user-group-management]]
```

命名规范：

- 全小写
- 使用连字符分隔
- 不带空格
- slug 一旦发布应尽量保持稳定

`raw` 文件通过页面内联 `sources` 追踪。外部代码文件不使用 wikilink，也不写入 capability / feature 的 `sources`，只通过 workflow 或 troubleshooting 的 `code_refs`、`api_entries`、`test_refs` 追踪。

---

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| capability/A 写入 `features: [[feature-F]]` | feature/F 的 `capabilities` 追加 A |
| capability/A 写入 `related_capabilities: [[capability-B]]` | capability/B 的 `related_capabilities` 追加 A |
| feature/F 写入 `workflow: [[workflow-W]]` | workflow/W 的 `features` 追加 F |
| feature/F 写入 `related_features: [[feature-G]]` | feature/G 的 `related_features` 追加 F |
| troubleshooting/T 写入 `features: [[feature-F]]` | feature/F 的排障入口摘要追加 T |

以下外部对象没有反向页面要求：

- `raw/*` 属于外部来源，只通过页面内联 `sources.path` 与 `sources.hash` 追踪
- 代码文件和 symbol 属于外部对象，只通过 workflow 页的 `code_refs` 追踪
- 接口入口通过 workflow 页的 `api_entries` 记录，不要求 wiki 反向链接
- 测试入口通过 workflow 页的 `test_refs` 记录，不要求 wiki 反向链接

---

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

正文建议分区：

- `## Overview`
- `## Business Value`
- `## Capability Boundary`
- `## Covered Features`
- `## Related Capabilities`
- `## Source Notes`

`capabilities` 页面只回答：

- 系统有哪些能力
- 能力解决什么业务问题
- 能力边界和作用效果是什么
- 它和哪些 feature 相关
- 它和其他 capability 如何协作

不要在 capability 页里展开：

- 具体实现
- 调用链
- `code_refs`
- `api_entries`
- `test_refs`

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

正文建议分区：

- `## Summary`
- `## User Scenarios`
- `## Functional Design`
- `## Parameters`
- `## Behavior Flow`
- `## Interactions`
- `## Constraints`
- `## Acceptance Notes`
- `## Engineering Entry`
- `## Source Notes`

`features` 页面负责：

- 说明具体功能是什么
- 说明参数、字段、取值范围和默认值
- 说明功能开关、配置和其他功能之间的联动
- 说明正常流程、异常语义和边界条件
- 说明设计思想和功能流转

`features` 页面不得写：

- 代码文件路径
- 函数名
- 模块内部实现
- `code_refs`
- `api_entries`
- `test_refs`

如需指向实现，只在 `Engineering Entry` 中写一句：实现定位见 `[[workflow-slug]]`。

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

正文建议分区：

- `## Summary`
- `## Entry Points`
- `## Call Chain`
- `## Key Logic`
- `## Data and State`
- `## Code References`
- `## Test References`
- `## Change Impact`
- `## Source Notes`

`workflows` 是面向编程的工程定位页，合并原流程页和模块页。它负责帮助后续 agent 快速定位代码，不负责复述完整业务背景。

默认粒度：每个 feature 最多一个 workflow。只有用户明确确认，或者两条调用链属于不同运行时服务且修改影响完全不同，才拆多个 workflow。

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

正文建议分区：

- `## Symptoms`
- `## Diagnosis`
- `## Evidence`
- `## Fix / Mitigation`
- `## Related Features`
- `## Related Workflows`
- `## Source Notes`

---

## Source 结构

来源信息内联写入对应页面，不再单独生成来源页面：

```yaml
sources:
  - path: "raw/designs/example.md"
    kind: design
    hash: ""
    title: ""
    confidence: medium
    notes: ""
```

使用规则：

- 每个关键事实必须能回到 `raw/`、已有 Wiki 页面或已核对代码证据
- 用户粘贴内容使用 `path: "pasted context"` 并在 `notes` 中说明
- 不确定内容必须标记 `confidence: low` 或在写入前通过对话确认
- capability / feature 的 `sources` 不写代码文件路径、函数名、handler、调用链或 `kind: code`；代码证据统一写入 workflow 或 troubleshooting 的 `code_refs`

---

## Code Ref 结构

`code_refs` 只写在 `wiki/workflows/` 或 `wiki/troubleshooting/`：

```yaml
code_refs:
  - path: "services/user/service.ts"
    kind: file
    note: "用户资料读写和状态同步的主服务实现。"
    confidence: high
    symbols:
      UserService#class: "用户资料服务主类。"
      UserService#updateProfile#method: "状态写入入口，修改时需要同步检查缓存刷新。"
```

使用规则：

- `code_refs` 以代码文件 `path` 为唯一粒度
- 同一个 `path` 在同一页面中只能出现一条 `code_refs`
- 顶层 `note` 只写文件级职责，不写每个方法的说明
- `symbols` 是关键入口索引，不是文件内方法清单
- `symbols` 最多 4 个，只列主入口、关键状态读写、配置处理、外发、副作用、恢复或排障入口
- `symbols` 使用 `<symbol>#<kind>: "<短说明>"` 格式，value 只写入口职责、风险点或排障关键说明
- 不得为了完整性列出文件内所有方法
- 没核对过的路径和 symbol 不得虚构
- capability 和 feature 页面禁止写代码引用，也禁止在 `sources` 中记录代码文件路径或 `kind: code`

---

## 检索顺序

按用户意图选择起点：

- 能力问题：`capabilities → features`
- 功能问题：`features → capabilities`
- 代码问题：`workflows → features → rg`
- 排障问题：`troubleshooting → workflows → features`

检索通道按「成本 / 速度由低到高」阶梯升档，详见 `references/zatools-qmd.md`：

1. 本地 `grep` / 文件搜索
2. `zatools qmd search`
3. `zatools qmd query`

任一档命中 top-K 且置信足够即停；只有当问题落到实现现实、代码归属，或文档证据仍不足时，才对 top-K 代码候选做一次定向本地排查。

---

## Workflow 约束

### 事实约束

- `raw/` 是只读原始资料
- 页面中的 `sources` 必须记录真实 `path` 与 `hash`
- facts 与 inference 必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据

### 角色约束

- capability 页是业务能力与系统能力的中心页
- feature 页是功能设计与行为说明页
- workflow 页是工程定位、代码引用和修改影响页
- 项目知识任务先由 `devwiki-project-router` 判断意图、身份、证据需求和检索边界，再路由到 `devwiki-ingest`、`devwiki-maintain`、`devwiki-query`、`devwiki-code-to-doc` 或 `devwiki-qmd-sync`

### 确认策略

- 待确认问题优先通过对话和用户对齐
- 中高风险写入必须先给 proposal，再落盘
- 多个候选归属、拆分边界不清、页面合并/重命名、代码证据不足时必须先问用户

---

## 操作说明

- 使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库并安装 skills
- 使用 `zatools devwiki link` 将已有 DevWiki 文档库关联到代码库，并在代码库中写入 AGENTS/CLAUDE 关联规则
- 如需下载 qmd models，初始化完成后可在 DevWiki 工作区内手动执行 `zatools qmd download --root .`
- 对已有工作区补做或修复 qmd collection 注册、索引刷新与状态检查时，使用 `devwiki-qmd-sync`
- 使用 `devwiki-project-router` 作为项目知识任务的默认总入口
- 使用 `devwiki-ingest` 吸收 raw 文档并生成或更新三层 Wiki 页面、术语和关系
- 使用 `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失、关系错误和 query 污染
- 使用 `devwiki-query` 查询 Wiki、raw、代码线索、设计意图和排障知识
- 使用 `devwiki-code-to-doc` 从代码、接口、配置项、日志或路由反向生成或更新 workflow 页面

这份 schema 应保持稳定，只在 DevWiki 的数据模型或工作流发生真实变化时修改。
