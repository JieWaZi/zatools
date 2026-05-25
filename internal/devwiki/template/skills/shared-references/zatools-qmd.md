# zatools qmd 使用约束

> 供所有在 DevWiki 工作区内做检索、同步和刷新操作的技能共享使用。

本文件同时规定两件事：
1. **检索路由**：如何在本地 Wiki 搜索、`zatools qmd search`、`zatools qmd query`、raw/code 核对之间选择
2. **调用约束**：真正使用 `zatools qmd ...` 时必须遵守的模型参数与环境要求

## 检索路由（本地 Wiki 优先，低置信升档）

DevWiki 技能做召回时，默认先用本地命令检索 DevWiki 文档层；这里的“本地优先”首先指 `wiki/`，不是代码仓全局搜索。

基础顺序：

```text
意图识别 → 本地 Wiki 搜索 → 命中质量判断 → qmd search → qmd query → raw/code 核对
```

任一阶段拿到足够强的 top-K 且置信足够即停；不要为了“保险”无边界扩大搜索。

### 第 0 步：意图识别

先判断用户问题属于哪一类：

| 意图类型 | 典型问题 | 默认策略 |
|---|---|---|
| `locate_exact` | 文件在哪里、哪个函数定义、哪个接口注册、错误码在哪里 | 本地 Wiki / 必要时本地代码精确定位 |
| `explain_topic` | 某功能是什么、怎么工作、有哪些边界 | 本地 Wiki 搜索；低置信或噪声大时升到 `qmd search` |
| `trace_implementation` | 怎么实现、调用链怎么走、状态写到哪里 | 先找 Wiki/workflow 候选，再做代码核对 |
| `troubleshoot` | 报错原因、不生效怎么查、日志从哪里来 | 先找 troubleshooting/workflow 候选，再按需核对 raw/code |
| `design_intent` | 为什么这么设计、整体架构是什么、概念关系是什么 | 本地 Wiki 搜索；不足时 `qmd search`，再不足才 `qmd query` |
| `wiki_maintenance` | 页面是否重复、过期、冲突、query 是否会命中旧内容 | 本地 Wiki 审计 + `qmd search/status/update` 按需验证 |
| `qmd_maintenance` | qmd 不可用、collection 没注册、索引异常 | 交给 `devwiki-qmd-sync` |

不要把所有关键词都当成精确锚点。`ssh`、`vip`、`auth`、`token`、`query`、`sync` 这类短词只是中锚点；如果用户问“怎么实现 / 怎么设计 / 怎么排障”，不能因为本地命中这些短词就停止。

### 第 1 档：本地 Wiki 搜索（默认起点）

默认先检索 DevWiki 文档层：

```text
wiki/index.md
wiki/glossary.md
wiki/topics/
wiki/workflows/
wiki/troubleshooting/
```

必要时再扩展到 `raw/`。只有当问题明确要求实现现实、代码入口、调用链、日志出处、配置读取点、测试入口，或需要写入/修正 代码定位 时，才进入代码搜索。

本地 Wiki 搜索置信判断：

| 置信等级 | 判断标准 | 后续动作 |
|---|---|---|
| high | 命中 1-5 个页面；有明确权威页；标题、summary 或正文多处匹配；目录与意图一致 | 直接读命中页，不必升档 |
| medium | 命中 6-20 条；有 2-4 个候选页面；需要读页面后判断主页面 | 先读候选页，仍无法排序则升档 |
| low | 0 命中；超过 20 条散点命中；主要命中 index/glossary；短词命中过泛；active/deprecated/report 混杂；页面冲突；无法判断权威页 | 必须升到 `zatools qmd search` |

### 第 2 档：`zatools qmd search`（噪声收敛和关键词召回）

当本地 Wiki 搜索低置信、噪声过大、无法排序，或问题本身偏语义/主题类时，使用 `qmd search`：

- 典型场景：「ssh 是怎么实现的」「SAML metadata 相关设计」「鉴权失败日志出处」「和支付回调相关的文档」
- 默认只召回 `wiki` collection；只有用户手动在 `config/search.yaml` 添加 raw 或 code collection 后，才会覆盖这些额外目录
- `qmd search` 不依赖向量，CPU 友好，适合作为本地 Wiki 搜索的升档
- `qmd search` 命中只是候选排序，最终结论必须回到真实 `wiki/`、`raw/` 或已核对代码文件

### 第 3 档：`zatools qmd query`（语义召回）

仅当 `qmd search` 和本地 Wiki 搜索仍不足，且问题本身是**概念 / 设计 / 意图 / 跨页面关系**类时才启用：

- 典型场景：「权限体系整体怎么设计的」「这一块当初为什么这么做」「和限流相关的整体架构是什么」
- 依赖 embedding + rerank，可能受模型、GPU、cache 和索引状态影响
- 作为语义召回，不作为事实来源
- 必须设置合理等待边界；失败或过慢时降级，不要静默卡住

### qmd 失败 fallback

对检索型任务，不要先用 `zatools qmd status` 探测；直接执行目标 `search` 或 `query`，失败后再降级。

| 失败场景 | fallback |
|---|---|
| `qmd search` 报错、超时、命令不存在、cache 不可写 | 降级为本地 Wiki 搜索 + 必要 raw 搜索 |
| `qmd query` 报错、超时、模型缺失、加速不可用 | 先降级到 `qmd search`；若 search 也失败，再本地 Wiki 搜索 |
| collection 未注册或文件数异常 | 本轮本地 Wiki 搜索，并建议 `devwiki-qmd-sync` / `zatools qmd sync --root . --apply` |
| index 明显过期 | 本轮可本地查，结尾建议 `zatools qmd update` |
| raw/code 未进入 collection | 明示 qmd 默认只搜 `wiki`，raw/code 仍需本地核对 |

降级后必须在回答中说明：

```text
本轮 qmd 不可用，已降级为本地 Wiki 搜索；结论只基于本轮可读证据。
```

几轮仍然无果时，停止扩散搜索，向用户追问 1-3 个具体锚点，不要继续盲目升档。

## 调用约束

### 统一入口

- 所有检索与维护命令都走 `zatools qmd ...`，不直接调用其他实现
- 在 Codex、Claude Code 等沙箱 agent 中执行前，先确认：
  - agent 对 `zatools qmd ...` 有执行权限
  - 项目根 `.cache/` 目录可写

二者任一不满足时，当前运行视为「索引检查受限」，按上文 fallback 处理。

### 模型参数必须显式传

如果当前任务位于 DevWiki 工作区内：

1. 先读 `config/search.yaml` 中的 `embed_model`、`rerank_model`、`generate_model`
2. 在所有 `zatools qmd ...` 命令上显式追加 `--embed-model`、`--rerank-model`、`--generate-model`
3. 若某项缺失，再回退到 CLI 内置默认值

不要依赖执行目录恰好是 DevWiki 根目录。

### 不做的事

- 不把 `zatools qmd status` 当成前置探测。对检索型任务，直接执行目标检索命令；失败才视为不可用
- 不要把 `zatools qmd` 的命中当成事实，它只是召回加速器
- 不要在每次同步后都自动 `embed`，按需触发
