---
name: "devwiki-scope"
description: "当研发动作开始前需要先收敛上下文、判断需求更像 new 还是 modify、定位相关历史文档与候选代码位置，并在低置信时向用户提出具体问题时使用。"
argument-hint: "<变更描述>"
---

# /devwiki-scope

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 在真正写设计、改代码、补文档之前，先做一次面向研发的变更定性与上下文收敛。`/devwiki-scope` 默认只读，不直接改写 wiki。

## Inputs

- `query`：本次变更描述，至少应包含一个明确目标
- 可选锚点：
  - capability / 模块名
  - 接口 URL
  - 关键文件
  - 关键函数 / 类
  - 页面路径 / 路由
  - 需求单号
- `config/project.yaml`：读取主代码目录

## Outputs

- 一份 scope 报告
- `classification`：`new`、`modify`、`unclear`
- 关联的 documents / capabilities / changes
- top-K 候选代码文件与符号
- 若置信度不足，给用户的 1 到 3 个具体问题

## DevWiki Interaction

### Reads

- `config/project.yaml`
- `wiki/documents/`
- `wiki/capabilities/`
- `wiki/changes/`
- `wiki/index.md`
- `raw/`
- 本地代码目录

### Writes

- 默认不写任何页面
- 仅在用户明确要求留档时，允许低风险输出到 `wiki/outputs/`


## Workflow

### Step 1: 收敛问题边界

先把用户问题压缩成一个可搜索的范围。至少确认：

- 这次要改什么
- 更偏业务能力还是系统能力
- 是否已经有入口锚点

如果用户只说“看下这个需求怎么做”，但没有任何明确功能名或锚点，应先追问 1 到 3 个最关键问题，而不是直接全库乱搜。

### Step 2: 用三层集合做首次召回

检索方式遵循 `references/zatools-qmd.md`，覆盖三类 collection：

- `raw`
- `wiki`
- `code`

检索目标包括：

- `wiki/documents/` 下相关需求、设计、特性说明、代码总结、复盘
- `wiki/capabilities/` 下相关能力页
- `wiki/changes/` 下历史 change
- `raw/` 中尚未完全结构化但可能有价值的原始资料
- 代码目录中的 top-K 候选文件

### Step 3: 归并结构化证据

把首次召回结果整理成三个证据桶：

- 已有 capability
- 历史 change
- 文档与原始资料

归并时要说明每条证据为什么相关，而不是只列文件名。

### Step 4: 推断 `new / modify / unclear`

判定时至少考虑：

- 是否命中已有 capability
- 是否命中历史设计 / 需求 / 特性说明
- 是否命中已有 change
- 是否在代码里找到可解释的已有实现

建议规则：

- `modify`：已有 capability、历史文档、代码实现明显命中
- `new`：未命中已有能力与实现，但有清晰新增目标
- `unclear`：证据冲突、过散，或缺少关键锚点

### Step 5: 对 top-K 代码候选做本地精查

不要停留在“文件名像”这一层。至少要核对：

- 文件为什么相关
- 关键函数 / 类 / symbol 是否存在
- 它是主实现、调用入口，还是仅仅辅助线索
- 是否还有更靠近入口的文件

`/devwiki-scope` 需要二次排查代码，但不要像 `/devwiki-feature-doc` 那样展开整条调用链。

### Step 6: 输出 scope 报告

scope 报告至少包含：

- 本次变更更像 `new`、`modify` 还是 `unclear`
- 相关 capability、change、document、raw 资料
- top-K 代码候选及理由
- 当前已知缺口
- 下一步建议：
  - 只是一般问答：建议 `/devwiki-ask`
  - 要系统梳理实现：建议 `/devwiki-feature-doc`
  - 已有知识明显漂移：建议 `/devwiki-refresh`

### Step 7: 低置信时向用户提问

在经过几轮有边界的检索后，如果仍然低置信，必须停止扩散搜索，转而向用户提 1 到 3 个具体问题。优先追问：

- 接口 URL
- 关键文件
- 关键函数
- 页面路径
- 已知 capability 名称

## Constraints

- **`zatools qmd ...` 是召回加速器，不是真相源**
- **事实与推断分离**：页面存在、文件存在、symbol 存在是事实；`new / modify / unclear` 是推断
- **默认只读**：`/devwiki-scope` 不应顺手改 wiki
- **搜索必须有边界**：几轮无果后要停下来提问
- **代码候选必须复核**：不能把纯关键词命中当成结论
- **不要虚构 capability 或实现关系**

## Error Handling

- **wiki 基本为空**：提示先执行 `/devwiki-init` 或 `/devwiki-ingest`
- **代码目录未配置或不存在**：仍可输出文档侧 scope，但必须说明未做代码核对
- **`zatools qmd ...` 不可用**：回退到本地搜索，并说明召回质量可能下降
- **命中过多且分散**：不要继续发散，直接向用户提问
- **找不到任何有效证据**：报告当前为空结果，并请求至少一个入口锚点
