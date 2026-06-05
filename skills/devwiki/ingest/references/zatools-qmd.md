# zatools qmd 使用约束

> `zatools qmd` 只负责检索层、collection 和索引维护。DevWiki 项目选择、结构化入口搜索和 read 视图统一看 `references/zatools-devwiki.md`。

## 适用范围

使用本文件的场景：

- 需要执行 `zatools qmd sync` 注册 collection；
- 需要执行 `zatools qmd update` 刷新索引；
- 需要执行 `zatools qmd status` 检查索引状态；
- 需要执行 `zatools qmd search` 或 `zatools qmd query` 做低置信升档召回；
- 需要解释 qmd 失败 fallback、模型参数、cache 或 collection 限制。

不要把本文件当作 DevWiki 项目路由文档；不要在这里维护 `zatools devwiki repo/search/read` 的通用规则。

## 常用命令

```bash
zatools qmd sync --root . --apply
zatools qmd update
zatools qmd status
zatools qmd search <query...>
zatools qmd query <question>
```

- `sync` 只同步 collection 注册，通常根据 `config/search.yaml` 生成或执行 `qmd collection add ...`；不等同于 update、embed 或 query。
- `update` 刷新已注册 collection 的索引。
- `status` 检查 qmd-first readiness 和索引状态。
- `search` / `query` 只是召回工具，命中内容不是事实源。

## 模型参数必须显式传

如果当前任务位于 DevWiki 工作区内：

1. 先读 `config/search.yaml` 中的 `embed_model`、`rerank_model`、`generate_model`。
2. 在所有需要模型的 `zatools qmd ...` 命令上显式追加 `--embed-model`、`--rerank-model`、`--generate-model`。
3. 若某项缺失，再回退到 CLI 内置默认值。

不要依赖执行目录恰好是 DevWiki 根目录。

## qmd 失败 fallback

对检索型任务，不要先用 `zatools qmd status` 探测；直接执行目标 `search` 或 `query`，失败后再降级。

| 失败场景 | fallback |
|---|---|
| `qmd search` 报错、超时、命令不存在、cache 不可写 | 降级为 `zatools devwiki search index/glossary` + 必要 raw 搜索 |
| `qmd query` 报错、超时、模型缺失、加速不可用 | 先降级到 `zatools devwiki search topic/workflow`；若 search 也失败，再 `index/glossary` |
| collection 未注册或文件数异常 | 本轮使用 DevWiki 结构化入口搜索，并建议在本地 DevWiki 工作区执行 `zatools qmd sync --root . --apply` |
| index 明显过期 | 本轮可本地查，结尾建议 `zatools qmd update` |
| raw/code 未进入 collection | 明示 qmd 默认只搜 `wiki`，raw/code 仍需本地核对 |

降级后必须说明：

```text
本轮 qmd 不可用，已降级为 DevWiki 结构化入口搜索；结论只基于本轮可读证据。
```

## 不做的事

- 不把 `zatools qmd status` 当成检索前置探测。
- 不要把 `zatools qmd` 的命中当成事实，它只是召回加速器。
- 不要在每次同步后都自动 `embed`，按需触发。
- 不在本文件维护 DevWiki repo、search、read 或页面路由规则。
