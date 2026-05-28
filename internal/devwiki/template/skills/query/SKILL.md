---
name: "devwiki-query"
description: "当用户询问项目能力、功能说明、配置规则、现有实现位置、排障知识、对外说明、比较分析或已有 DevWiki 知识，并且不要求修改代码时使用。若用户要求开发、修改、修复、重构、补测试或提交代码，应使用 devwiki-code。"
argument-hint: "<问题>"
---

# /devwiki-query

## 核心约束（内联）

- **证据落地**：每个关键结论必须落到真实来源（wiki 页面、raw 文件或文档内已记录的代码线索）；`qmd` 只是召回加速器不是真相源；事实与推断必须拆开表达。
- **文档优先**：本 Skill 默认完整消费 DevWiki 文档后再回答。实现类问题优先读 workflow card → workflow core → workflow explain；不因为用户问“怎么实现”就自动搜索真实代码。
- **视图分层读取**：先用 `--view card` 判断命中，再用 `--view core` 回答主问题，只有 core 不够时读 `--view explain`。card 只放判断“是不是这个页面”的信息，core 放高频主结论，explain 放低频补充。
- **按需加载参考文档**：
  - 仅本地搜索低置信、需 qmd 升档时 → 读 `references/zatools-qmd.md`
  - 仅用户要求保存回答、沉淀结论、写入文件时 → 读 `references/mutation-safety.md`

不要凭空回答项目事实；先查 DevWiki 文档，优先基于 topic / workflow / troubleshooting / raw 总结；每个关键结论都要能追溯来源。

## 与 devwiki-code 的边界

- 本 Skill 是只读查询：回答知识、规则、文档中的实现说明、代码线索、影响面和排障线索。
- 如果用户表达了要修改代码、开发功能、修 bug、调整接口/配置/业务逻辑、重构、补测试或提交代码，本 Skill 只做简短转交：使用 `devwiki-code`。
- 如果用户明确要求查代码、核对当前实现、找具体文件/函数/行号、确认真实调用链或运行态，本 Skill 不直接搜索代码；先说明需要升级到 `devwiki-code` 做当前代码核查。
- 不要在修改类请求中继续执行 query 的 locate_code 全流程；修改类请求由 `devwiki-code` 走 `index → workflow card/core → 必要 topic core → 代码核对 → 测试 → 实现 → 验证`。
- 只有用户明确要求保存回答、沉淀结论或写入报告，且 `repo info <project>` 显示 `source.type=local` 时，本 Skill 才写 `wiki/outputs/`；remote source 默认只在对话中输出报告内容，不写本地文件。

## Outputs

- 先给出语义识别结果：explain_topic / locate_code / troubleshoot / public_answer / compare
- 命中的 Topic / Workflow / Troubleshooting 页面
- 知识缺口、冲突和待确认项
- 必要的代码定位线索、修改影响和测试建议，只使用 DevWiki 文档中已有信息；当前代码核查转交 `devwiki-code`
- 只有用户明确要求保存回答、沉淀结论、写入报告，且 `source.type=local` 时，才写入 `wiki/outputs/<query-slug>.md` 并追加 `wiki/log.md`；remote source 只输出报告正文和建议保存位置

## DevWiki Interaction

### Reads

**统一项目配置：**
- 先从当前代码仓 `AGENTS.md` / `CLAUDE.md` 的 DevWiki link block 读取 `DevWiki project`。
- 如果 link block 没有 project，执行 `zatools devwiki repo info`；该命令无参数时只输出已配置 project 名称 JSON 数组。只有一个 project 时直接使用；多个 project 时列出候选并让用户选择；没有配置时提示先执行 `zatools devwiki repo add <project> ...`。
- 使用 `zatools devwiki repo info <project>` 确认统一项目配置；输出默认 JSON，`source` 决定本地文档库或远端 HTTP API，`code_repos` 提供已绑定代码仓路径。
- 后续 `devwiki search/read` 默认使用 `--project <project>`。不要读取本地项目文件来推导文档库路径，也不要手工指定文档库根目录。

**结构化定位（串行降级）：**
- 第 1 层：`zatools devwiki search index <关键词...> --project <project>` → 命中则停，选最佳候选
- 第 2 层：index 无结果 → `zatools devwiki search glossary <关键词...> --project <project>`
- 第 3 层：glossary 无结果 → 按语义使用 `zatools devwiki search <topic|workflow> <关键词...> --project <project>`
- `index` 返回 `type`、`description`、`slug`；`glossary` 返回 `glossary`、`type`、`description`、`slug`；两者不返回 `score`
- 禁止并行搜多个源，禁止同时读多个候选的 card

**视图分层读取（禁止用 Read 工具直接读 topic/workflow 文件）：**
- `zatools devwiki read <topic|workflow> <slug> --view card --project <project>` — 判断命中
- `zatools devwiki read <topic|workflow> <slug> --view core --project <project>` — 回答主问题
- `zatools devwiki read <topic|workflow> <slug> --view explain --project <project>` — 补充细节
- `zatools devwiki repo info <project-slug>` 默认输出 JSON，同时返回 DevWiki `source` 和已绑定代码仓 `code_repos`；需要查看本地代码仓路径时读取 `code_repos[].path`。

**代码目录：**
- 本 Skill 默认不读取真实代码目录，也不执行 `rg` 代码搜索。
- 只有用户明确要求查代码、核对当前实现、找文件函数或行号、确认真实调用链/日志出处/运行态时，停止 query 的代码搜索动作，转交 `devwiki-code`。

### Writes

- 默认不写任何文件
- 只有用户明确要求保存回答或沉淀结论，且 `repo info <project>` 显示 `source.type=local` 时，才允许（需先读 `references/mutation-safety.md`）：
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`
- 如果 `source.type=remote`，不要创建本地 `wiki/outputs` 或追加本地 `wiki/log.md`；只输出报告正文，并说明当前远端项目未执行本地落盘。

## Query Principles

1. 不凭空回答项目事实。
2. 先查 Wiki 和 raw 来源；query 不自动核对真实代码。
3. 每个关键结论都有来源。
4. explain_topic 回答禁止直接展开代码、函数、文件路径、调用链、测试入口或修改方式。
5. locate_code 只进入 Workflow 文档层面；locate_code 默认回答 Workflow 文档中的实现入口、模块职责、状态流/数据流、副作用和文档中已记录的代码线索。
6. 如果文档不足以支撑当前代码事实，不要静默查代码；说明知识缺口，并建议改用 `devwiki-code` 做代码核查。
7. `zatools qmd` 只是召回工具，不是真相源。
8. 读取 view 时遵守知识经济学放置规则：先用 card 判断命中，再用 core 回答主问题，只有 core 不够时读取 explain。
9. 搜索和读取串行降级，不并行：`devwiki search index` → `devwiki search glossary` → `devwiki search topic/workflow`，候选逐个 card 验证，确认匹配才往下走。不为了“快”而并发读多个源或多个候选。

## 目录选择规则

| 用户意图 | 用户实际在问 | 优先目录 | 辅助目录 |
|---|---|---|---|
| explain_topic | 能力边界、功能规则、配置、状态、联动、流程规则 | `wiki/topics/` | `wiki/glossary.md`, `raw/` |
| locate_code | 文档记录的代码入口、调用链、接口、内部逻辑、实现机制、影响范围 | `wiki/workflows/` | `wiki/topics/`, `workflow explain` |
| troubleshoot | 报错、不生效、怎么排查、怎么修复 | `wiki/troubleshooting/` | `wiki/workflows/`, `wiki/topics/` |
| public_answer | 对外说明、客户口径、官网/文档口径 | public 可见 Topic | 不读代码 |

典型顺序：

- 能力/功能问题：`topics`
- 实现问题：`workflows → topics → workflow explain`
- 排障问题：`troubleshooting → workflows → topics`

## 去重与权威来源规则

- 能力边界、功能规则、参数、联动和设计流转以 topic 页面为准。
- 调用链、关键逻辑、代码定位、修改影响和测试入口以 workflow 页面为准。
- 日志、错误码、排障步骤以 troubleshooting 页面为准。
- 其他页面如果出现重复内容，只能作为摘要和导航，不作为权威来源。

## Workflow

### Step 1: 意图识别

1. 从当前代码仓 DevWiki link block 读取 project；找不到时执行 `zatools devwiki repo info` 发现候选，无法唯一确定时让用户选择。
2. 使用 `zatools devwiki repo info <project>` 确定统一项目配置；CLI 根据 `source` 自动选择本地或远端查询。
3. 如果用户要求修改代码、开发功能、修 bug、调整接口/配置/业务逻辑、重构、补测试或提交代码，停止本 Skill，转交 `devwiki-code`。
4. 判断问题语义：
   - `explain_topic`：用户想了解能力、功能、配置、边界、规则、联动。
   - `locate_code`：用户想基于 DevWiki 文档了解代码入口、调用链、接口、数据流、实现机制、影响面、测试入口，但不要求本轮修改代码，也没有明确要求核对当前代码仓。
   - `troubleshoot`：用户想解决报错、不生效、异常现象、诊断路径、修复建议。
   - `public_answer`：用户要求对外说明、客户口径、官网/文档口径。
   - `compare`：用户要求比较两个主题、方案或实现路径。

### Step 2: 串行定位候选页面

**原则：串行降级，逐层深入。每一层拿到结果后先判断，够用就停，不往下走。禁止并行搜多个源。**

从用户问题提取多关键词，按以下顺序串行搜索：

**第 1 层：`devwiki search index`**

```bash
zatools devwiki search index "<关键词1>" "<关键词2>" --project <project>
```

命中后根据 `description`、`type`、`slug` 选出**最匹配的一个**入口。**index 命中后，禁止再跑当前 Step 的任何后续层（glossary、topic/workflow search），直接使用返回的 `type` 和 `slug` 进入 Step 3。**

**第 2 层：index 无结果 → `devwiki search glossary`**

```bash
zatools devwiki search glossary "<关键词1>" "<关键词2>" --project <project>
```

命中后根据 `glossary` 和 `description` 判断是否能直接路由到返回的 `type` / `slug`。如果只是提供同义词，提取新关键词后回到第 1 层；如果能直接路由，进入 Step 3。

**第 3 层：glossary 无结果 → `devwiki search topic/workflow`**

根据语义选择 kind：

```bash
# explain_topic → topic
zatools devwiki search topic "<关键词1>" "<关键词2>" --project <project>

# locate_code / troubleshoot → workflow 优先
zatools devwiki search workflow "<关键词1>" "<关键词2>" --project <project>
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

1. 对 Step 2 选出的**单个**最佳候选，用 `zatools devwiki read <type> <slug> --view card --project <project>` 读取导航卡。
2. card 确认匹配 → 进入 Step 4 深度阅读。
3. card 不匹配 → 回到 Step 2 的结果中选**下一个**候选，再次 card 验证。不要并发读多个 card。
4. 所有候选都不匹配 → 回到 Step 2 的下一层继续搜索。

### Step 4: 按语义深度阅读

**所有 topic/workflow 页面读取必须使用 `zatools devwiki read`，禁止用 Read 工具直接读文件。**

按语义控制深度：

| 语义 | 阅读路径 | 止步条件 |
|---|---|---|
| explain_topic | topic card → topic core → (topic explain) | core 足够即停，不触发 workflow |
| locate_code | topic card → topic core → workflow card → workflow core → workflow explain | Workflow 文档足以回答实现机制、入口、影响面和验证方式即停，不进入真实代码搜索 |
| troubleshoot | workflow card → workflow core → (workflow explain) → topic core | 排障路径清晰即停；需要日志出处或运行态核实时转 `devwiki-code` |
| public_answer | topic card → topic core | core 足够即停，不读代码 |
| compare | 各对象 topic card → topic core | 差异清晰即停 |

每一层读取后判断：当前 view 是否足以回答用户问题？足够则停，不够才升到下一层。

### Step 5: 显式代码核查升级

本 Skill 不直接搜索真实代码。以下情况停止 query 的代码搜索动作，转交 `devwiki-code`：

- 只有用户明确要求查代码、核对当前实现、找文件函数或行号
- 用户要求确认真实调用链、日志出处、配置读取点、测试入口或运行态
- 用户要求修改建议且需要确认当前代码约束
- wiki / raw / workflow 证据互相冲突，或文档不足以支撑当前代码事实
- 排障问题必须确认真实日志、运行态或调用路径

交接回答应说明：

- 已基于 DevWiki 文档确认了哪些入口、模块职责、状态流/数据流或副作用
- 哪些结论仍需要当前代码核查
- 建议使用 `devwiki-code` 继续，并让 `devwiki-code` 按 workflow 锚点读取 `references/code-tracing.md` 后定向搜索代码

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

- Workflow 文档描述的实现入口
- 主要模块职责
- 状态流 / 数据流 / 副作用
- 文档中已记录的代码线索
- Topic 规则如何映射到实现
- 修改影响是什么（仅限文档已记录）
- 如何测试验证（仅限文档已记录）
- 如果没有查代码，明确说明：本轮基于 DevWiki 文档总结，未展开当前代码核查。

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

只有用户明确要求保存回答、沉淀结论、写入报告，且 `repo info <project>` 显示 `source.type=local` 时，才先读 `references/mutation-safety.md`，再创建 `wiki/outputs/<query-slug>.md` 并追加 `wiki/log.md`。如果 `source.type=remote`，只输出报告正文和建议保存位置，不写本地文件。
