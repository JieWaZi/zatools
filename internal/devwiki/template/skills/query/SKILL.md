---
name: "devwiki-query"
description: "当用户询问项目能力、功能说明、配置规则、现有实现位置、排障知识、对外说明、比较分析或已有 DevWiki 知识，并且不要求修改代码时使用。若用户要求开发、修改、修复、重构、补测试或提交代码，应使用 devwiki-code。"
argument-hint: "<问题>"
---

# /devwiki-query

## 核心约束（内联）

- **证据落地**：每个关键结论必须落到真实来源（wiki 页面、raw 文件或已核对代码）；`qmd` 只是召回加速器不是真相源；事实与推断必须拆开表达。
- **视图分层读取**：先用 `--view card` 判断命中，再用 `--view core` 回答主问题，只有 core 不够时读 `--view explain`。card 只放判断“是不是这个页面”的信息，core 放高频主结论，explain 放低频补充。
- **按需加载参考文档**：
  - 仅本地搜索低置信、需 qmd 升档时 → 读 `references/zatools-qmd.md`
  - 仅 locate_code / troubleshoot / 代码核实时 → 读 `references/code-tracing.md`
  - 仅用户要求保存回答、沉淀结论、写入文件时 → 读 `references/mutation-safety.md`

不要凭空回答项目事实；先查 DevWiki，再按需核对代码；每个关键结论都要能追溯来源。

## 与 devwiki-code 的边界

- 本 Skill 是只读查询：回答知识、规则、现有实现、代码位置、影响面和排障线索。
- 如果用户表达了要修改代码、开发功能、修 bug、调整接口/配置/业务逻辑、重构、补测试或提交代码，本 Skill 只做简短转交：使用 `devwiki-code`。
- 不要在修改类请求中继续执行 query 的 locate_code 全流程；修改类请求由 `devwiki-code` 走 `index → workflow card/core → 必要 topic core → 代码核对 → 测试 → 实现 → 验证`。
- 只有用户明确要求保存回答、沉淀结论或写入报告时，本 Skill 才写 `wiki/outputs/`；默认不写代码，也不写 Wiki 页面。

## Outputs

- 先给出语义识别结果：explain_topic / locate_code / troubleshoot / public_answer / compare
- 命中的 Topic / Workflow / Troubleshooting 页面
- 知识缺口、冲突和待确认项
- 必要的代码定位线索、修改影响和测试建议，仅当语义为 locate_code、troubleshoot 或明确要求实现核对
- 只有用户明确要求保存回答、沉淀结论、写入报告时，才写入 `wiki/outputs/<query-slug>.md` 并追加 `wiki/log.md`

## DevWiki Interaction

### Reads

**始终读取：**
- `config/project.yaml`（全读，14 行极小。路径相对于项目根目录，不是 `wiki/config/`）

**结构化定位（串行降级）：**
- 第 1 层：`zatools devwiki search index <关键词...> --root <真实文档库根目录>` → 命中则停，选最佳候选
- 第 2 层：index 无结果 → `zatools devwiki search glossary <关键词...> --root <真实文档库根目录>`
- 第 3 层：glossary 无结果 → 按语义使用 `zatools devwiki search <topic|workflow> <关键词...> --root <真实文档库根目录>`
- `index` 返回 `type`、`description`、`slug`；`glossary` 返回 `glossary`、`type`、`description`、`slug`；两者不返回 `score`
- 禁止并行搜多个源，禁止同时读多个候选的 card

**视图分层读取（禁止用 Read 工具直接读 topic/workflow/troubleshooting 文件）：**
- `zatools devwiki read <topic|workflow|troubleshooting> <slug> --view card --root <真实文档库根目录>` — 判断命中
- `zatools devwiki read <topic|workflow|troubleshooting> <slug> --view core --root <真实文档库根目录>` — 回答主问题
- `zatools devwiki read <topic|workflow|troubleshooting> <slug> --view explain --root <真实文档库根目录>` — 补充细节

**代码目录：**
- 仅当语义为 locate_code、troubleshoot，或用户明确要求当前实现、代码定位、配置项定位、日志关键字定位、修改影响或排障核实时读取

### Writes

- 默认不写任何文件
- 只有用户明确要求保存回答或沉淀结论时，才允许（需先读 `references/mutation-safety.md`）：
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`

## Query Principles

1. 不凭空回答项目事实。
2. 先查 Wiki 和 raw 来源，再按需核对代码。
3. 每个关键结论都有来源。
4. explain_topic 回答禁止直接展开代码、函数、文件路径、调用链、测试入口或修改方式。
5. 只有 locate_code / troubleshoot / 明确实现追问，才允许进入代码层面。
6. 如果文档已经足够回答，就不要为了“更稳”再默认展开代码阅读。
7. `zatools qmd` 只是召回工具，不是真相源。
8. 读取 view 时遵守知识经济学放置规则：先用 card 判断命中，再用 core 回答主问题，只有 core 不够时读取 explain。
9. 搜索和读取串行降级，不并行：`devwiki search index` → `devwiki search glossary` → `devwiki search topic/workflow`，候选逐个 card 验证，确认匹配才往下走。不为了“快”而并发读多个源或多个候选。

## 目录选择规则

| 用户意图 | 用户实际在问 | 优先目录 | 辅助目录 |
|---|---|---|---|
| explain_topic | 能力边界、功能规则、配置、状态、联动、流程规则 | `wiki/topics/` | `wiki/glossary.md`, `raw/` |
| locate_code | 代码在哪里、调用链、接口、内部逻辑、当前实现、影响范围 | `wiki/workflows/` | `wiki/topics/`, 之后才 `rg` |
| troubleshoot | 报错、不生效、怎么排查、怎么修复 | `wiki/troubleshooting/` | `wiki/workflows/`, `wiki/topics/` |
| public_answer | 对外说明、客户口径、官网/文档口径 | public 可见 Topic | 不读代码 |

典型顺序：

- 能力/功能问题：`topics`
- 实现问题：`workflows → topics → rg`
- 排障问题：`troubleshooting → workflows → topics`

## 去重与权威来源规则

- 能力边界、功能规则、参数、联动和设计流转以 topic 页面为准。
- 调用链、关键逻辑、代码定位、修改影响和测试入口以 workflow 页面为准。
- 日志、错误码、排障步骤以 troubleshooting 页面为准。
- 其他页面如果出现重复内容，只能作为摘要和导航，不作为权威来源。

## Workflow

### Step 1: 意图识别

1. 读取 `config/project.yaml`，确定代码仓配置和默认语言。
2. 如果用户要求修改代码、开发功能、修 bug、调整接口/配置/业务逻辑、重构、补测试或提交代码，停止本 Skill，转交 `devwiki-code`。
3. 判断问题语义：
   - `explain_topic`：用户想了解能力、功能、配置、边界、规则、联动。
   - `locate_code`：用户想了解代码、入口、调用链、接口、数据流、当前实现、影响面、测试入口，但不要求本轮修改代码。
   - `troubleshoot`：用户想解决报错、不生效、异常现象、诊断路径、修复建议。
   - `public_answer`：用户要求对外说明、客户口径、官网/文档口径。
   - `compare`：用户要求比较两个主题、方案或实现路径。

### Step 2: 串行定位候选页面

**原则：串行降级，逐层深入。每一层拿到结果后先判断，够用就停，不往下走。禁止并行搜多个源。**

从用户问题提取多关键词，按以下顺序串行搜索：

**第 1 层：`devwiki search index`**

```bash
zatools devwiki search index "<关键词1>" "<关键词2>" --root <真实文档库根目录>
```

命中后根据 `description`、`type`、`slug` 选出**最匹配的一个**入口。**index 命中后，禁止再跑当前 Step 的任何后续层（glossary、topic/workflow search），直接使用返回的 `type` 和 `slug` 进入 Step 3。**

**第 2 层：index 无结果 → `devwiki search glossary`**

```bash
zatools devwiki search glossary "<关键词1>" "<关键词2>" --root <真实文档库根目录>
```

命中后根据 `glossary` 和 `description` 判断是否能直接路由到返回的 `type` / `slug`。如果只是提供同义词，提取新关键词后回到第 1 层；如果能直接路由，进入 Step 3。

**第 3 层：glossary 无结果 → `devwiki search topic/workflow`**

根据语义选择 kind：

```bash
# explain_topic → topic
zatools devwiki search topic "<关键词1>" "<关键词2>" --root <真实文档库根目录>

# locate_code / troubleshoot → workflow 优先
zatools devwiki search workflow "<关键词1>" "<关键词2>" --root <真实文档库根目录>
```

命中后根据 `title`、`slug`、`score` 取最匹配候选，使用返回的 `slug` 进入 Step 3。

**第 4 层：全部无结果 → 低置信升档**

- 先读 `references/zatools-qmd.md`
- 再按其中路由规则升档到 `zatools qmd query`
- `devwiki search` 命中只是候选排序，最终结论必须回到真实 `wiki/` / `raw/` 页面
- `qmd search/query` 报错、超时、collection 未注册时，明示“本轮 qmd 不可用，已降级”

如果 `wiki/index.md`、`wiki/glossary.md` 或目标目录缺失，或者真实页面仍不足以支撑结论，输出：

```text
当前 Project Brain 没有足够信息支持该结论。
```

并建议先执行：

```text
devwiki-ingest
```

### Step 3: 候选验证（view=card，逐个读）

1. 对 Step 2 选出的**单个**最佳候选，用 `zatools devwiki read <type> <slug> --view card --root <真实文档库根目录>` 读取导航卡。
2. card 确认匹配 → 进入 Step 4 深度阅读。
3. card 不匹配 → 回到 Step 2 的结果中选**下一个**候选，再次 card 验证。不要并发读多个 card。
4. 所有候选都不匹配 → 回到 Step 2 的下一层继续搜索。

### Step 4: 按语义深度阅读

**所有 topic/workflow/troubleshooting 页面读取必须使用 `zatools devwiki read`，禁止用 Read 工具直接读文件。**

按语义控制深度：

| 语义 | 阅读路径 | 止步条件 |
|---|---|---|
| explain_topic | topic card → topic core → (topic explain) | core 足够即停，不触发 workflow |
| locate_code | topic card → topic core → workflow card → workflow core → (workflow explain) → 读 `code-tracing.md` → 代码核对 | 代码证据足够即停。**如果 workflow core 的代码表格已覆盖入口、核心逻辑、持久化三层且每行包含文件路径+行号+函数名，视为证据足够，直接据此组织回答，跳过实际代码文件读取。** |
| troubleshoot | troubleshooting card → troubleshooting core → workflow core → 读 `code-tracing.md` → 代码核对 | 排障路径清晰即停 |
| public_answer | topic card → topic core | core 足够即停，不读代码 |
| compare | 各对象 topic card → topic core | 差异清晰即停 |

每一层读取后判断：当前 view 是否足以回答用户问题？足够则停，不够才升到下一层。

### Step 5: 按需核对代码

以下情况必须核对代码：

- 问题语义识别为 `locate_code`
- 用户问「在哪里」「哪个文件」「哪个函数」「哪个接口」「当前实现是不是这样」
- 用户要求修改建议、影响分析、配置项定位、日志关键字定位
- 排障问题必须确认运行时行为或日志出处
- wiki / raw 证据不足以支撑结论

核对前先读 `references/code-tracing.md`。**但如果 workflow core 的代码表格已包含文件路径、行号和函数名等精确锚点，跳过 code-tracing.md，直接用表格锚点进入代码核对。**

核对顺序：

1. 优先从已读取的 workflow `core` view 中获取代码锚点。
2. 若已有明确代码锚点，用 `rg` 定向搜索。
3. 如果没有候选目录，再扩大到配置代码仓根。
4. 至少确认入口文件、关键函数、接口注册点、配置读取点、日志打印点或测试入口中的一层证据。

### Step 6: 按语义组织回答

每次回答先给一句语义识别：

```text
语义识别：explain_topic / locate_code / troubleshoot / public_answer / compare
```

explain_topic 使用 Topic 回答重点：

- 主题解决什么问题
- 核心行为和关键规则
- 功能边界：包含什么、不包含什么
- 关键配置与状态
- 需要看实现时进入哪个 Workflow

locate_code 使用 Workflow 回答重点：

- 代码在哪里
- 入口在哪里
- Topic 规则如何映射到实现
- 修改影响是什么
- 如何测试验证

troubleshoot 使用 Troubleshooting 回答重点：

- 现象
- 诊断路径
- 可能原因
- 修复 / 恢复
- 相关 Topic / Workflow

public_answer 回答重点：

- 对外结论
- 可公开范围
- 不应展开的内部信息
- 依据

compare 回答重点：

- 比较对象
- 相同点
- 差异点
- 适用场景
- 依据 / 待确认

### Step 7: 按需沉淀答案

只有用户明确要求保存回答、沉淀结论、写入报告时，才先读 `references/mutation-safety.md`，再创建 `wiki/outputs/<query-slug>.md` 并追加 `wiki/log.md`。
