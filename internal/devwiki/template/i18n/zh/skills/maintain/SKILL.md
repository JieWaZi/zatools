---
name: "devwiki-maintain"
description: "对 DevWiki 进行证据一致性与知识健康维护。用于发现并修正 Wiki 与 raw/source/code 之间的冲突、遗漏、过期、过度摘要、引用缺失、历史机制不再适用、query 命中过期内容等问题。"
argument-hint: "<待维护范围，例如 wiki 全量、某个 capability/feature/workflow、某个 raw 文件、某次 query 失败案例>"
---

# /devwiki-maintain

## 一、设计来源与吸收原则

```text
raw/source/code = 证据层，只读，不直接改
wiki 当前页       = 当前理解层，可以维护和重写
outputs/report   = 维护过程报告，不是事实入口
relations/index/glossary = query 入口控制层
```

---

## 二、使用前置

开始前先读取：

- `references/evidence-grounding.md`
- `references/zatools-qmd.md`
- 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
- 涉及代码核对时，再读 `references/code-tracing.md`

---

## 三、维护目标

Maintain 的目标不是“整理格式”，而是保证 query 后续使用的是当前正确知识。

该 Skill 要发现并修正：

- Wiki 与 raw/source/code 不一致；
- Wiki 使用旧机制、旧结论、旧配置；
- raw 中关键设计点被 Wiki 漏掉；
- Wiki 过度摘要，导致关键规则、边界、状态、配置、异常场景丢失；
- 多个 Wiki 页面之间存在冲突；
- 多个 raw/source/code 对同一结论描述不一致；
- Wiki 结论缺少来源；
- 设计稿与代码实现存在差异但未归位；
- 过期页面仍可能被 qmd/query 检索并用于回答；
- relations/index/glossary 未同步，导致 Agent 走错入口。

---

## 四、核心硬规则：必须见任何中间文件

Maintain 可以生成差异审计，但不得把差异审计长期写成 active Wiki 页面。

禁止创建或保留这类 active 页面：

```text
wiki/sources/*implementation-errata*.md
wiki/features/*errata*.md
wiki/workflows/*errata*.md
wiki/*设计稿与实现差异*.md
```

除非用户明确要求保存审计报告，否则差异审计只输出在本次响应或 proposal 中。

如果用户要求保存报告，只能写入：

```text
wiki/outputs/<topic>-maintain-report-YYYY-MM-DD.md
```

并且必须使用：

```yaml
status: report
exclude_from_query: true
visibility: internal
```

报告顶部必须写：

```markdown
> 这是维护过程报告，不是功能事实来源。
> 当前功能规则以对应 Feature 为准。
> 当前实现路径以对应 Workflow 为准。
```

有效结论必须合并回权威页面：

| 差异类型 | 正确归位 |
|---|---|
| 能力边界差异 | Capability |
| 功能行为、规则、配置、边界差异 | Feature |
| 代码入口、调用链、实现差异 | Workflow |
| 故障现象、日志、修复路径差异 | Troubleshooting |
| 维护过程对照表、审计表 | outputs/report，不进入 active query |


本 Skill 默认处于 `discussion_only` 模式，除非用户明确授权写入，否则不得修改任何 Wiki 文件。

### 写入模式

| 模式 | 说明 | 允许动作 |
|---|---|---|
| `discussion_only` | 只讨论、只分析、只输出 proposal | 不得创建、修改、删除任何文件 |
| `dry_run` | 模拟写入，展示将要修改的内容 | 不得真正落盘 |
| `confirmed_write` | 用户明确授权写入 | 可按 proposal 修改 Wiki 文件 |

默认模式：

```text
discussion_only
```
---

## 五、问题类型

Maintain 必须按以下类型归类问题。

| 类型 | 含义 | 典型处理 |
|---|---|---|
| 覆盖遗漏 | raw/source 有关键事实，Wiki 未覆盖 | 补充对应权威页面 |
| 过度摘要 | Wiki 太薄，导致关键规则、边界、异常丢失 | 按模板重写相关小节 |
| 无来源结论 | Wiki 写了结论，但找不到 raw/source/code 支持 | 标记待确认，不能编造 source |
| 证据冲突 | 多个 source 或 Wiki 之间描述不一致 | 输出冲突表，需确认后改 |
| 历史失效 | Wiki 内容曾经适用，但当前版本/实现不再适用 | 标记历史范围，更新当前结论 |
| 实现偏差 | 设计文档与代码实现不一致 | Feature 写功能结论，Workflow 写实现差异 |
| 差异报告误落盘 | maintain 把 errata/report 写成 active Wiki | 移到 outputs/report 或删除，结论合并回权威页 |
| 关系错误 | relations/index/glossary 指向错误或缺失 | 修正关系和入口 |
| 查询污染 | 旧页面或报告页被 query 命中并误导回答 | 降级、排除、改入口、更新索引 |
| 模板不合规 | 标题、frontmatter、来源、状态字段不符合规范 | 低风险可直接修 |

---

## 七、维护等级

### 7.1 可直接修正

满足以下条件可直接修：

- Markdown 小节标题不统一；
- frontmatter 字段缺失；
- source path 格式错误；
- index/relations/glossary 漏更新；
- 明显断链；
- search_terms 缺失；
- 页面状态字段缺失；
- log 未记录；
- 维护报告误放 active 且未被其他页面引用，可移入 outputs 并加 `exclude_from_query: true`；
- 低风险措辞统一。

直接修正后仍需写维护报告。

### 7.2 需要 proposal 后修正

以下情况必须先输出 proposal：

- raw 中有关键规则遗漏，需要补充 Feature；
- Wiki 内容过度摘要，需要重写大段内容；
- 旧机制需要标记为历史；
- 多个页面之间结论冲突；
- Feature / Workflow 归属需要调整；
- 页面需要合并、拆分、重命名；
- active errata 页面需要拆解并合并回权威页面；
- 需要改 query 入口，避免继续命中过期内容。

### 7.3 必须人工确认

以下情况不得自动落盘：

- 多个 raw/source/code 对同一规则互相矛盾；
- 设计与代码实现不一致，且无法判断应以谁为准；
- 删除页面；
- 删除业务规则；
- 将 active 改为 deprecated；
- 改变能力/功能主边界；
- 影响外部客户、升级兼容、告警、接口行为的结论；
- 没有证据但需要写成确定事实。

---

## 八、工作流程

### Phase 1：确定维护范围

先明确维护范围：全量 Wiki、某个 Capability / Feature / Workflow / Troubleshooting、某份 raw/source、某次 query 错误回答、最近变更页面，或被错误创建的 errata/report 页面。

如果用户没有指定范围，默认优先检查：

1. 用户刚刚指出的问题相关页面；
2. active 页面；
3. query 频繁命中的页面；
4. relations/index/glossary 中作为入口的页面；
5. 最近更新页面。

### Phase 2：读取上下文

不得只读一个页面就修改。至少读取：目标 Wiki 页面、frontmatter sources、对应 raw/source、相关 Capability / Feature / Workflow / Troubleshooting、`relations.yml`、`index.md`、`glossary.md`。涉及实现差异时核对代码；涉及 errata/report 时读取其引用的 raw/code，并判断结论应归位到哪里。

### Phase 3：证据审计

检查每个关键结论是否能追溯到 source；source 是否存在、过期、冲突；页面是否把设计意图写成当前实现；是否把历史机制写成当前机制；是否把 maintain report 当作事实来源。

```markdown
## 证据审计

| Wiki 结论 | 当前来源 | 证据状态 | 问题类型 | 建议 |
|---|---|---|---|---|
|  |  | 有证据 / 缺证据 / 冲突 / 过期 / 错误引用 report |  |  |
```

### Phase 4：覆盖审计

从 raw/source/code 抽取关键知识信号，再检查 Wiki 是否覆盖。

| 页面类型 | 重点检查 |
|---|---|
| Capability | 能力目标、能力边界、覆盖 Feature、能力关系、能力级约束 |
| Feature | 功能目标、核心行为、关键规则、关键概念、重要配置、边界异常、验收关注点 |
| Workflow | 代码入口、调用链、关键逻辑、状态读写、配置处理、实现差异、测试引用、修改影响 |
| Troubleshooting | 现象、日志、诊断路径、可能原因、修复/恢复、相关功能 |

### Phase 5：差异归位审计

如果发现“设计稿与实现差异”“errata”“implementation difference”类内容，必须执行归位审计。

```markdown
## 差异归位审计

| 差异项 | 当前所在页面 | 正确归位 | 是否已合并 | 后续动作 |
|---|---|---|---|---|
|  | errata/report | Feature / Workflow / Capability / Troubleshooting / outputs | 是 / 否 |  |
```

处理规则：差异项不能长期只存在 errata/report；功能级差异合并到 Feature；实现级差异合并到 Workflow；排障级差异合并到 Troubleshooting；能力边界差异合并到 Capability；合并后 errata/report 必须移入 outputs 或删除，并排除 query。

### Phase 6：冲突与历史失效审计

检查页面之间冲突、raw/source/code 之间冲突、历史机制仍被写成当前机制、errata/report 与权威页面之间冲突。

| 情况 | 处理 |
|---|---|
| 新证据明确覆盖旧证据 | 更新当前结论，旧内容移入“历史说明” |
| 新旧证据不确定谁有效 | 标记冲突，等待确认 |
| 旧机制仍有版本价值 | 保留但标明适用版本/时间/条件 |
| 旧机制已完全不适用 | proposal 中建议 deprecated 或移入历史页 |
| 代码实现与设计不同 | Feature 写功能结论，Workflow 写当前实现差异 |
| errata/report 命中 query | 移入 outputs，设置 `exclude_from_query: true`，合并结论回权威页 |

### Phase 7：查询污染检查

重点避免 query 继续用旧内容或维护报告回答。检查过期页面、errata/report 是否 active，report 是否缺少 `exclude_from_query: true`，index/relations/glossary 是否仍指向旧页面或 report，qmd 搜索是否让旧页面/report 排在权威页前。

### Phase 8：输出维护 Proposal

高风险改动必须先输出 proposal。

```markdown
# Maintain Proposal

## 维护范围

## 发现的问题

| 页面 | 问题类型 | 严重级别 | 说明 | 建议动作 |
|---|---|---|---|---|

## 差异归位计划

| 差异项 | 当前所在 | 目标页面 | 修改方式 | 是否需确认 |
|---|---|---|---|---|

## 证据与来源

## Query 污染风险

## 需要你确认的问题

## 不建议自动修改的内容
```

### Phase 9：确认后落盘

用户确认后再执行高风险修改：更新权威页面，把 errata/report 中的有效结论合并回权威页面；维护报告写入 `wiki/outputs/` 并设置 `exclude_from_query: true`；更新 `relations.yml`、`index.md`、`glossary.md`、`log.md`；最后执行或提示执行：

```bash
zatools qmd update
zatools qmd status
```

禁止把维护报告写成 active Wiki 页面。

### Phase 10：维护后验证

再次搜索原问题关键词，检查 qmd 是否优先命中权威页面，errata/report 是否不会成为 query 入口，旧页面是否明确标记历史/废弃，relations/index/glossary 是否指向当前机制，并用 2 到 5 个 seed question 测试 query 是否还会答旧内容。

---

## 十、禁止事项

### 10.1 通用禁止

- 不要修改 raw/source 原始资料。
- 不要未读 source 就修改 Wiki。
- 不要只看单页上下文就重写结论。
- 不要把不确定内容写成确定事实。
- 不要静默处理冲突。
- 不要跳过 proposal 直接做高风险改动。
- 不要删除页面，除非用户明确确认。
- 不要维护完 Wiki 却不更新 index/relations/glossary。
- 不要维护完 Wiki 却不更新 qmd 索引。

### 10.2 差异报告禁止

- 不要把设计稿与实现差异长期写成 active Wiki 页面。
- 不要把 errata/report 放到 `wiki/sources/` 作为事实来源。
- 不要让 errata/report 成为 query 主入口。
- 不要在 Feature/Workflow 的 sources 中引用 maintain report 来支撑事实。
- 不要只写差异报告而不合并回权威页面。
- 不要在 index/relations/glossary 中指向 errata/report 作为当前结论。
- 不要用“以本页 + 代码为准”描述 report；当前结论应以 Feature/Workflow 为准。

### 10.3 证据禁止

- 不要编造 citation/source。
- 不要给无来源事实硬配来源。
- 不要删除无来源事实来掩盖问题，应先 flag。
- 不要把设计文档中的计划写成当前实现。
- 不要把代码当前行为写成产品设计，除非明确标注“实现现状”。

### 10.4 冲突处理禁止

- 不要自动选择看起来更合理的一方。
- 不要把多个版本的规则混写成一个规则。
- 不要把旧版本规则留在 active 页面中却不标注适用范围。
- 不要在未确认情况下把 active 页面改成 deprecated。
- 不要在未确认情况下合并或拆分主页面。

### 10.5 Query 防污染禁止

- 不要只改正文，不改 summary/status/search_terms。
- 不要只改页面，不改 index/relations/glossary。
- 不要保留多个互相冲突的 active 入口。
- 不要让旧页面继续作为主入口。
- 不要让 glossary 把旧术语优先指向旧机制。
- 不要让 outputs/report 参与正常 query。

---