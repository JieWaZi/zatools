---
name: "devwiki-init"
description: "当需要基于现有 raw 文档为单产品仓库启动第一版 DevWiki 知识骨架时使用，尤其适用于还没有 capability 和 feature 结构化页面，或刚完成 setup 需要第一次建立 wiki 的场景。"
argument-hint: "[范围说明]"
---

# /devwiki-init

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 从 `raw/` 启动第一版 DevWiki。目标不是一次写满所有知识，而是先建立可信的 capabilities / features 骨架，并给出 `init proposal` 等待确认。

## Inputs

- `scope`（可选）：初始化范围说明，例如“用户管理相关文档”或“全部 raw 文档”
- `raw/*/*.md` — 原始需求、设计、功能说明、测试方案
- `config/project.yaml` — 主代码目录、语言与代码仓配置

## Outputs

- `wiki/capabilities/*.md` — 初始化阶段识别出的能力页
- `wiki/features/*.md` — 初始化阶段识别出的功能页
- `wiki/index.md` — 更新后的目录
- `wiki/log.md` — `init` 提案与落盘日志
- 一份给用户确认的 `init proposal`

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录与语言配置
- `raw/*/*.md` — 作为事实来源
- `wiki/index.md` — 若已经存在，用于避免重复初始化
- `wiki/capabilities/*.md` — 若已有能力页，用于避免重复创建
- `wiki/features/*.md` — 若已有功能页，用于避免重复创建
- 本地代码目录 — 若配置存在，用于做第一轮轻量功能入口扫描

### Writes

- CREATE `wiki/capabilities/*.md`
- CREATE `wiki/features/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: 检查初始化前提

1. 确认当前目录是 DevWiki 根目录，且已经执行过 `setup`
2. 检查 `raw/` 是否至少包含一类 Markdown 文档
3. 如果 `wiki/capabilities/` 与 `wiki/features/` 已有较多内容，先判断这是不是首次初始化
4. 若看起来已经初始化过，不要直接覆盖，应先向用户说明现状并建议改用 `/devwiki-ingest` 或 `/devwiki-refresh`

### Step 2: 扫描 raw 文档并建立 feature 候选

1. 遍历 `raw/*/*.md`
2. 对每份文档提取：
   - 标题
   - 可能关联的功能主题
   - 从父目录推断出的来源类型
   - `source_path`
   - `source_hash`
3. `source_hash` 必须基于当前文件内容重新计算，不能沿用旧值
4. 对于没有一级标题的文档，使用文件名作为降级标题，但要在提案里标记为低质量来源
5. 将相关 raw 资料归并成 feature 候选，目标路径落在 `wiki/features/<feature-slug>.md`
6. 不再创建 document 镜像页；raw 文件继续保留在来源层

### Step 3: 归并 capability 候选

1. 从标题、章节、小节名、重复业务术语里抽取候选能力
2. capability 视角必须业务优先：回答系统有什么能力，而不是怎么实现
3. 如果多个 raw 文档都在支撑同一能力，优先合并成一个 capability
4. 将 capability 候选与 feature 候选建立映射
5. 若能力边界不清，保留为待确认项，不要硬定

### Step 4: 做第一轮代码线索扫描

按 `references/zatools-qmd.md` 的阶梯召回规则执行，**默认 local-first**：

1. 读取 `config/project.yaml` 中的主代码目录
2. 若代码目录存在：
   - 先用本地 `grep` / 文件搜索定位关键锚点
   - 命中不足再升档到 `zatools qmd search`
   - 初始化阶段尽量不使用 `zatools qmd query`；如确有概念级召回需求且本地加速可用再考虑
3. 仅做轻量扫描，目标是给 feature 候选补充初始 `code_refs`、`api_entries` 与 `test_refs`
4. 初始化阶段不要深挖完整调用链；深挖属于 `/devwiki-ask`、`/devwiki-feature-doc`
5. 对初始代码线索必须标注置信度，低置信命中不能写成主引用

### Step 5: 生成 init proposal

`init proposal` 至少包含：

- 将新建哪些 capability
- 将新建哪些 feature
- 哪些 raw 文档被映射到哪些 feature
- 哪些 feature 支撑哪些 capability
- 哪些 code refs 只是线索，哪些已经较高置信
- 哪些点需要用户确认

同时按风险分层：

- 低风险：追加日志、刷新索引、记录 source hash
- 中风险：将 feature 挂到已有 capability、追加辅助 code clue
- 高风险：新建 capability、新建 feature、写入主 code refs

### Step 6: 等待用户确认

1. 所有中高风险动作都必须等待用户确认
2. 如果提案里有多个能力边界不清、或代码命中过散，就不要直接写入
3. 这类情况下先向用户提 1 到 3 个具体问题，再修正提案
4. 没有确认前，不要落盘 capability 和 feature

### Step 7: 落盘并更新导航

在用户确认后：

1. 写入 `wiki/capabilities/`
2. 写入 `wiki/features/`
3. 更新 `wiki/index.md`
4. 在 `wiki/log.md` 追加 `init | proposal-applied`
5. 落盘成功后执行：

```bash
zatools qmd update
zatools qmd status
```

6. 如果当前任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

## Constraints

- **raw/ 只读**：不得修改 `raw/` 下文件
- **不得虚构**：没有 raw 或代码证据支持的 capability / feature 不能硬写
- **source_hash 必须真实**：必须反映当前源文档内容
- **中高风险必须确认**：尤其是新建 capability、新建 feature、写入高置信 code refs
- **初始化不等于完全消化**：`init` 只做第一版骨架，不代替后续 `/devwiki-ingest`
- **代码扫描要克制**：初始化阶段做轻扫，不做长链路深挖

## Error Handling

- **raw 为空**：提示用户先准备原始文档，再执行 `/devwiki-init`
- **已有大量 wiki 内容**：提示用户可能已经初始化过，建议改用 `/devwiki-ingest` 或 `/devwiki-refresh`
- **代码目录未配置**：允许仅基于 raw 初始化，但要明确说明未完成代码关联
- **`zatools qmd ...` 不可用**：回退到本地关键词搜索，不中断初始化
- **多个 capability 候选冲突**：停止扩散搜索，转而向用户提问
