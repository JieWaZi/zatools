---
name: "devwiki-reset"
description: "当需要按受控 scope 重置生成态 DevWiki 内容时使用，例如 wiki、raw、log 或 checkpoints。"
argument-hint: "[scope list]"
---

# /devwiki-reset

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`


> 在受控 scope 内重置 DevWiki 生成内容。默认先预览；破坏性执行必须显式确认。

## Inputs

- `scope` — 逗号分隔的 scope：`wiki`、`raw`、`log`、`checkpoints` 或 `all`
- 可选 `--dry-run` — 仅预览
- 可选 `--yes` — 确认执行 reset

## Outputs

- reset plan
- 可选 applied reset result
- 追加到 `wiki/log.md` 的 reset 日志

## DevWiki Interaction

### Reads

- `wiki/capabilities/`
- `wiki/features/`
- `wiki/outputs/`
- `wiki/graph/`
- `wiki/index.md`
- `wiki/log.md`
- `raw/*/`
- `wiki/.checkpoints/`

### Writes

- DELETE `wiki/capabilities/`、`wiki/features/`、`wiki/outputs/`、`wiki/graph/` 下命中的生成文件
- DELETE 选定 `raw/` 子目录下的文件
- RESET `wiki/index.md`
- 只有在选择 `log` scope 时才 RESET `wiki/log.md`


## Workflow

### Step 1: 生成 reset 计划

1. 将 `all` 展开为 `wiki,raw,log,checkpoints`
2. 收集候选删除项
3. 永远不要删除 `.gitkeep`
4. 缺失路径按 no-op 处理，不视为失败

### Step 2: 展示计划

预览中必须列出：

- 选定的 scopes
- 待删除目标
- 待重置目标
- 当前是否 dry-run

### Step 3: 确认破坏性执行

1. 如果传了 `--dry-run`，不得写入任何内容
2. 如果会删除文件，且未传 `--yes`，就停在预览
3. 只有拿到显式确认后，才允许执行 reset

### Step 4: 应用 reset

确认后：

1. 删除计划中的文件
2. 若选择了 `wiki` scope，则用基础模板重写 `wiki/index.md`
3. 若选择了 `log` scope，则用基础模板重写 `wiki/log.md`
4. 如果 log 仍然存在，尽量在其中追加一条带日期的 reset 摘要

## Constraints

- **默认先预览**：不要隐式执行删除
- **不得删除 `.gitkeep`**：目录骨架必须稳定
- **scope 必须有边界**：不要越过所选 scope 扩散
- **raw reset 是破坏性的**：删除来源资料前必须确认

## Error Handling

- **未知 scope**：报告合法 scope 并停止
- **未提供 scope**：要求用户给出明确 scope
- **路径本来就不存在**：按 no-op 处理
