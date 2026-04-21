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
│   ├── features/                ← 原始功能说明
│   └── tests/                   ← 测试方案与测试记录
└── wiki/
    ├── capabilities/            ← 业务能力地图
    ├── features/                ← 功能说明页
    ├── outputs/                 ← 问答沉淀、临时汇总、导出结果
    ├── graph/                   ← 自动生成的关系图与缺口分析
    ├── index.md                 ← Wiki 总目录与导航入口
    └── log.md                   ← 追加式操作日志
```

当前项目根的桥接运行时文件会要求 agent 在处理 DevWiki 任务前先阅读 `./{{WORKSPACE_DIR}}/{{RUNTIME_FILE}}`。项目级 skills 与 `.zatools-lock.json` 位于当前项目根，不在本目录内。

`raw/` 是事实来源层，`wiki/` 是结构化知识层。`config/search.yaml` 只保存检索配置，不替代事实内容。

---

## 页面类型

新的 DevWiki 只保留两类人工维护的一等知识页。代码不是一等页面，而是通过 `code_refs` 被 feature 页结构化引用。

| 目录 | 文件名 | 职责 |
|------|--------|------|
| `wiki/capabilities/` | `{slug}.md` | 某个业务能力或系统能力的中心页 |
| `wiki/features/` | `{slug}.md` | 某个具体功能的说明页，负责串起流程、约束、入口和代码线索 |

生成目录不是人工知识实体：

- `wiki/outputs/`
- `wiki/graph/`

关键显式路径：

- `wiki/capabilities/`
- `wiki/features/`
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

`raw` 文件与外部代码文件不使用 wikilink，而是通过 `sources`、`code_refs`、`api_entries`、`test_refs` 追踪。

---

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| capability/A 写入 `features: [[feature-F]]` | feature/F 的 `capabilities` 追加 A |
| capability/A 写入 `related_capabilities: [[capability-B]]` | capability/B 的 `related_capabilities` 追加 A |
| feature/F 写入 `related_features: [[feature-G]]` | feature/G 的 `related_features` 追加 F |

以下外部对象没有反向页面要求：

- `raw/*` 属于外部来源，只通过 `sources.path` 与 `sources.hash` 追踪
- 代码文件和 symbol 属于外部对象，只通过 `code_refs` 追踪
- 接口入口通过 `api_entries` 记录，不要求 wiki 反向链接
- 测试入口通过 `test_refs` 记录，不要求 wiki 反向链接

---

## 页面模板

### wiki/capabilities/{slug}.md

```yaml
---
title: ""
slug: ""
type: business
status: active
summary: ""
business_scenarios: []
features: []
related_capabilities: []
owner: ""
tags: []
confidence: 0.8
last_verified_at: YYYY-MM-DD
verification_status: pending
---
```

正文建议分区：

- `## Overview`
- `## Business Scenarios`
- `## Boundaries`
- `## Related Features`
- `## Related Capabilities`
- `## Known Gaps`
- `## Notes`

`capabilities` 页面只回答：

- 系统有哪些能力
- 这些能力解决什么业务问题
- 它和哪些 feature 相关
- 能力边界和关系是什么

不要在 capability 页里展开：

- 具体实现
- 调用链
- `code_refs`
- `api_entries`
- `test_refs`

`type` 只能是：

- `business`
- `system`

### wiki/features/{slug}.md

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

`sources` 推荐结构：

```yaml
sources:
  - path: "raw/requirements/user-group.md"
    kind: requirement
    hash: ""
  - path: "raw/designs/user-group.md"
    kind: design
    hash: ""
```

正文建议分区：

- `## Summary`
- `## Supported Capabilities`
- `## Business Flow`
- `## Constraints`
- `## Entry Points`
- `## Code Clues`
- `## Test Entry Points`
- `## Source Trace`
- `## Open Questions`

`features` 页面负责：

- 说明这个功能支撑哪些 capability
- 说明业务流程和约束
- 给出接口入口、代码线索、测试入口
- 给出足够定位实现的索引

`features` 页面不需要做：

- 完整调用链展开
- 大段实现细节解释
- 与 capability 页重复复述业务总结

---

## Code Ref 结构

推荐的 `code_refs` 结构：

```yaml
code_refs:
  - path: "services/user/service.ts"
    kind: file
    symbol: ""
    note: "用户 CRUD 主服务入口"
    confidence: 0.9
  - path: "services/user/group.ts"
    kind: function
    symbol: "createUserGroup"
    note: "用户组创建逻辑的关键函数"
    confidence: 0.8
```

使用规则：

- 如果整个文件都相关，直接存文件列表
- 如果文件很大但只涉及局部逻辑，使用 `path + symbol`
- 没核对过的路径和 symbol 不得虚构
- `code_refs` 只放在 feature 页，不放在 capability 页

`api_entries` 推荐结构：

```yaml
api_entries:
  - method: POST
    path: "/api/user-groups"
    note: "创建用户组入口"
```

`test_refs` 推荐结构：

```yaml
test_refs:
  - path: "tests/user_group_test.go"
    note: "用户组主流程测试"
```

---

## 检索顺序

DevWiki 的检索建立在三个 collection 之上：

- `wiki`
- `raw`
- `code`

检索通道按「成本 / 速度由低到高」阶梯升档，详见 `references/zatools-qmd.md`：

1. 本地 `grep` / 文件搜索（默认起点）
2. `zatools qmd search`（关键词召回）
3. `zatools qmd query`（语义召回，无 GPU/加速时走硬性 fallback）

任一档命中 top-K 且置信足够即停；只有当问题落到实现现实、代码归属，或文档证据仍不足时，才对 top-K 代码候选做一次定向本地排查。

如果 `wiki / raw` 已经足够回答问题，应先直接给出结论，并把“是否需要再做一次代码核对版汇总”作为可选后续，而不是默认继续读代码。

`qmd` 只负责加速召回，不是事实存储层，不能覆盖 `raw/`、`wiki/` 与真实代码内容。

---

## Workflow 约束

### 事实约束

- `raw/` 是只读原始资料
- feature 页中的 `sources` 必须记录真实 `path` 与 `hash`
- capabilities 与 features 必须建立在 raw 文档证据或已核对代码证据之上
- 事实与推断必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据

### 角色约束

- capability 页是业务能力与系统能力的中心页
- feature 页是功能说明与入口索引页
- `new / modify / unclear` 只允许作为 `/devwiki-ask` 的即时输出
- 不要创建 `wiki/changes/` 页面

### 变更风险策略

- 低风险：
  - 追加日志
  - 更新 `index.md`
  - 刷新 feature 页里的 `sources.hash`
- 中风险：
  - 将 feature 挂到已有 capability
  - 追加辅助性 `code_refs`
  - 补充 `api_entries` 或 `test_refs`
- 高风险：
  - 新建 capability
  - 变更 capability 边界
  - 新建 feature
  - 重挂 feature 与 capability 的关系
  - 替换主 `code_refs`

任何中高风险写入都必须先显式确认。

### 搜索纪律

- 禁止没有边界地盲搜
- 若经过几轮检索后仍然低置信，应停止继续扩散，并向用户提 1 到 3 个具体问题
- 优先询问接口 URL、关键文件、关键函数、路由、feature 名称或 capability 名称

### 编辑纪律

- 写正向关系时，同步写反向关系
- 不要静默删除坏链；应报告问题，或分流到 `/devwiki-refresh`、`/devwiki-check`
- 只需改一个字段或章节时，不要整页重写
- capability 和 feature 不要重复复述同一段正文

---

## 操作说明

- 使用 `zatools devwiki init` 初始化 DevWiki 工作区、安装 skills，并生成项目根桥接运行时文件
- 如需下载 qmd models，初始化完成后可在 DevWiki 工作区内手动执行 `zatools qmd download --root .`
- 使用 `devwiki-setup` 解释初始化约束、安装范围与运行时使用方式
- 对已有工作区补做或修复 qmd collection 注册、索引刷新与状态检查时，使用 `devwiki-qmd-sync`
- 使用 `devwiki-init` 基于现有 `raw/` 建立第一版 `capabilities + features` 知识骨架
- 使用 `devwiki-ingest` 吸收增量 raw 文档
- 使用 `devwiki-ask` 做通用问答或开发前的变更定性（`new / modify / unclear`）
- 使用 `devwiki-feature-doc` 在缺少功能页时从代码反向整理 feature
- 当 wiki 与 raw 或代码漂移时使用 `devwiki-refresh`
- 使用 `devwiki-check` 做确定性健康检查

这份 schema 应保持稳定，只在 DevWiki 的数据模型或工作流发生真实变化时修改。
