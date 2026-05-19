# zatools qmd 使用约束

> 供所有在 DevWiki 工作区内做检索、同步和刷新操作的技能共享使用。

本文件同时规定两件事：
1. **检索路由**：如何在本地搜索、`zatools qmd search`、`zatools qmd query` 之间做阶梯选择
2. **调用约束**：真正使用 `zatools qmd ...` 时必须遵守的模型参数与环境要求

## 检索路由（阶梯召回）

DevWiki 技能做召回时，默认按「成本 / 速度由低到高」升档，任一档命中 top-K 且置信足够即停，不必继续升档。

### 第 1 档：本地 `grep` / 文件搜索（默认起点）

当问题里已经包含具体锚点时，**优先本地搜索**，不走 `zatools qmd ...`：

- 已知符号名、函数名、类名
- 已知文件名、目录名、路径片段
- 接口 URL、路由、错误码
- capability slug、ticket 号、commit hash

此档通常可以直接给出答案，尤其在没有 GPU 的机器上，速度与确定性都最好。

### 第 2 档：`zatools qmd search`（关键词召回）

当问题有关键词但不知道具体落点时用 `qmd search`：

- 典型场景：「SAML metadata 相关代码」「鉴权失败日志出处」「和支付回调相关的文档」
- 默认只召回 `wiki` collection；只有用户手动在 `config/search.yaml` 添加 raw 或 code collection 后，才会覆盖这些额外目录
- 不依赖向量，CPU 友好

### 第 3 档：`zatools qmd query`（语义召回）

仅当前两档都召回不足，且问题本身是**概念 / 设计 / 意图**类时才启用：

- 典型场景：「权限体系整体怎么设计的」「这一块当初为什么这么做」「和限流相关的整体架构是什么」
- 依赖 embedding + rerank，无 GPU 时会明显变慢
- 作为上层召回，不作为第一选择

### 升档与停止

- 任一档拿到足够强的 top-K 就停止升档，不做无意义的语义兜底
- 跨档使用时，前一档的命中应作为下一档的锚点（例如第 1 档命中的文件路径可以作为 `qmd search` 的缩窄范围）
- 几轮仍然无果时，**停止扩散搜索，向用户追问 1–3 个具体锚点**，不要继续盲目升档

### 硬性 fallback

执行 `zatools qmd query` 前必须先判断本地是否有可用加速：

- 若未检测到 GPU / 加速器，或确认当前环境只能在 CPU 上跑 embed / rerank，直接报告「当前无可用 GPU/加速，降级为 `qmd search` + 本地文本搜索」，跳过 `qmd query`
- 不允许静默卡在 `qmd query` 上等待
- 若该档失败（命令报错、超时、cache 不可写等），视为 `zatools qmd` 路径本次运行不可用，退回到第 1 / 第 2 档并在答复里明示「降级检索」

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
