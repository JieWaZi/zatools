---
name: "devwiki-reset"
description: "当需要按范围清理 DevWiki 生成内容、修复失败初始化后的残留状态，或为重新执行 init / ingest 准备干净工作区时使用。"
argument-hint: "--scope wiki|raw|log|checkpoints|all [--project-root <devwiki-root>]"
---

# /devwiki-reset

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/mutation-safety.md`

> 按范围重置 DevWiki。`/devwiki-reset` 是破坏性技能，必须先给 dry-run 计划，再等待用户明确确认。

## Inputs

- `--scope`：必填，可为 `wiki`、`raw`、`log`、`checkpoints`、`all`
- `--project-root`：可选；若当前目录就是 DevWiki 工作区，可直接用 `.`
- 当前工作区中的 `raw/`、`wiki/`

## Outputs

- 一份 dry-run 删除 / 重置计划
- 用户确认后的执行结果摘要
- 可选：写入 `wiki/log.md` 的 reset 记录

## DevWiki Interaction

### Reads

- `raw/`
- `wiki/documents/`
- `wiki/capabilities/`
- `wiki/changes/`
- `wiki/outputs/`
- `wiki/graph/`
- `wiki/.checkpoints/`

### Writes

- DELETE `raw/` 下命中 scope 的文件
- DELETE `wiki/documents/`、`wiki/capabilities/`、`wiki/changes/`、`wiki/outputs/`、`wiki/graph/` 下命中的生成文件
- RESET `wiki/index.md`
- 可选 RESET `wiki/log.md`

## Workflow

### Step 1: 规范化 scope 并做 dry-run

先把 scope 规范化，再执行：

```bash
zatools devwiki tool reset --scope <scope> --project-root <devwiki-root>
```

该命令只输出计划，不执行删除。必须把 `delete` 与 `reset` 分开向用户展示。

### Step 2: 向用户解释风险

解释每种 scope 的含义：

- `wiki`：清空生成知识页与输出，保留目录骨架
- `raw`：清空原始资料，风险最高
- `log`：重置 `wiki/log.md`
- `checkpoints`：清理中间状态
- `all`：以上全部

如果 scope 包含 `raw` 或 `all`，必须明确说明 `raw/` 删除通常不可恢复。

### Step 3: 等待明确确认

确认前禁止真正执行。应明确说明：

```text
即将按 scope=<scope> 删除 N 个文件、重置 M 个文件。请明确确认后继续。
```

只有在用户明确同意后，才允许进入下一步。

### Step 4: 执行重置

用户确认后执行：

```bash
zatools devwiki tool reset --scope <scope> --project-root <devwiki-root> --yes
```

读取返回结果并确认：

- 实际删除了哪些文件
- 实际重置了哪些文件
- 删除数与重置数是否符合预期

### Step 5: 记录 reset 日志

如果本次 scope 不包含 `log`，可追加一条低风险日志：

```bash
zatools devwiki tool log --wiki-root <devwiki-root>/wiki --message "reset | scope=<scope>"
```

### Step 6: 给出下一步建议

根据重置范围给建议：

- 刚清空 `raw/`：提醒先补原始资料
- 只清空 `wiki`：建议 `/devwiki-init` 或 `/devwiki-ingest`
- 只清空 `checkpoints`：建议继续之前的 ingest / refresh 流程

常见下一步：

- `/devwiki-init`
- `/devwiki-ingest`
- `/devwiki-setup`

## Constraints

- **必须先 dry-run，再确认，再执行**：不得跳过计划展示
- **没有明确确认，不得带 `--yes`**
- **保留骨架文件**：`.gitkeep` 不删
- **不要误删安装状态**：不得触碰当前项目根下的 `.agents/` 与 `.zatools-lock.json`
- **`raw/` 是高风险区**：scope 含 `raw` 时必须二次提醒
- **重置结果必须可解释**：删除和重置分别报告，不能只说“已清空”

## Error Handling

- **scope 缺失或非法**：列出合法值并停止，不要猜测
- **`zatools devwiki tool reset` 失败**：直接报错，禁止用临时命令代替批量删文件
- **`wiki/` 或 `raw/` 缺失**：按现状生成空计划并解释原因
- **日志追加失败**：报告 warning，但不要掩盖 reset 主结果
- **用户取消确认**：明确说明本次只完成 dry-run，没有真正修改任何文件
