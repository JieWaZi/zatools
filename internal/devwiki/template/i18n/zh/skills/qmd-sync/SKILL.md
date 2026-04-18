---
name: "devwiki-qmd-sync"
description: "当 DevWiki 工作区已经存在，但 `zatools qmd` collection 还未注册、注册结果可疑、索引落后，或需要把 `zatools qmd` 检索模式恢复到可用状态时使用。"
argument-hint: "[--root <devwiki-root>]"
---

# /devwiki-qmd-sync

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`


> 为已有 DevWiki 工作区补做或修复 `zatools qmd` collection 注册、索引刷新与状态检查。默认优先给 dry-run 与状态，再决定是否真正执行。

## Inputs

- `--root <devwiki-root>`：可选；默认当前目录
- `config/search.yaml`
- 当前环境对 `zatools qmd ...` 的本地执行能力

## Outputs

- `zatools qmd` collection dry-run 命令
- 可选：实际执行的 `zatools qmd` collection 注册结果
- 最新 `zatools qmd update` 结果
- 最新 `zatools qmd status` 摘要
- 若需要更高质量语义召回：是否建议继续执行 `embed`

## DevWiki Interaction

### Reads

- `config/search.yaml`
- `config/project.yaml`
- 本机 `zatools qmd` collection 与索引状态

### Writes

- 不直接写 wiki 页面
- 允许更新 `zatools qmd` 的 collection / index / embedding 状态


## Workflow

### Step 1: 检查前提

1. 确认当前目录或 `--root` 指向真实 DevWiki 根目录
2. 确认 `config/search.yaml` 存在
3. 确认当前环境可以执行 `zatools qmd status`
4. 若 `zatools qmd ...` 无法执行，直接进入 fallback：打印需要执行的命令，并明确说明当前无法启用 `zatools qmd` 检索模式

### Step 2: 先看 dry-run 命令

先执行：

```bash
zatools qmd sync --root <devwiki-root>
```

检查生成的 collection add 命令是否和当前工作区一致，再决定是否真正执行。

### Step 3: 按需注册或修复 collection

如果 collection 尚未注册、路径不对、或用户明确要求修复，则执行：

```bash
zatools qmd sync --root <devwiki-root> --apply
```

如果这里只需要核对，不要盲目执行。

### Step 4: 刷新索引并查看状态

collection 状态无误后，继续执行：

```bash
zatools qmd update
zatools qmd status
```

至少确认：

- DevWiki 对应的 collection 已存在
- `raw / wiki / code` 对应 collection 的文件数不是明显异常的 0
- `Pending`、`Updated` 等状态与当前工作区规模相符

### Step 5: 只在需要时补 embed

如果当前任务明确依赖 `zatools qmd query` 做更高质量的语义召回，且 `status` 显示存在待生成向量，再询问用户是否继续执行：

```bash
zatools qmd embed
```

默认不要在每次同步后都自动跑 `embed`。

### Step 6: 给出结论

结论至少包含：

- `zatools qmd` collection 是否已正确注册
- 索引是否已刷新
- 是否仍有 pending embeddings
- 当前后续技能是否可以进入 `zatools qmd` 检索模式
- 若仍不健康，下一步是修 config、重跑 sync，还是退回 fallback 模式

## Constraints

- **先 dry-run，后 apply**
- **不要把 `zatools qmd` 的命中结果当成事实**
- **collection 注册与索引刷新分开处理**
- **`embed` 默认按需执行，不是每次强制执行**
- **如果 `zatools qmd ...` 不可用，必须明确说明当前为 fallback 模式**

## Error Handling

- **`config/search.yaml` 缺失**：提示先运行 `/devwiki-setup` 或 `zatools devwiki init`
- **`zatools qmd ...` 无法执行**：停止 apply，打印应执行的命令并说明未启用 `zatools qmd` 检索模式
- **sync apply 失败**：报告失败，保留 dry-run 输出，不要假装 collection 已注册
- **status 显示文件数异常或 collection 不存在**：提示继续排查 collection 路径或重新执行 sync
