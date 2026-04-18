---
name: "devwiki-ingest"
description: "当需要将一份或一批新的 raw 文档纳入 DevWiki，并决定它们应该落在哪些 documents、capabilities、changes、code refs 上时使用，尤其适用于增量补充知识和持续维护 wiki 的场景。"
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

- 新建或更新的 `wiki/documents/**/*.md`
- 新建或更新的 `wiki/capabilities/*.md`
- 新建或更新的 `wiki/changes/*.md`
- 写入或修正后的 `code refs`
- 更新后的 `wiki/index.md`
- 追加到 `wiki/log.md` 的 ingest 记录
- 一份给用户确认的 ingest proposal

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录
- `raw/*/*.md` — 待纳入的源文档
- `wiki/documents/**/*.md` — 查重、比对 source_path 与 source_hash
- `wiki/capabilities/*.md` — 匹配已有能力
- `wiki/changes/*.md` — 匹配已有变更
- `wiki/index.md` — 定位已有页面
- 本地代码目录 — 用于补充候选代码位置和局部二次排查

### Writes

- CREATE / EDIT `wiki/documents/**/*.md`
- CREATE / EDIT `wiki/capabilities/*.md`
- CREATE / EDIT `wiki/changes/*.md`
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: 解析来源并做去重检查

1. 展开 `source` 对应的文档列表
2. 对每份文档提取：
   - 标题
   - `doc_type`
   - `source_path`
   - `source_hash`
3. 用 `source_path`、`source_hash`、标题三种线索检查是否已被收录
4. 若完全相同且无内容变化，可在提案中标记为“无需更新”
5. 若 `source_path` 相同但 `source_hash` 变化，视为已有 document 的增量更新

### Step 2: 匹配已有 documents / capabilities / changes

1. 先匹配 document 自身是否已有镜像页
2. 再匹配最可能关联的 capability
3. 再匹配最可能关联的 change
4. 匹配时优先看：
   - 标题和别名
   - 已有关联文档
   - capability 的 `code_refs`
   - change 的 `change_classification`
5. 如果多个能力相似，不要硬选一个；先保留多候选

### Step 3: 召回候选代码线索

1. 若配置了代码目录：
   - 先执行 `zatools qmd status`
   - 若 `zatools qmd status` 正常，优先用 `zatools qmd query` 在 `raw / wiki / code` 三层召回，再用 `zatools qmd get` / `zatools qmd multi-get` 读取 top-K 结果
   - 再在代码目录做本地关键词搜索
2. 召回后对 top-K 候选文件做局部二次排查
3. 至少确认：
   - 文件为何相关
   - 关键函数或符号是否存在
   - 它更像主引用还是辅助 clue
4. 如果代码命中过散、无法定位到具体文件或函数，不要写高置信 `code refs`

### Step 4: 形成 ingest proposal

提案至少要回答：

- 是否新建 document
- 是否更新已有 document
- 是否挂到已有 capability
- 是否新建 capability
- 是否更新已有 change
- 是否新建 change
- 是否写入或修正 code refs

提案中必须把 action 分成风险层级：

- 低风险：新建 document 镜像页、刷新 `source_hash`、追加日志
- 中风险：挂到已有 capability、补充辅助 code refs、追加已有 change 关联
- 高风险：新建 capability、新建 change、修改主 code refs、改变 change 分类

### Step 5: 低置信度时向用户提问

自主检索几轮后，如果仍不清楚，必须向用户提问。典型场景：

- 一个文档同时命中多个 capability
- 多个变更记录都可能相关
- 文档提到的接口、模块、函数在代码中找不到
- 代码只命中一堆散点，没有形成可解释的聚类

提问要求：

- 一次只问 1 到 3 个最关键问题
- 优先追问入口锚点：接口 URL、关键文件、关键函数、页面路径、需求单号
- 不要泛泛地说“请提供更多上下文”

### Step 6: 等待用户确认

1. 所有中高风险写入都要等待确认
2. 尤其是以下动作不能默认执行：
   - 新建 capability
   - 新建 change
   - 把文档改挂到另一个 capability
   - 写入主 code refs
3. 只有提案得到确认后，才允许落盘

### Step 7: 落盘并刷新导航

确认后：

1. 写入或更新 `wiki/documents/`
2. 写入或更新 `wiki/capabilities/`
3. 写入或更新 `wiki/changes/`
4. 更新 `wiki/index.md`
5. 在 `wiki/log.md` 追加 `ingest | proposal-applied`
6. 落盘成功后执行：

```bash
zatools qmd update
zatools qmd status
```

7. 如果后续任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

必要时在回答末尾建议下一步：

- 需要变更定性：建议 `/devwiki-scope`
- 文档不够但代码有实现：建议 `/devwiki-feature-doc`
- 已有知识和代码不一致：建议 `/devwiki-refresh`

## Constraints

- **raw/ 只读**：不得修改源文档
- **不得虚构 code refs**：没有看过代码的文件和函数不能写成高置信引用
- **不得偷偷合并 capability**：能力归属不清时必须保留待确认
- **中高风险必须确认**：尤其是 capability、change、主 code refs 的写入和改挂
- **source_hash 必须更新**：文档内容变化后必须刷新
- **代码线索不是设计文档**：`ingest` 只补局部代码证据，不替代 `/devwiki-feature-doc`

## Error Handling

- **source 为空或路径不存在**：要求用户提供有效的 raw 文档路径
- **文档类型无法判断**：提示用户先整理到合适目录，或补充转换步骤
- **`zatools qmd ...` 不可用**：回退到本地搜索，不中断 ingest
- **代码目录未配置**：允许只基于文档更新 wiki，但必须注明未核对代码
- **多个 capability / change 都很像**：停止扩散，向用户提问
