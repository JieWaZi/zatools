# zatools devwiki 使用约束

> 供所有需要定位、读取或维护 DevWiki 项目知识的技能共享使用。

## 统一项目配置

1. 先从当前代码仓 `AGENTS.md` / `CLAUDE.md` 的 DevWiki link block 读取 `DevWiki project`。
2. 如果 link block 没有 project，执行 `zatools devwiki repo info`。该命令无参数时只输出已配置 project 名称 JSON 数组。
3. 只有一个 project 时直接使用；多个 project 时列出候选并让用户选择；没有配置时提示先执行 `zatools devwiki repo add <project> ...`。
4. 使用 `zatools devwiki repo info <project>` 确认统一项目配置。输出默认 JSON，`source` 决定本地文档库或远端 HTTP API，`code_repos` 提供已绑定代码仓路径。
5. 后续 `zatools devwiki search/read` 默认使用 `--project <project>`。不要读取本地项目文件来推导文档库路径，也不要手工指定文档库根目录。

## 结构化入口优先，低置信升档

这里的“结构化入口优先”指 DevWiki 的 index/glossary 表格，不是代码仓全局搜索。

基础顺序：

```text
意图识别 → devwiki search index → devwiki search glossary → devwiki search topic/workflow → qmd query → raw/code 核对
```

任一阶段拿到足够强的 top-K 且置信足够即停；不要为了“保险”无边界扩大搜索。

### 意图识别

| 意图类型 | 典型问题 | 默认策略 |
|---|---|---|
| `locate_exact` | 文件在哪里、哪个函数定义、哪个接口注册、错误码在哪里 | 本地 Wiki / 必要时本地代码精确定位 |
| `explain_topic` | 某功能是什么、怎么工作、有哪些边界 | `devwiki search index/glossary`；低置信或噪声大时升到 `devwiki search topic` |
| `trace_implementation` | 怎么实现、调用链怎么走、状态写到哪里 | query 场景先读 Wiki/workflow；明确要求当前代码核查时转 `devwiki-code` |
| `troubleshoot` | 报错原因、不生效怎么查、日志从哪里来 | 先找 workflow/topic 候选；troubleshooting 可作为页面内容和导航线索，再按需核对 raw/code |
| `design_intent` | 为什么这么设计、整体架构是什么 | `devwiki search index/glossary`；不足时 `devwiki search`，再不足才 `qmd query` |
| `wiki_maintenance` | 页面是否重复、过期、冲突、query 是否会命中旧内容 | 本地 Wiki 审计 + `qmd search/status/update` 按需验证 |

不要把所有关键词都当成精确锚点。`ssh`、`vip`、`auth`、`token`、`query`、`sync` 这类短词只是中锚点；如果用户问“怎么实现 / 怎么设计 / 怎么排障”，不能因为本地命中这些短词就停止。

### 第 1 档：Index / Glossary

默认先检索 DevWiki 结构化入口：

```bash
zatools devwiki search index <query...> --project <project>
zatools devwiki search glossary <query...> --project <project>
```

`index` 返回 `type`、`description`、`slug`；`glossary` 返回 `glossary`、`type`、`description`、`slug`。这两个命令是结构化表格检索；在 `--project` 下由 CLI 根据统一配置选择本地文档库或远端 HTTP API，不依赖 qmd，也不输出 `score`。agent 必须根据 `description` 和用户问题做语义打分。

结构化搜索置信判断：

| 置信等级 | 判断标准 | 后续动作 |
|---|---|---|
| high | index/glossary 命中 1-5 个入口；`type` 与意图一致；`description` 明确覆盖用户问题 | 直接读命中页，不必升档 |
| medium | 命中 6-20 条；有 2-4 个候选入口；需要读 card 后判断主页面 | 先读候选页，仍无法排序则升档 |
| low | 0 命中；超过 20 条散点命中；短词命中过泛；active/deprecated/report 混杂；页面冲突；无法判断权威页 | 必须升到 `zatools devwiki search <topic|workflow> <query...> --project <project>` |

### 第 2 档：Topic / Workflow Search

当 index/glossary 低置信、噪声过大、无法排序，或问题本身偏语义/主题类时，使用 `devwiki search topic/workflow`：

```bash
zatools devwiki search topic <query...> --project <project>
zatools devwiki search workflow <query...> --project <project>
zatools devwiki search workflow 防脑裂 网关 ha-group gateway --project <project>
```

- 多个关键词应作为多个参数传入，不要合并成一个带空格的字符串。
- 多 query 结果使用 RRF（Reciprocal Rank Fusion）按各关键词下的排名融合排序，`score` 为融合后的相对百分比。
- `devwiki search topic/workflow` 底层调用 `qmd search`，不依赖向量，CPU 友好，适合作为结构化入口搜索的升档。
- `devwiki search` 命中只是候选排序，最终结论必须回到真实 `wiki/`、`raw/` 或已核对代码文件。

### 第 3 档：qmd Query

只有 `devwiki search` 和结构化入口搜索仍不足，且问题本身是概念、设计、意图、跨页面关系类时，才读取 `references/zatools-qmd.md` 并升到 `zatools qmd query`。

### qmd 失败 fallback

qmd 报错、超时、collection 未注册或 cache 不可写时，说明：

```text
本轮 qmd 不可用，已降级为 DevWiki 结构化入口搜索；结论只基于本轮可读证据。
```

raw/code 仍需本地核对；qmd 不是真相源。

## 视图分层读取

读取 Topic / Workflow 时先用 card 判断，再按需读取 core/explain：

```bash
zatools devwiki read <topic|workflow> <slug> --view card --project <project>
zatools devwiki read <topic|workflow> <slug> --view core --project <project>
zatools devwiki read <topic|workflow> <slug> --view explain --project <project>
```

- `slug` 必须使用 search 结果中的 `slug` 字段，不要用文件名。
- 候选逐个 card 验证，确认匹配才读取 core/explain；不要并行读多个候选的 card。
- 只读 query 类 skill 必须用 `zatools devwiki read`，禁止直接读取 topic/workflow 文件。
- 写入类 skill 已确认需要修改本地 Wiki 文件时，可以读取目标本地 Markdown 文件；因为目标就是修改该文件。remote source 不执行本地写入。

## Query Principles

1. 不凭空回答项目事实。
2. 先查 Wiki 和 raw 来源；query 不自动核对真实代码。
3. 每个关键结论都有来源。
4. explain_topic 回答禁止直接展开代码、函数、文件路径、调用链、测试入口或修改方式。
5. locate_code 只进入 Workflow 文档层面；locate_code 默认回答 Workflow 文档中的实现入口、模块职责、状态流/数据流、副作用和文档中已记录的代码线索。
6. 如果文档不足以支撑当前代码事实，不要静默查代码；说明知识缺口，并建议改用 `devwiki-code` 做代码核查。
7. `zatools qmd` 只是召回工具，不是真相源。
8. 读取 view 时遵守知识经济学放置规则：先用 card 判断命中，再用 core 回答主问题，只有 core 不够时读取 explain。
9. 搜索和读取串行降级，不并行：`devwiki search index` → `devwiki search glossary` → `devwiki search topic/workflow`，候选逐个 card 验证，确认匹配才往下走。

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
