# zatools devwiki 使用约束

> 供所有需要定位、读取或维护 DevWiki 项目知识的技能共享使用。

## 统一项目配置

1. 先从当前代码仓 `AGENTS.md` / `CLAUDE.md` 的 DevWiki link block 读取 `DevWiki project`。
2. 如果 link block 没有 project，执行 `zatools devwiki repo info`。该命令无参数时只输出已配置 project 名称 JSON 数组。
3. 只有一个 project 时直接使用；多个 project 时列出候选并让用户选择；没有配置时提示先执行 `zatools devwiki repo add <project> ...`。
4. 使用 `zatools devwiki repo info <project>` 确认统一项目配置。输出默认 JSON，`active_source` 是当前激活来源类型，决定使用 `sources.local` 还是 `sources.remote`；`sources` 保存该 project 已配置的 local/remote 来源；`code_repos` 提供已绑定代码仓路径。
5. 如果同一 project 已同时配置本地和远端，使用 `zatools devwiki repo use <project> <local|remote>` 切换当前来源，不要通过重复 `repo add` 覆盖另一边配置。
6. 后续 `zatools devwiki search/read` 默认使用 `--project <project>`。不要读取本地项目文件来推导文档库路径，也不要手工指定文档库根目录。

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
| `trace_implementation` | 怎么实现、调用链怎么走、状态写到哪里 | query 场景先读 Wiki/workflow；缺少代码锚点且明确要求当前代码核查时，建议显式使用 `$devwiki-code` |
| `troubleshoot` | 报错原因、不生效怎么查、日志从哪里来 | 先找 workflow/topic 候选；troubleshooting 可作为页面内容和导航线索，再按需核对 raw/code |
| `design_intent` | 为什么这么设计、整体架构是什么 | `devwiki search index/glossary`；不足时 `devwiki search`，再不足才 `qmd query` |
| `wiki_maintenance` | 页面是否重复、过期、冲突、query 是否会命中旧内容 | 本地 Wiki 审计 + `qmd search/status/update` 按需验证 |

不要把所有关键词都当成精确锚点。`ssh`、`vip`、`auth`、`token`、`query`、`sync` 这类短词只是中锚点；如果用户问“怎么实现 / 怎么设计 / 怎么排障”，不能因为本地命中这些短词就停止。

## 用户确认门与自我反思

只要通过 index/glossary/topic/workflow/qmd search 找到候选文档，且准备把其中某个候选作为后续读取、代码定位、实现或维护方向，必须先停在用户确认门。不要直接读取 core/explain，不要进入 code-tracing，不要开始修改代码或文档。

确认前先做一次自我反思：

- 用户问题是否只是短词或宽泛领域词，导致关键词命中但语义不一定匹配；
- 候选的 `type`、标题、`description` 或 card 是否真的覆盖用户意图；
- active / deprecated / report / proposal / troubleshooting 是否混杂，是否可能把过程报告当事实页；
- 候选是否只因为同名词、配置项或接口路径命中，但业务场景、模块边界或时间状态不一致；
- 是否没有可靠候选，或命中结果显示文档可能不存在、过期、冲突。

确认问题必须直接、具体，使用当前 Agent 环境的结构化用户确认工具询问用户是否对当前查询结果满意，
- Codex 使用 askuserquestion / request_user_input
- Cursor 使用 AskQuestion / ask questions tool
- Claude 使用 AskUserQuestion
如果当前环境没有结构化提问工具，必须以普通聊天问题停止并等待用户回复，不得继续执行后续流程。提问时列出最多 3 个候选，并说明每个候选的：

- `type`、`slug` / 标题；
- 为什么看起来匹配；
- 支撑证据来自 search description 还是 card；
- 当前不确定点。

候选摘要必须合并 search 结果和 card 信息：`type/slug/title/score/description` 来自搜索结果，`status/summary/confidence/适合回答/不适合回答` 来自 card。card 信息不足以说明“为什么匹配”或“哪里不确定”时，不得把候选视为 high confidence，只能以 medium/low confidence 向用户确认。

提问不能空泛，不要只问“这个对吗”。应明确询问：“我准备沿候选 A 继续，是否对当前查询结果满意？如果不是，请指出更接近的功能名、接口、模块或页面。”

用户明确确认满意后，才能继续读取已确认候选的 core/explain，并进入后续回答、代码核对、实现或维护流程。用户未确认、用户不满意、表示不确定、指出方向不对，或候选证据不足时，必须停止沿当前候选推进，改为请用户补充锚点、调整关键词或重新搜索。没有可靠候选时，不要编造页面或静默继续；明确反馈“本轮没有找到可靠匹配文档”，再向用户确认下一步。

没有可靠候选时，不要把分散命中综合成新主题、新能力或推荐口径；只允许说明未找到依据、列出检索过的入口和请求用户补充锚点。尤其是 explain_topic 场景，缺少独立 Topic 或用户确认的候选时，不能用多个页面里的同词命中拼出能力边界。


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
| high | index/glossary 命中 1-5 个入口；`type` 与意图一致；`description` 明确覆盖用户问题 | 进入用户确认门；用户确认满意后才读命中页 |
| medium | 命中 6-20 条；有 2-4 个候选入口；需要读 card 后判断主页面 | 必要时只读 card 帮助排序，然后进入用户确认门；仍无法排序则升档 |
| low | 0 命中；超过 20 条散点命中；短词命中过泛；active/deprecated/report 混杂；页面冲突；无法判断权威页 | 先升档到 `zatools devwiki search <topic|workflow> <query...> --project <project>`；升档后找到候选再进入用户确认门；仍无可靠候选时再向用户反馈并请求补充锚点 |


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
6. 如果文档不足以支撑当前代码事实，不要静默查代码；说明知识缺口。缺少代码锚点且需要 DevWiki 定位入口时，建议显式使用 `$devwiki-code` 做代码核查；已有明确锚点时，按普通代码查看处理。
7. `zatools qmd` 只是召回工具，不是真相源。
8. 读取 view 时遵守知识经济学放置规则：先用 card 判断命中，再用 core 回答主问题，只有 core 不够时读取 explain。
9. 搜索和读取串行降级，不并行：`devwiki search index` → `devwiki search glossary` → `devwiki search topic/workflow`，候选逐个 card 验证，确认匹配才往下走。
10. 没有可靠候选、没有独立 Topic/Workflow 或用户未确认候选时，不组织推断性答案；只输出“当前 Project Brain 没有足够信息支持该结论。”、检索过的入口和可补充的锚点。

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
