---
name: "devwiki-ask"
description: "当需要基于 DevWiki 已有 capabilities、feature 页面、原始文档和代码线索回答问题，或者在研发动作前收敛上下文、判断需求是 new 还是 modify 时使用。"
argument-hint: "<问题或变更描述>"
---

# /devwiki-ask

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> DevWiki 的统一问答与变更定性入口。默认只读，不写 wiki；根据用户意图决定是否附加变更定性环节。

## Inputs

- `question`：自然语言输入。可以是查询问题，也可以是变更描述
- `--format`（可选）：输出格式，默认 `markdown`，可选 `table` / `bullets` / `timeline`
- `--save-output`（可选）：仅当用户明确要求把答案沉淀到 `wiki/outputs/` 时使用

## Outputs

- **始终**：一份带引用的综合回答
- **当意图为开发/变更**：附带 `classification: new / modify / unclear` + 候选 feature / 候选代码 top-K + 下一步建议
- **可选**：若用户明确要求保存，则写入 `wiki/outputs/<query-slug>.md`
- **附带**：相关 capabilities、相关 features、代码位置、已知缺口

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录、语言和代码仓配置
- `wiki/index.md` — 定位候选页面
- `wiki/capabilities/*.md` — 能力聚合页
- `wiki/features/*.md` — 功能说明页，包含 source trace、入口、代码线索、测试入口
- `wiki/outputs/*.md` — 历史问答沉淀
- `raw/*/*.md` — 当 wiki 摘要不够时回源查看原始文档
- 本地代码目录 — 仅当问题涉及实现现实、文件/函数归属、接口路径，或 wiki/raw 证据不足时再核对

### Writes

- 默认不写任何内容
- 只有用户明确要求沉淀答案时，才允许：
  - CREATE `wiki/outputs/<query-slug>.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: 意图识别与范围收敛

1. 读取 `config/project.yaml`，确定主代码目录
2. 从用户输入里判断意图类型：
   - **查询/问答类**
     - 常见语气：「以前怎么设计」「哪个 capability 负责」「哪个 feature 支撑」「哪个文件负责」「当前实现是不是这样」
     - 走 Step 2 → wiki/raw 已足够则 Step 4，否则 Step 3 → Step 4 → 结束
   - **开发/变更类**
     - 常见语气：「要改…」「要新增…」「这个需求怎么做」「改造…」「加一个…」「把…迁移成…」
     - 走 Step 2 → 需要实现核对时 Step 3 → Step 5（变更定性）→ Step 4 → 结束
3. 边界模糊时默认按查询类处理，并在结论处提醒用户如果需要做变更定性，可以继续要求输出 `new / modify / unclear`

### Step 2: 召回候选资料

按 `references/zatools-qmd.md` 的阶梯召回规则执行：

1. **默认先用本地 `grep` / 文件搜索**定位已知锚点（符号、文件、接口 URL、capability slug、feature slug 等）
2. 命中不足再升档 `zatools qmd search`，在 `raw / wiki / code` 三类 collection 上做关键词召回
3. 仍不足且问题属于概念/设计/意图类，才考虑升档 `zatools qmd query`；无 GPU/加速时按共享规则走硬性 fallback
4. 候选数量受控：top-K（K ≤ 12），优先读高相关 capability 页和 feature 页；wiki 不够时再回源 `raw/`
5. 先判断 wiki/raw 是否已经闭环：
   - 如果文档已经足够回答，就不要为了“更稳”再默认展开代码阅读
   - 如果页面已经足够回答，就不要为了“更稳”再默认展开代码阅读
   - 只有当页面证据不足、问题明确要求实现核对，或你要做开发/变更定性时，才进入 Step 3

### Step 3: 按需核对代码（仅在必要时）

只有当 wiki/raw 证据不足、问题明确要求实现核对，或你要做开发/变更定性时，才进入 Step 3。

只有当文档证据不足、问题明确要求实现核对，或你要做开发/变更定性时，才进入 Step 3。

以下情况必须核对本地代码，不能只根据 wiki 或 raw 回答：

- 「当前实现是不是这样」
- 「哪个文件 / 哪个函数负责」
- 「某接口现在怎么走」
- 「最近改动可能影响哪些代码」
- 开发/变更类里，相关 feature 页与已有 `code_refs` 仍不足以支撑定性或下一步建议时

核对时：

1. 优先读取相关 feature 页里已有的 `code_refs`、`api_entries`、`test_refs`
2. 若这些线索不够，再在代码目录中做定向搜索（第 1 档 / 第 2 档为主）
3. 至少确认入口文件、关键函数、接口注册点或关键调用边界中的一层证据
4. 如果代码与文档冲突，要明确指出「代码现状」和「文档描述」分别是什么
5. 如果只是补充开发/变更线索，优先停在 feature 层入口锚点；不要默认把每个候选文件都读深

### Step 4: 组织回答

回答必须包含以下部分：

1. **直接答案**：先给结论，不要先堆检索过程
2. **证据引用**：每个关键论断都附带来源路径
   - wiki 引用：`wiki/capabilities/...`、`wiki/features/...`
   - raw 引用：`raw/...`
   - 代码引用：文件路径，必要时补符号名
3. **区分层级**：
   - 已确认事实
   - 合理推断
   - 尚未确认 / 缺口
4. **相关项列表**：
   - 相关 capabilities
   - 相关 features
   - 相关 code refs（仅当本轮真的核对了代码，或现有 `code_refs` 本身就是证据时列出）
5. **下一步建议**：
   - feature 页缺失或过旧：建议 `/devwiki-feature-doc`
   - wiki 漂移或索引失效：建议 `/devwiki-refresh`
   - 缺少原始资料：建议 `/devwiki-ingest`
   - 需要健康检查：建议 `/devwiki-check`
6. **若当前回答主要基于 wiki/raw**：在结尾补一句「如需，我可以再基于代码做一次核对版汇总」
7. **若已走 Step 5**：把变更定性结果并入回答开头，紧跟直接答案

### Step 5: 变更定性（仅当意图为开发/变更）

把 Step 2 / Step 3 的证据归并成三个桶，再做定性：

1. **三类证据桶**
   - 已有 capability
   - 已有 feature
   - raw 来源资料
2. **候选 feature / 候选代码 top-K**
   - 先用 capability 页、feature 页、已有 `code_refs`、`api_entries`、`test_refs` 组织 top-K
   - 只有需要核对实现现实或入口归属时，才补看文件内容
   - 说明该 feature 或文件为什么相关
   - 若已核对代码，再说明关键函数 / 类 / symbol 是否存在
3. **给出 classification**
   - `modify`：已有 capability 或 feature 明显命中
   - `new`：未命中已有 capability / feature，但目标清晰
   - `unclear`：证据冲突、过散，或缺少关键锚点
4. **下一步建议**
   - `new / modify` 清晰：可直接进入设计或 `/devwiki-feature-doc`
   - `unclear`：要么给出追问（Step 6），要么建议 `/devwiki-refresh`、`/devwiki-ingest` 补齐前置

### Step 6: 低置信度时追问用户

如果自主检索几轮后仍然不清楚，必须 ask the user，而不是硬答。典型场景：

- 找到多个候选 capability / feature，无法判断哪个才是问题主体
- 用户提到的接口、文件、函数、路由在本地都找不到
- 代码依赖动态注册、配置下发、网关转发或外部服务，无法仅凭当前仓库确认
- 问题范围过大，例如「权限体系怎么实现的」，但实际上跨多个子能力
- 开发/变更类问题里缺少任何入口锚点

提问要求：

- 一次只问 1 到 3 个最关键问题
- 问题要具体，不要泛泛地说「请补充更多上下文」
- 优先追问入口锚点：接口 URL、关键文件、关键函数、页面路径、已知 feature 名、已知 capability 名

### Step 7: 按需沉淀答案

仅当用户明确要求保存问答结果时：

1. 将回答整理为 `wiki/outputs/<query-slug>.md`
2. 写清原始问题、引用来源、结论、未确认项；若包含变更定性，也要落到输出里
3. 在 `wiki/log.md` 追加一条 `ask | <question-summary> | saved-output`
4. 若回答暴露 wiki 事实错误，不要偷偷改现有页面，应建议或执行 `/devwiki-refresh`

## Constraints

- **不得虚构**：所有关键结论必须来自 DevWiki 实际内容或本地代码
- **raw/ 只读**：不得修改 `raw/` 下文件
- **默认不写 wiki**：除非用户明确要求保存输出
- **wiki/raw 足够时就停**：如果 `wiki/` 和 `raw/` 已能回答且问题不关心实现现实，不要默认继续读代码
- **代码问题必须看代码**：凡是涉及实现行为、路径、函数职责的问题，都必须核对代码
- **事实与推断分离**：classification 是推断，页面/文件/symbol 存在是事实
- **检索必须有边界**：几轮无果后要停下来向用户追问
- **引用必须真实存在**：不要引用不存在的页面、文件或符号
- **不要创建 change 页面**：`new / modify / unclear` 只存在于回答期

## Error Handling

- **wiki 基本为空**：提示用户先执行 `/devwiki-init` 或 `/devwiki-ingest`
- **缺少代码目录配置**：先用 `wiki/` 与 `raw/` 回答，并明确说明未核对代码
- **`zatools qmd ...` 不可用**：按 `references/zatools-qmd.md` 的 fallback 规则走，只用本地搜索
- **`zatools qmd query` 环境不支持**：停在 `zatools qmd search` + 本地搜索，不要静默卡住
- **无相关结果**：坦诚说明当前无足够证据，并建议 `/devwiki-ingest`、`/devwiki-feature-doc` 或给出更具体锚点
- **用户要求保存但证据不足**：先补问或建议 `/devwiki-refresh` / `/devwiki-feature-doc`，不要把低置信度内容直接沉淀
