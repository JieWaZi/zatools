---
name: "devwiki-query"
description: "当用户询问项目功能、设计细节、能力边界、代码位置、流程、配置、排障、对外说明或已有知识时使用。该 Skill 基于 DevWiki、glossary、zatools qmd 检索和必要的 rg 代码搜索回答问题。"
argument-hint: "<问题>"
---

# /devwiki-query

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`
> - 涉及写入、保存回答或沉淀新结论时，再读 `references/mutation-safety.md`

基于 Project Brain 回答问题。不要凭空回答项目事实；先查 DevWiki，再按需核对代码；每个关键结论都要能追溯来源。

## Inputs

- `question`：自然语言问题
- `role`（可选）：developer / internal_non_developer / external_user
- `format`（可选）：markdown / table / bullets / timeline
- `save-output`（可选）：仅当用户明确要求保存回答时使用

## Outputs

- 带来源的回答
- 命中的 capability / feature / workflow / troubleshooting
- 必要的代码定位线索
- 知识缺口、冲突和待确认项
- 修改影响和测试建议（仅当问题涉及开发或变更）
- 对外回答版本（仅当用户身份或问题场景要求 public 口径）
- 沉淀建议：值得 / 不需要

## DevWiki Interaction

### Reads

- `config/project.yaml`
- `config/search.yaml`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/capabilities/*.md`
- `wiki/features/*.md`
- `wiki/workflows/*.md`
- `wiki/troubleshooting/*.md`
- 本地代码目录：仅当问题要求实现现实、代码定位、配置项定位、日志关键字定位、修改影响或排障核实时读取

### Writes

- 默认不写任何文件
- 只有用户明确要求保存回答或沉淀结论时，才允许：
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`

## Query Principles

必须做到：

1. 不凭空回答项目事实。
2. 先查 Wiki 和 raw 来源，再按需核对代码。
3. 每个关键结论都有来源。
4. 明确标记知识缺口。
5. 明确标记冲突。
6. 对开发者问题给出代码定位线索。
7. 对外部用户问题只使用 public 可见知识。
8. `zatools qmd` 只是召回工具，不是真相源。

## 目录选择规则

| 用户意图 | 用户实际在问 | 优先目录 | 辅助目录 |
|---|---|---|---|
| 能力解释 | 为什么存在、有什么能力、能力边界是什么 | `wiki/capabilities/` | `wiki/features/` |
| 功能说明 | 功能是什么、参数如何取值、功能如何联动、设计怎么流转 | `wiki/features/` | `wiki/capabilities/`, `wiki/workflows/` |
| 代码定位 / 修改影响 | 代码在哪里、调用链怎么走、怎么改、影响哪些功能 | `wiki/workflows/` | `wiki/features/`, 之后才 `rg` |
| 故障排查 | 报错、不生效、怎么排查、怎么修复 | `wiki/troubleshooting/` | `wiki/workflows/`, `wiki/features/` |

典型顺序：

- 能力问题：`capabilities → features`
- 功能问题：`features → capabilities`
- 代码问题：`workflows → features → rg`
- 排障问题：`troubleshooting → workflows → features`

## 去重与权威来源规则

- 能力定义以 capability 页面为准。
- 功能规则、参数、联动和设计流转以 feature 页面为准。
- 调用链、关键逻辑、代码引用、修改影响和测试入口以 workflow 页面为准。
- 日志、错误码、排障步骤以 troubleshooting 页面为准。

其他页面如果出现重复内容，只能作为摘要和导航，不作为权威来源。

## Workflow

### Step 1: 意图识别与范围收敛

1. 读取 `config/project.yaml`，确定代码仓配置和默认语言。
2. 读取 `wiki/index.md`、`wiki/glossary.md`，建立全局上下文。
3. 判断用户身份：
   - 默认 `developer`
   - 用户说“客户、外部用户、官网、对外说明”时，按 `external_user`
   - 用户说“产品、测试、售前、支持、运维”时，按 `internal_non_developer`
4. 判断问题类型：

```text
explain_feature
locate_code
troubleshoot
compare
public_answer
design_detail
change_impact
```

如果 `wiki/index.md` 或 `wiki/glossary.md` 缺失，输出：

```text
当前 Project Brain 没有足够信息支持该结论。
```

并建议先执行：

```text
devwiki-ingest
```

### Step 2: 召回候选资料

查询词来源：

1. 用户原始问题中的关键词、接口、配置项、错误码、日志片段、功能名。
2. `wiki/glossary.md` 中的术语、别名、注意事项。
3. `wiki/index.md` 中的 capability / feature / workflow / troubleshooting 入口。
4. 候选页面 frontmatter 和正文链接中的 capability / feature / workflow / troubleshooting 关系。

召回规则：

1. 根据“目录选择规则”确定首查目录。
2. 默认先本地搜索 DevWiki 文档层：首查目录、辅助目录、`wiki/index.md`、`wiki/glossary.md`，必要时再查 `raw/`。
3. 召回分档、低置信升档和 qmd fallback 统一遵守 `references/zatools-qmd.md`。
4. `zatools qmd search` 只作为候选排序；命中后必须读取真实 `wiki/` / `raw/` 页面，再按事实归属去重。
5. 对 `public_answer` 只读取 public 可见页面和可公开摘录，不读代码仓库。
6. 候选数量受控：top-K（K ≤ 12），优先读高相关页面。

如果 `qmd search/query` 报错、超时、collection 未注册、cache 不可写或模型缺失，按 `references/zatools-qmd.md` 降级为本地 Wiki 搜索，并在回答中明示“本轮 qmd 不可用，已降级”。

如果文档已经足够回答，就不要为了“更稳”再默认展开代码阅读。

### Step 3: 按需核对代码

以下情况必须核对代码：

- 用户问「在哪里」「哪个文件」「哪个函数」「哪个接口」「当前实现是不是这样」
- 用户要求修改建议、影响分析、配置项定位、日志关键字定位
- 排障问题必须确认运行时行为或日志出处
- wiki / raw 证据不足以支撑结论

核对顺序：

1. 优先读取相关 workflow 页中的 `code_refs`、`api_entries`、`test_refs`。
2. 若已有明确代码锚点，用 `rg` 定向搜索。
3. 如果没有候选目录，再扩大到配置代码仓根。
4. 至少确认入口文件、关键函数、接口注册点、配置读取点、日志打印点或测试入口中的一层证据。

### Step 4: 回答

回答结构：

```markdown
## 结论

## 依据

## 代码定位线索

## 冲突 / 待确认项

## 修改影响 / 测试建议

## 沉淀建议
```

只有涉及代码定位、修改、排障时才输出代码定位线索。

### Step 7: 按需沉淀答案

只有用户明确要求保存回答、沉淀结论、写入报告时，才允许写 `wiki/outputs/<query-slug>.md` 并追加 `wiki/log.md`。保存前先给出拟写路径和摘要。

## Error Handling

- **Wiki 证据不足**：说明不足点，建议执行 `devwiki-ingest` 或 `devwiki-code-to-doc`。
- **检索低置信**：停止扩散，向用户问 1 到 3 个具体问题。
- **文档与代码冲突**：明确列出「文档描述」和「代码现状」。
- **需要当前实现但未核对代码**：不要给确定结论。
