---
name: "devwiki-ingest"
description: "当用户提供设计文档、需求文档、接口文档、配置说明、部署文档、测试文档、会议纪要、排障记录、代码逻辑片段或讨论结论，并要求消化、导入、生成 Wiki、构建知识库时使用。该 Skill 将原始资料转换为 capability、feature、workflow、troubleshooting、术语、关系和需要对话确认的问题。本版统一使用中文标题，并明确 Capability / Feature / Workflow 三层边界：Capability 是能力地图，Feature 是功能契约，Workflow 是实现路径。"
argument-hint: "<文档路径、目录、文本片段或待摄入范围>"
---

# /devwiki-ingest

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`
> - 生成 capability 页面前，优先读取 `references/capability_template.md`
> - 生成 feature 页面前，优先读取 `references/feature_template.md`
> - 生成 workflow 页面前，优先读取 `references/workflow_template.md`

将一份或一批原始资料转成可维护知识。不要直接写散文总结；必须先做分类、控粒度、给 proposal，等用户确认中高风险内容后再落盘。

本 Skill 的核心目标：

1. 从原始资料中抽取可复用的知识，而不是搬运原文；
2. 保留后续问答、开发、测试、排障会依赖的关键设计信号；
3. 避免 capability / feature / workflow / troubleshooting 之间重复维护；
4. 避免 feature 过薄，导致后续 Agent 无法回答规则、边界、联动问题；
5. Markdown 标题统一使用中文，避免中英文小节名混用。

---

## 一、三层主模型

DevWiki 的主知识围绕三层和排障组织：

| 层级 | 路径 | 作用 |
|---|---|---|---|
| capability | `wiki/capabilities/<slug>.md` | 业务/系统能力：能力边界、作用效果、覆盖功能、能力协作 |
| feature | `wiki/features/<slug>.md` | 功能契约/功能知识：功能目标、核心行为、关键规则、关键概念、重要配置、联动、边界、验收关注点 |
| workflow | `wiki/workflows/<slug>.md` | 面向编程的工程定位：入口、调用链、关键逻辑、代码引用、修改影响、实现差异核对 |
| troubleshooting | `wiki/troubleshooting/<slug>.md` | 故障现象、日志、错误码、诊断路径、证据、修复建议 |
---

## 二、Capability / Feature / Workflow 三层边界

一句话区分：

```text
Capability = 系统具备什么能力、能力边界是什么
Feature = 具体功能的行为和规则是什么
Workflow = 功能在代码中的实现路径怎么走
```

三层页面不能互相复制内容。每一层只维护自己负责的“权威事实”，其他层只做摘要和链接。

| 层级 | 核心问题 | 权威内容 | 不应该写什么 |
|---|---|---|---|
| Capability | 系统具备什么能力？能力边界是什么？ | 能力定义、业务价值、能力范围、覆盖 Feature、能力间关系、能力级约束 | 具体功能规则、状态机、决策表、代码路径 |
| Feature | 这个功能怎么表现？有哪些关键规则？ | 功能目标、用户场景、触发条件、核心行为、关键规则、关键概念、重要配置、边界异常、验收关注点 | 代码入口、函数名、调用链、实现分支、完整排障步骤 |
| Workflow | 这个功能在代码里怎么实现？ | 代码入口、调用链、类/模块/函数、状态读写、配置处理、异常实现、测试引用、修改影响 | 完整业务背景、完整 Feature 规则复述、能力价值说明 |

---

### 2.1 三层之间的引用方式

正确关系：

```text
Capability → 列出并链接 Feature
Feature → 说明功能规则，并链接 Workflow
Workflow → 映射 Feature 规则到代码实现
Troubleshooting → 链接 Feature / Workflow，提供排障路径
```

推荐写法：

* [[feature-vip-failover]]：负责 VIP 接管行为，详细规则见 Feature 页面。

* 实现定位见：[[workflow-vip-failover]]

---

### 2.2 三层归属判断表

| 信息类型 | 写入位置 | 说明 |
|---|---|---|
| 能力定义、业务价值、能力边界 | Capability | 作为能力页权威事实 |
| 覆盖哪些功能 | Capability | 只列 Feature 摘要和链接 |
| 功能目标、功能行为、功能规则 | Feature | 作为功能页权威事实 |
| 用户场景、触发条件、边界异常 | Feature | 面向理解和测试 |
| 状态/角色的功能含义 | Feature | 只解释功能影响 |
| 具体状态判断代码 | Workflow | 写代码入口和判断位置 |
| 决策规则的功能结果 | Feature | 保留规则摘要或关键表 |
| 决策规则的代码分支 | Workflow | 映射到实现位置 |
| 配置对行为的影响 | Feature | 写功能影响 |
| 配置读取/校验/下发代码 | Workflow | 写实现路径 |
| 代码路径、函数名、调用链 | Workflow | Feature 禁止写 |
| 故障现象、日志、修复步骤 | Troubleshooting | Feature 只链接 |
| 修改影响文件、测试文件 | Workflow | 写工程影响 |

---

## 三、输入
- `raw/**/*.md`：已放入 DevWiki 的原始需求、设计、功能说明、测试方案等；
- `config/project.yaml`：项目名称、语言、agent、代码仓配置；
- `config/search.yaml`：qmd collection 和模型配置。

可处理资料类型：

- 设计文档、需求文档、接口文档、配置说明、部署文档、测试文档；
- 会议纪要、排障记录、变更评审记录；
- 用户粘贴的代码逻辑、运行日志或用户与 AI 的讨论结论。

---

## 四、输出

每次 ingest 只允许创建或更新：

- `wiki/capabilities/<slug>.md`
- `wiki/features/<slug>.md`
- `wiki/workflows/<slug>.md`
- `wiki/troubleshooting/<slug>.md`
- `wiki/relations.yml`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `wiki/outputs/<slug>.md`：仅当用户明确要求保存报告

---

## 五、粒度规则

默认一个功能主题最多对应：

- 1 个 capability
- 1 个 feature
- 1 个 workflow

不要因为出现多个 API、多个模块、多个分支就自动拆多个 workflow。只有满足以下条件之一，才允许在 proposal 中建议拆分：

- 用户明确要求拆分；
- 文档里包含多个明显独立功能；
- 两条调用链属于不同运行时服务，且修改影响完全不同；
- 单页会超过可维护长度，且拆分后的边界能用自然语言说清楚；
- 原始资料包含多个互不依赖的功能目标，无法归为同一 feature。

拆分前必须先对话确认。

---

## 六、来源规则

来源信息内联写入目标页面：

```yaml
sources:
  - path: "raw/designs/example.md"
    kind: design
    hash: ""
    title: ""
    confidence: medium
    notes: ""
```

要求：

- 每个重要事实必须能回到 `raw/`、已有 Wiki 页面或已核对代码证据。
- 用户粘贴内容使用 `path: "pasted context"`，并在 `notes` 中说明来源。
- 不确定内容不得写成确定事实。
- 历史设计、会议纪要、排障记录要标明时间和适用版本。
- 文档冲突必须进入 proposal，不能静默选择一个版本。
- 如果设计稿与代码核对不一致，Feature 中写“设计意图/功能规则”，Workflow 中写“实现差异/代码证据”。

---

## 七、页面模板

### 7.1 Capability 模板

```text
references/feature_template.md
```
### 7.2 Feature 模板
```text
references/feature_template.md
```
### 7.3 Workflow 模板

```text
references/workflow_template.md
```

### 7.4 Troubleshooting 模板

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

---

## 五、工作流程

### Phase 1：解析输入

1. 展开 `source` 范围，得到待处理资料列表。
2. 提取每份资料的标题、类型、版本、日期、owner、hash、关键术语。
3. 初步识别候选 Capability、Feature、Workflow、Troubleshooting。
4. 批量处理前，先抽样 3 到 5 份验证抽取质量、命名和页面粒度。

不要直接从原始资料跳到最终 Wiki 页面。

---

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

说明：

- 知识信号不是最终 Wiki。
- 不要求所有分类都存在。
- 如果某类信号很弱，不要强行生成对应页面。
- 如果某类信号很强，必须在 proposal 中说明建议写入哪里。

---

### Phase 3：检查已有知识

1. 先读取或搜索：
    - `wiki/index.md`
    - `wiki/relations.yml`
    - `wiki/glossary.md`
2. 再搜索：
    - `wiki/capabilities/`
    - `wiki/features/`
    - `wiki/workflows/`
    - `wiki/troubleshooting/`
3. 本地命中不足时，按 `references/zatools-qmd.md` 使用：

```bash
zatools qmd search "<关键词>"
```

4. 命中相似页面时，判断是：
    - 更新已有页面；
    - 新增页面；
    - 标记冲突；
    - 需要用户确认。

不要在未检查已有知识的情况下直接新建页面。

---

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

拆分、重命名、合并、改主关系前必须先让用户确认。

---

### Phase 5：读取模板并生成草稿

根据目标页面类型读取对应模板：

- Capability：读取 capability 模板；
- Feature：读取 feature 模板；
- Workflow：读取 workflow 模板；
- Troubleshooting：使用内置模板。

生成草稿时：

1. 只写该页面类型负责的内容；
2. 其他层内容只写摘要和链接；
3. 所有关键事实必须带来源；
4. 不确定内容写入“来源说明”或 proposal 的待确认问题；
5. 页面小节标题统一使用中文。

---

### Phase 6：输出 Ingest Proposal

写入前必须先输出 proposal。

```markdown
# Ingest Proposal

## 输入来源

## 建议写入

| 页面 | 类型 | 动作 | 原因 | 置信度 |
|---|---|---|---|---|

## 分类判断

| 内容 | 建议归属 | 不放到其他分类的原因 |
|---|---|---|

## 关系更新建议

## 术语更新建议

## 冲突与不确定内容

## 需要你确认的问题

## 暂不写入的内容
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

---

### Phase 7：确认后落盘

用户确认后再执行：

1. 创建或更新目标页面。
2. 更新 `wiki/relations.yml`。
3. 更新 `wiki/index.md`。
4. 更新 `wiki/glossary.md`。
5. 追加 `wiki/log.md`。
6. 如果用户要求保存报告，写入 `wiki/outputs/<slug>.md`。
7. 执行或提示执行：

```bash
zatools qmd update
zatools qmd status
```

---

## 七、关系文件规则

`wiki/relations.yml` 只保存摘要关系，不做全量关系数据库。

推荐结构：

```yaml
capabilities:
  <capability-slug>:
    features:
      - <feature-slug>
    related_capabilities: []

features:
  <feature-slug>:
    capabilities:
      - <capability-slug>
    workflow: <workflow-slug>
    related_features: []
    troubleshooting: []

workflows:
  <workflow-slug>:
    features:
      - <feature-slug>
    touches:
      - path: ""
        symbol: ""

troubleshooting:
  <troubleshooting-slug>:
    features:
      - <feature-slug>
    workflows:
      - <workflow-slug>
```

完整关系优先写在页面 front matter，未来可由脚本生成 SQLite 索引。

---

## 八、落盘前检查清单

落盘前检查：

- 是否已读取对应模板；
- 是否每个重要事实都有来源；
- 是否没有跳过 proposal；
- 是否没有在检查已有知识前直接新建页面；
- 是否没有把 Capability / Feature / Workflow / Troubleshooting 的职责混写；
- 是否没有复制其他页面的权威内容；
- 是否没有编造代码路径、函数、接口、模块名；
- 是否页面小节标题统一为中文；
- 是否更新了 `relations.yml`、`index.md`、`glossary.md` 和 `log.md`。

---

## 九、禁止事项

### 9.1 通用禁止

- 不要把原始文档直接改写成 Wiki。
- 不要只生成普通摘要。
- 不要在未检查已有 Wiki 的情况下新建页面。
- 不要把不确定内容写成确定事实。
- 不要静默处理冲突。
- 不要批量生成大量碎片页面。
- 不要为了填模板强行生成无意义小节。
- 不要混用中英文小节标题。

### 9.2 Capability 禁止

- 不要写代码路径、函数名、handler、调用链。
- 不要复制 Feature 的完整功能规则。
- 不要写具体 Feature 的状态机、决策表和配置细节。
- 不要把 Capability 写成 Feature 汇总长文。

### 9.3 Feature 禁止

- 不要写代码文件路径、函数名、handler、调用链。
- 不要写具体实现分支。
- 不要把 Feature 写成详细设计原文的一比一复刻。
- 不要把 Feature 简化成只有“功能摘要”的薄文档。
- 不要遗漏会影响后续问答质量的关键功能规则。
- 不要复制 Workflow 的实现路径。

### 9.4 Workflow 禁止

- 不要复制 Capability 的能力价值说明。
- 不要复制 Feature 的完整功能规则。
- 不要在没有代码证据时写确定代码路径。
- 不要把 Workflow 写成业务说明文。
- 不要遗漏修改影响和测试验证建议。

### 9.5 Troubleshooting 禁止

- 不要写完整功能全貌。
- 不要写大量实现背景。
- 不要把未确认的现场经验写成通用结论。
- 不要和 Feature / Workflow 重复维护规则和实现路径。

### 9.6 Glossary 写入限制

`wiki/glossary.md` 只保存**全局术语**，不要写入单个 Feature 的局部概念。

#### 允许写入

满足以下任一条件才可写入：

- 多个页面都会用到；
- 是业务 / 系统 / 架构核心概念；
- 有缩写、别名、历史叫法，容易歧义。

#### 不要写入

- 配置项、字段、状态文件、代码常量；
- API、函数名、类名、handler；
- 功能规则、决策规则、状态机条目；
- 仅用于检索的关键词。

#### 分流规则

- 局部概念 → Feature `关键概念`
- 配置 / 状态 → Feature 或 Workflow
- 代码符号 → Workflow
- 检索词 → `search_terms`
- 日志 / 错误码 → Troubleshooting
单次 ingest 默认最多新增 **3 个** glossary 术语；超过 3 个必须在 proposal 中说明并等待确认。
---

## 十、Ingest Report

确认落盘后输出到终端（只能到终端）：

```markdown
# Ingest Report

## 已处理来源

## 新增页面

## 更新页面

## 关系更新

## 术语更新

## 分类结果

| 页面 | 类型 | 动作 | 说明 |
|---|---|---|---|

## 跳过内容

## 已确认决策

## 待确认问题

## 下一步建议
```