---
name: "devwiki-ask"
description: "当需要查询 DevWiki 中已有能力、历史变更、原始文档和代码线索，并回答“以前怎么设计”“关联哪些文档”“哪些代码最相关”“当前实现行为是什么”之类问题时使用。"
argument-hint: "<问题>"
---

# /devwiki-ask

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 对 DevWiki 做通用问答。默认只回答，不写入 wiki；只有用户明确要求沉淀结果时，才考虑写入 `wiki/outputs/`。

## Inputs

- `question`：自然语言问题
- `--format`（可选）：输出格式，默认 `markdown`，可选 `table` / `bullets` / `timeline`
- `--save-output`（可选）：仅当用户明确要求把答案沉淀到 `wiki/outputs/` 时使用

## Outputs

- **始终**：一份带引用的综合回答
- **可选**：若用户明确要求保存，则写入 `wiki/outputs/<query-slug>.md`
- **附带**：相关能力、相关变更、相关文档、相关代码位置、已知缺口、下一步建议

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录、语言和代码仓配置
- `wiki/index.md` — 定位候选页面
- `wiki/documents/**/*.md` — 查需求、设计、特性说明、代码总结、接口文档、测试方案、复盘
- `wiki/capabilities/*.md` — 查能力聚合页
- `wiki/changes/*.md` — 查变更记录
- `wiki/outputs/*.md` — 若历史问答已经沉淀，可复用其结论但仍要核对引用
- `raw/*/*.md` — 当 wiki 摘要不够时，回源查看原始文档
- 本地代码目录 — 当问题涉及代码行为、接口实现、文件位置、关键函数时必须核对代码

### Writes

- 默认不写任何内容
- 只有用户明确要求沉淀答案时，才允许：
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: 建立检索范围

1. 读取 `config/project.yaml`，确定主代码目录
2. 识别问题类型：
   - 历史设计类
   - 文档关联类
   - 能力归属类
   - 变更影响类
   - 代码定位类
   - 实现行为类
3. 若问题明显是在做研发前置判断，优先提醒用户这类问题也可能适合 `/devwiki-scope`

### Step 2: 召回候选资料

1. 先读 `wiki/index.md`，按关键词和能力名召回候选 slug
2. 在 `wiki/documents/`、`wiki/capabilities/`、`wiki/changes/` 中优先挑选高相关页面
3. 检索方式遵循 `references/zatools-qmd.md`
4. 控制候选数量，优先读取 top-K（K ≤ 12），避免上下文失控
5. 如果 wiki 页面只给了摘要，再回源读取对应 `raw/` 文档

### Step 3: 定向核对代码

以下问题必须核对本地代码，不能只根据 wiki 或 raw 回答：

- “当前实现是不是这样”
- “哪个文件 / 哪个函数负责”
- “某接口现在怎么走”
- “最近改动可能影响哪些代码”

核对时：

1. 优先读取 capability / change 中已有 `code_refs`
2. 若 `code_refs` 不够，再在代码目录中做定向搜索
3. 至少确认入口文件、关键函数或关键调用链中的一层证据
4. 如果代码与文档冲突，要明确指出“代码现状”和“文档描述”分别是什么

### Step 4: 组织回答

回答必须包含以下部分：

1. **直接答案**：先给结论，不要先堆检索过程
2. **证据引用**：每个关键论断都附带来源路径
   - wiki 引用：`wiki/capabilities/...`、`wiki/changes/...`、`wiki/documents/...`
   - raw 引用：`raw/...`
   - 代码引用：文件路径，必要时补符号名
3. **区分层级**：
   - 已确认事实
   - 合理推断
   - 尚未确认 / 缺口
4. **相关项列表**：
   - 相关 capabilities
   - 相关 changes
   - 相关 documents
   - 相关 code refs
5. **下一步建议**：
   - 需要变更定性：建议 `/devwiki-scope`
   - 文档缺失但代码存在：建议 `/devwiki-feature-doc`
   - wiki 漂移或索引失效：建议 `/devwiki-refresh`
   - 缺少原始资料：建议 `/devwiki-ingest`
   - 需要健康检查：建议 `/devwiki-check`

### Step 5: 低置信度时追问用户

如果自主检索几轮后仍然不清楚，必须 ask the user，而不是硬答。如果置信度仍低，向用户提 1 到 3 个具体问题。典型场景：

- 找到多个候选 capability / change，无法判断哪个才是问题主体
- 用户提到的接口、文件、函数、路由在本地都找不到
- 代码依赖动态注册、配置下发、网关转发或外部服务，无法仅凭当前仓库确认
- 问题范围过大，例如“权限体系怎么实现的”，但实际上跨多个子能力

提问要求：

- 一次只问 1 到 3 个最关键问题
- 问题要具体，不要泛泛地说“请补充更多上下文”
- 优先追问入口锚点：接口 URL、关键文件、关键函数、页面路径、最近变更单

### Step 6: 按需沉淀答案

仅当用户明确要求保存问答结果时：

1. 将回答整理为 `wiki/outputs/<query-slug>.md`
2. 写清原始问题、引用来源、结论、未确认项
3. 在 `wiki/log.md` 追加一条 `ask | <question-summary> | saved-output`
4. 若回答暴露 wiki 事实错误，不要偷偷改现有页面，应建议或执行 `/devwiki-refresh`

## Constraints

- **不得虚构**：所有关键结论必须来自 DevWiki 实际内容或本地代码
- **raw/ 只读**：不得修改 `raw/` 下文件
- **默认不写 wiki**：除非用户明确要求保存输出
- **代码问题必须看代码**：凡是涉及实现行为、路径、函数职责的问题，都必须核对代码
- **事实与推断分离**：不要把推断写成事实
- **允许承认不知道**：证据不足时要明确写“当前 DevWiki 中证据不足”
- **引用必须真实存在**：不要引用不存在的页面、文件或符号
- **候选数量受控**：优先 top-K，不要无边界乱搜

## Error Handling

- **wiki 基本为空**：提示用户先执行 `/devwiki-init` 或 `/devwiki-ingest`
- **缺少代码目录配置**：先用 `wiki/` 与 `raw/` 回答，并明确说明未核对代码
- **`zatools qmd ...` 不可用**：回退到本地文本搜索和目录排查
- **无相关结果**：坦诚说明当前无足够证据，并建议 `/devwiki-ingest`、`/devwiki-feature-doc` 或给出更具体锚点
- **用户要求保存但证据不足**：先补问或建议 `/devwiki-refresh` / `/devwiki-feature-doc`，不要把低置信度内容直接沉淀
