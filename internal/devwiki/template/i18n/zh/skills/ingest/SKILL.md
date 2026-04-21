---
name: "devwiki-ingest"
description: "当需要将一份或一批新的 raw 文档纳入 DevWiki，并决定它们应该如何更新 capabilities、features 与 feature 级代码线索时使用。"
argument-hint: "<文档路径或范围>"
---

# /devwiki-ingest

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 将新的 raw 文档纳入 DevWiki 提案流。`ingest` 负责增量吸收资料、匹配已有知识、补初步代码线索，并在确认后更新 wiki。

## Inputs

- `source`：一个文档路径、一个目录，或一批待纳入的 raw 文档
- `raw/*/*.md` — 新增或更新的原始文档
- `config/project.yaml` — 主代码目录、语言与代码仓配置

## Outputs

- 新建或更新的 `wiki/capabilities/*.md`
- 新建或更新的 `wiki/features/*.md`
- 写入或修正后的 `sources`、`code_refs`、`api_entries`、`test_refs`
- 更新后的 `wiki/index.md`
- 追加到 `wiki/log.md` 的 ingest 记录
- 一份给用户确认的 ingest proposal

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录
- `raw/*/*.md` — 待纳入的源文档
- `wiki/features/*.md` — 用 `sources.path` 与 `sources.hash` 查重
- `wiki/capabilities/*.md` — 匹配已有能力
- `wiki/index.md` — 定位已有页面
- 本地代码目录 — 用于补充候选代码位置和局部二次排查

### Writes

- CREATE / EDIT `wiki/capabilities/*.md`
- CREATE / EDIT `wiki/features/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: 解析来源并做去重检查

1. 展开 `source` 对应的文档列表
2. 对每份文档提取：
   - 标题
   - 可能关联的功能主题
   - 从父目录推断出的来源类型
   - `source_path`
   - `source_hash`
3. 用三种线索检查是否已被收录到现有 feature：
   - `sources.path`
   - `sources.hash`
   - 标题 / 功能名相似度
4. 若 `sources.path` 相同但 hash 变化，视为已有 feature 的增量更新
5. 若 raw 文件是新的，但明显属于已有 feature，应优先更新该 feature，而不是强行新建

### Step 2: 匹配已有 capabilities / features

1. 先匹配最可能关联的 feature 页
2. 再匹配最可能支撑它的 capability 页
3. 匹配时优先看：
   - 标题和别名
   - 已有 capability-feature 关系
   - feature 的 `code_refs`、`api_entries`、`test_refs`
   - raw 文档里的重复业务术语
4. 如果多个能力都像，不要硬选一个；先保留多候选
5. 如果 raw 来源明显描述了一个新功能，则把“新建 feature”保留为提案

### Step 3: 召回候选代码线索

按 `references/zatools-qmd.md` 的阶梯召回规则执行，**默认 local-first**：

1. 若配置了代码目录：
   - 先用本地 `grep` / 文件搜索定位已知锚点（符号、文件、目录、接口 URL）
   - 命中不足再升档到 `zatools qmd search` 做关键词召回
   - 仅当前两档都不足、且需要概念级召回时才考虑 `zatools qmd query`；无 GPU/加速时按共享规则走硬性 fallback
2. 召回后对 top-K 候选文件做局部二次排查
3. 至少确认：
   - 文件为何相关
   - 关键函数或符号是否存在
   - 它更像主引用还是辅助 clue
4. 如果代码命中过散、无法定位到具体文件或函数，不要写高置信 `code refs`

### Step 4: 形成 ingest proposal

提案至少要回答：

- 是否更新已有 feature
- 是否新建 feature
- 是否挂到已有 capability
- 是否新建 capability
- 是否写入或修正 feature 级代码线索

提案中必须把 action 分成风险层级：

- 低风险：刷新 source hash、追加日志、刷新索引
- 中风险：挂到已有 capability、补充辅助 code refs
- 高风险：新建 capability、新建 feature、修改主 code refs

### Step 5: 低置信度时向用户提问

自主检索几轮后，如果仍不清楚，必须向用户提问。典型场景：

- 一个文档同时命中多个 capability
- 多个 feature 页都可能相关
- 文档提到的接口、模块、函数在代码中找不到
- 代码只命中一堆散点，没有形成可解释的入口

提问要求：

- 一次只问 1 到 3 个最关键问题
- 优先追问入口锚点：接口 URL、关键文件、关键函数、页面路径、需求单号
- 不要泛泛地说「请提供更多上下文」

### Step 6: 等待用户确认

1. 所有中高风险写入都要等待确认
2. 尤其是以下动作不能默认执行：
   - 新建 capability
   - 新建 feature
   - 把 feature 改挂到另一个 capability
   - 写入主 code refs
3. 只有提案得到确认后，才允许落盘

### Step 7: 落盘并刷新导航

确认后：

1. 写入或更新 `wiki/capabilities/`
2. 写入或更新 `wiki/features/`
3. 更新 `wiki/index.md`
4. 在 `wiki/log.md` 追加 `ingest | proposal-applied`
5. 落盘成功后执行：

```bash
zatools qmd update
zatools qmd status
```

6. 如果后续任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

必要时在回答末尾建议下一步：

- 需要变更定性或进一步问答：建议 `/devwiki-ask`
- raw 文档不足但代码里功能明确：建议 `/devwiki-feature-doc`
- 已有知识和新证据不一致：建议 `/devwiki-refresh`

## Constraints

- **raw/ 只读**：不得修改源文档
- **不得虚构 code refs**：没有看过代码的文件和函数不能写成高置信引用
- **不得偷偷重划 capability 边界**：归属不清时必须保留待确认
- **中高风险必须确认**：尤其是 capability、feature、主 code refs 的写入和改挂
- **source hash 必须更新**：文档内容变化后必须刷新到 feature 页
- **ingest 不是完整设计流程**：`ingest` 只补知识和局部代码证据，不替代 `/devwiki-feature-doc`

## Error Handling

- **source 为空或路径不存在**：要求用户提供有效的 raw 文档路径
- **文档类型无法判断**：提示用户先整理到合适目录
- **`zatools qmd ...` 不可用**：回退到本地搜索，不中断 ingest
- **代码目录未配置**：允许只基于 raw 更新 wiki，但必须注明未核对代码
- **多个 capability 候选都很像**：停止扩散，向用户提问
