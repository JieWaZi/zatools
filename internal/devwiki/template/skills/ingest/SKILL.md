---
name: "devwiki-ingest"
description: "当用户提供设计文档、需求文档、接口文档、配置说明、部署文档、测试文档、会议纪要、排障记录、代码逻辑片段或讨论结论，并要求消化、导入、生成 Wiki、构建知识库时使用。该 Skill 将原始资料转换为 capability、feature、workflow、troubleshooting、术语、入口导航和需要对话确认的问题。"
argument-hint: "<文档路径、目录、文本片段或待摄入范围>"
---

# /devwiki-ingest

将一份或一批原始资料转成可维护知识。不要直接写散文总结；必须先分类、控粒度、检查已有 Wiki、输出 proposal，等用户在 proposal 之后显式确认后再落盘。

## 快速执行规则

1. 默认处于 `discussion_only`：用户要求“生成 Wiki / 导入资料 / 构建知识库”只表示启动 ingest 分析流程，不等于允许落盘。
2. 写入前必须先输出 Ingest Proposal；只有用户在 proposal 后明确回复“确认落盘”或“按 proposal 写入”，才进入 `confirmed_write`。
3. 不要从原始资料直接跳到最终页面；先抽取知识信号，再判断写入位置。
4. 写新页面前必须先检查 `wiki/index.md`、`wiki/glossary.md` 和相关 Wiki 目录，避免重复建页。
5. 默认一个功能主题最多对应 1 个 Capability、1 个 Feature、1 个 Workflow；拆分、合并、重命名或改主关系前必须让用户确认。
6. Capability / Feature 只写业务和功能事实；代码路径、函数名、handler、调用链和 `kind: code` 只进入 Workflow 或 Troubleshooting。
7. 页面小节标题统一使用中文，避免中英文小节名混用。

本 Skill 的核心目标：

- 从原始资料中抽取可复用的知识，而不是搬运原文；
- 保留后续问答、开发、测试、排障会依赖的关键设计信号；
- 避免 capability / feature / workflow / troubleshooting 之间重复维护；
- 避免 feature 过薄，导致后续 Agent 无法回答规则、边界、联动问题。

## Reference 读取策略

默认只读本文件，不要在任务开始时一次性读取所有 references。只有触发条件满足时再读取对应 reference：

| 触发条件 | 读取 |
|---|---|
| 写入、重命名、拆分、合并或改主关系前 | `references/mutation-safety.md` |
| 判断页面边界、事实来源、`sources` 或 `code_refs` 时 | `references/evidence-grounding.md` |
| 本地 Wiki 命中低置信、噪声过大或需要 qmd 时 | `references/zatools-qmd.md` |
| 需要结合当前代码、核对实现或写入 `code_refs` 时 | `references/code-tracing.md` |
| 确认要生成 Capability / Feature / Workflow 页面草稿时 | 只读取对应模板：`references/capability_template.md`、`references/feature_template.md` 或 `references/workflow_template.md` |

模板和共享 reference 是决策点工具，不是启动成本。先按本文件完成分类和 proposal；进入具体页面草稿或证据细节时再读取对应文件。

## 知识层级

| 层级 | 路径 | 作用 |
|---|---|---|
| capability | `wiki/capabilities/<slug>.md` | 业务/系统能力：能力边界、作用效果、覆盖功能、能力协作 |
| feature | `wiki/features/<slug>.md` | 功能契约/功能知识：功能目标、核心行为、关键规则、关键概念、重要配置、联动、边界、验收关注点 |
| workflow | `wiki/workflows/<slug>.md` | 面向编程的工程定位：入口、调用链、关键逻辑、代码引用、修改影响、实现差异核对 |
| troubleshooting | `wiki/troubleshooting/<slug>.md` | 故障现象、日志、错误码、诊断路径、证据、修复建议 |

一句话区分：

```text
Capability 是能力地图：系统具备什么能力、能力边界是什么
Feature 是功能契约：具体功能的行为和规则是什么
Workflow 是实现路径：功能在代码中的实现路径怎么走
```

## 输入和输出

输入可以来自：

- `raw/**/*.md` 中的设计、需求、功能说明、测试方案、会议纪要、排障记录；
- 用户粘贴的文本、代码逻辑、运行日志或讨论结论；
- 已有 Wiki 页面和必要的代码核对结果。

进入 `confirmed_write` 后，每次 ingest 最多允许创建或更新：

- `wiki/capabilities/<slug>.md`
- `wiki/features/<slug>.md`
- `wiki/workflows/<slug>.md`
- `wiki/troubleshooting/<slug>.md`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `wiki/outputs/<slug>.md`：仅当用户明确要求保存报告

## 工作流程

### Phase 1：解析输入

1. 展开 `source` 范围，得到待处理资料列表。
2. 提取每份资料的标题、类型、版本、日期、owner、hash、关键术语。
3. 初步识别候选 Capability、Feature、Workflow、Troubleshooting。
4. 批量处理前，先抽样 3 到 5 份验证抽取质量、命名和页面粒度。

不要直接从原始资料跳到最终 Wiki 页面。

### Phase 2：抽取知识信号

对每份资料先输出“知识信号”，用于防遗漏和后续分类。

```markdown
## 知识信号

### 能力信号

### 功能信号

### 工程信号

### 排障信号

### 关键术语

### 与已有知识的可能关系

### 冲突或不确定内容

### 需要用户确认的问题
```

知识信号不是最终 Wiki。如果某类信号很弱，不要强行生成对应页面；如果某类信号很强，必须在 proposal 中说明建议写入哪里。

### Phase 3：检查已有知识

1. 先读取或搜索：
   - `wiki/index.md`
   - `wiki/glossary.md`
2. 再搜索：
   - `wiki/capabilities/`
   - `wiki/features/`
   - `wiki/workflows/`
   - `wiki/troubleshooting/`
3. 本地 Wiki 命中低置信、噪声过大或无法排序时，再读取 `references/zatools-qmd.md` 并按其中策略使用 `zatools qmd search`。
4. 如果 `zatools qmd search` 报错、超时、collection 未注册或 cache 不可写，降级为本地 Wiki 搜索，并在 proposal 中说明本轮 qmd 不可用。
5. 命中相似页面时，判断是更新已有页面、新增页面、标记冲突，还是需要用户确认。

不要在未检查已有知识的情况下直接新建页面。

### Phase 4：分类和控粒度

按以下原则决定是否生成页面：

1. 一个功能主题默认最多对应：
   - 1 个 Capability；
   - 1 个 Feature；
   - 1 个 Workflow。
2. Troubleshooting 只在资料包含明确故障、日志、诊断或修复路径时生成。
3. 不要因为出现多个接口、多个字段、多个分支就拆成多个页面。
4. 只有在以下情况才建议拆分：
   - 用户明确要求；
   - 文档包含多个明显独立功能；
   - 两条调用链属于不同运行时服务，且修改影响完全不同；
   - 单页会过长，且拆分边界清晰；
   - 资料中出现独立排障闭环。

拆分、重命名、合并、改主关系前必须先让用户确认，并读取 `references/mutation-safety.md`。

### Phase 5：生成草稿

根据目标页面类型读取对应模板：

- Capability：读取 `references/capability_template.md`
- Feature：读取 `references/feature_template.md`
- Workflow：读取 `references/workflow_template.md`
- Troubleshooting：使用本文件的简化模板；如后续拆出专用模板，再按专用模板读取

生成草稿时：

1. 只写该页面类型负责的内容；
2. 其他层内容只写摘要和链接；
3. 所有关键事实必须带来源；
4. 不确定内容写入“来源说明”或 proposal 的待确认问题；
5. 页面小节标题统一使用中文。

### Phase 6：输出 Ingest Proposal

写入前必须先输出 proposal。

```markdown
# Ingest Proposal

## 输入来源

## 拟写入文件

| 路径 | 类型 | 动作 | 风险等级 | 原因 | 置信度 |
|---|---|---|---|---|---|

## 分类判断

| 内容 | 建议归属 | 不放到其他分类的原因 |
|---|---|---|

## 入口和链接更新建议

| 路径 | 动作 | 原因 |
|---|---|---|

## 术语更新建议

## 冲突与不确定内容

## 需要你确认的问题

## 暂不写入的内容

## 等待确认

如果以上路径、动作和内容摘要都认可，请明确回复“确认落盘”或“按 proposal 写入”。
```

必须先确认的情况：

- 多个 Capability 名称都合理；
- 一个文档可能拆成多个 Feature；
- 功能边界和能力边界混在一起；
- 文档描述与现有 Wiki 冲突；
- 旧页面需要重命名、合并、删除或改主关系；
- 需要写确定代码路径但尚未核对代码；
- Workflow 是否需要拆分无法判断；
- 原文包含大量规则表，不确定保留在 Feature 还是拆到 Workflow。

### Phase 7：确认后落盘

只有在用户明确确认 Ingest Proposal 后，才进入 `confirmed_write` 并执行：

1. 创建或更新目标页面。
2. 更新 `wiki/index.md`。
3. 更新 `wiki/glossary.md`。
4. 追加 `wiki/log.md`。
5. 如果用户要求保存报告，写入 `wiki/outputs/<slug>.md`。
6. 执行或提示执行：

```bash
zatools qmd update
zatools qmd status
```

7. 如果本次修改了 `wiki/capabilities/`、`wiki/features/` 或 `wiki/workflows/`，必须执行：

```bash
zatools devwiki graph --check
```

如果 graph check 失败，不得宣称 ingest 完成；先修复错误，或把需要人工确认的关系问题带回 proposal。

## 详细约束

下面内容是写作和落盘前的校验规则。先按“快速执行规则”和“工作流程”推进；只有进入对应决策点时再回看本节或读取 reference。

### 来源和证据

- 每个重要事实必须能回到 `raw/`、已有 Wiki 页面或已核对代码证据。
- 用户粘贴内容使用 `path: "pasted context"`，并在 `notes` 中说明来源。
- Capability / Feature 的 `sources` 不写代码文件路径、函数名、handler、调用链或 `kind: code`。
- 代码证据统一写入 Workflow 或 Troubleshooting 的 `code_refs`；写入前读取 `references/evidence-grounding.md`。
- 不确定内容不得写成确定事实。
- 历史设计、会议纪要、排障记录要标明时间和适用版本。
- 文档冲突必须进入 proposal，不能静默选择一个版本。
- 如果设计稿与代码核对不一致，Feature 中写“设计意图/功能规则”，Workflow 中写“实现差异/代码证据”。

### Troubleshooting 简化模板

仅当资料包含明确故障、日志、诊断或修复路径时生成。

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

# <故障名>

## 现象

## 诊断路径

## 日志与错误关键字

## 可能原因

## 修复 / 恢复

## 相关功能

## 相关工程流程

## 来源说明

## 检索词
```

### 落盘前检查

- 是否处于 `confirmed_write` 模式；
- 是否有用户在 Ingest Proposal 后的明确确认；
- 实际写入路径是否完全包含在 Ingest Proposal 的“拟写入文件”表内；
- 实际写入动作是否完全匹配 Ingest Proposal；
- 是否没有写入 proposal 未列出的文件；
- 是否已读取对应模板；
- 是否每个重要事实都有来源；
- 是否没有跳过 proposal；
- 是否没有在检查已有知识前直接新建页面；
- 是否没有把 Capability / Feature / Workflow / Troubleshooting 的职责混写；
- 是否没有复制其他页面的权威内容；
- 是否没有编造代码路径、函数、接口、模块名；
- 是否页面小节标题统一为中文；
- 是否更新了 `index.md`、`glossary.md` 和 `log.md`。

### 禁止事项

- 不要把原始文档直接改写成 Wiki。
- 不要只生成普通摘要。
- 不要在未检查已有 Wiki 的情况下新建页面。
- 不要把不确定内容写成确定事实。
- 不要静默处理冲突。
- 不要批量生成大量碎片页面。
- 不要为了填模板强行生成无意义小节。
- 不要混用中英文小节标题。
- Capability 不写代码路径、函数名、handler、调用链，也不复制 Feature 的完整功能规则。
- Feature 不写代码文件路径、函数名、handler、调用链，也不复制 Workflow 的实现路径。
- Workflow 不复制 Capability 的能力价值说明，也不复制 Feature 的完整功能规则。
- Troubleshooting 不写完整功能全貌，不写大量实现背景，不把未确认的现场经验写成通用结论。

### Glossary 写入限制

`wiki/glossary.md` 只保存**全局术语和稳定入口概念**，不要写入单个 Feature 的局部概念。

Glossary 应优先沉淀业务能力、系统能力、跨页面主题、稳定领域概念和常用别名。它是后续 query 的“概念入口”，不是从 raw 文档中摘出来的名词清单。

#### 允许写入

满足以下任一条件才可写入：

- 多个页面都会用到；
- 是业务 / 系统 / 架构核心概念；
- 有缩写、别名、历史叫法，容易歧义。

#### 不要写入

- 单个进程、文件名、脚本、状态文件、数据文件；
- 配置项、字段、配置键、CSV 列、状态码、动作码、代码常量；
- API、函数名、类名、handler；
- 功能规则、决策规则、状态机条目；
- 日志关键字、错误片段、排障命令；
- 仅用于检索的关键词。

不要把单个进程、文件名、字段名、配置键、CSV 列、动作码、函数名、日志关键字或检索词当作 glossary 术语。这些内容应该进入 Feature、Workflow、Troubleshooting 或页面 frontmatter 的 `search_terms`。

#### 命名规则

术语名称优先使用能力型或主题型表达，而不是原文中的实现名词。

- 好例子：`告警采集与外发`
- 坏例子：`告警节点进程`

#### 说明写法

一个术语说明要回答“这个能力/主题解决什么问题、覆盖哪些关键行为、和哪些场景相关”。不要只解释某个低层对象的字面含义。

#### 分流规则

- 局部概念 → Feature `关键概念`
- 配置 / 状态 → Feature 或 Workflow
- 代码符号 → Workflow
- 检索词 → `search_terms`
- 日志 / 错误码 → Troubleshooting

单次 ingest 默认最多新增 **3 个** glossary 术语；超过 3 个必须在 proposal 中说明并等待确认。
