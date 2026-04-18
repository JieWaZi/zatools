# zatools qmd 使用约束

> 供所有通过 `zatools qmd ...` 做检索、同步和刷新操作的 DevWiki 技能共享使用。

统一使用 `zatools qmd ...` 执行检索与维护命令，不要直接调用其他命令。

在 Codex、Claude Code 等沙箱环境里，还要确认 agent 可以成功执行 `zatools qmd ...`，且项目根 `.cache` 目录可写；否则 `zatools qmd ...` 状态检查可能失败，此时要报告为“索引检查受限”。

规则：如果当前任务位于 DevWiki 工作区内，先读取 `config/search.yaml` 中的 `embed_model`、`rerank_model`、`generate_model`，然后在所有 `zatools qmd ...` 命令上显式追加 `--embed-model`、`--rerank-model`、`--generate-model`。若配置缺失，再回退到 CLI 内置默认值。

对于检索型任务，不要把 `zatools qmd status` 当成前置探测。直接执行目标 `zatools qmd ...` 检索命令；如果该命令失败，就把本次运行视为 `zatools qmd` 路径不可用，明确说明当前为降级检索，并回退到本地索引、文本搜索或文件搜索。

对于检索型任务，优先做双路召回：一路走 `zatools qmd ...`，一路走本地索引或直接文件 / 代码搜索。若当前 runtime 支持 delegation 或并行 agent，则优先并行执行两路并由主 agent 汇总；若不支持，则顺序执行，但保留同样的回退语义。
