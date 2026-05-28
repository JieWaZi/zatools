---
name: "devwiki-project-router"
description: "当用户提出项目功能、设计文档、代码定位、代码修改、故障排查、Wiki 构建、Wiki 查询、Wiki 健康维护、代码反向成文或 qmd 检索层维护相关请求时使用。"
argument-hint: "<问题、任务或文档范围>"
---

# /devwiki-project-router

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`

> DevWiki 的总入口。它负责判断用户意图、证据来源、检索边界和下一步 Skill；不要把所有问题都直接回答掉。

## Inputs

- `request`：用户的自然语言问题、任务描述、文档范围、故障现象或待沉淀内容
- `context`（可选）：用户贴出的文档片段、对话记录、错误日志、代码锚点或已有 Wiki 路径
- `role`（可选）：用户明确说明的身份或面向对象，例如开发、产品、测试、运维、客户、官网、外部用户

## Outputs

- **始终先输出路由判断**：

```text
判断：这是 [意图类型]，命中 [目标 Skill]，需要/不需要 qmd，需要/不需要代码搜索。
```

- **然后执行目标 Skill 流程**，或在目标 Skill 尚未安装 / 尚未实现时，说明应该转入哪个目标 Skill 并给出下一步输入要求
- **必要时输出追问**：当意图、身份、证据范围或写入风险不清楚时，先问 1 到 3 个具体问题

## Routing Targets

| 意图类型 | 目标 Skill | 职责 |
|---|---|---|
| ingest | `devwiki-ingest` | 从原始资料建立或更新 DevWiki，生成 TopicTask / WorkflowTask / TroubleshootingTask、术语和入口导航 |
| topic_write | `devwiki-topic` | 创建或维护 `wiki/topics/`，只写主题边界、功能规则、关键状态和关联 Workflow |
| workflow_write | `devwiki-workflow` | 创建或维护 `wiki/workflows/`，只写工程入口、代码定位、调用链、修改影响和验证方式 |
| maintain | `devwiki-maintain` | 对已有 Wiki 做证据一致性、过期内容、引用缺失、入口错误和 query 污染维护 |
| code | `devwiki-code` | 基于 DevWiki workflow 定位并修改当前代码仓，开发功能、修 bug、重构、补测试或提交代码 |
| query | `devwiki-query` | 只读查询已有 Wiki、raw 和文档内代码线索，回答能力、功能、工程定位和排障问题；真实代码核查转 `devwiki-code` |
| code_to_doc | `devwiki-code-to-doc` | 从代码、接口、配置项、日志或路由反向生成或更新 DevWiki 页面 |
| qmd_maintenance | qmd maintenance commands | 直接执行 `zatools qmd sync/update/status`，不再路由到独立 qmd-sync skill |

## Intent Detection

### 1. 文档摄入 / Wiki 构建类

命中条件：

- 用户说「第一次构建」「初始化 Wiki」「从 20 个文档建立第一版框架」
- 用户给出一批原始文档，希望从零生成项目知识库
- 用户说「我有设计文档」「解析这个文档」「导入这批文档」「消化文档」「ingest」
- 用户提供新增或变化的需求、设计、测试、会议纪要等资料

路由到：

```text
devwiki-ingest
```

默认需要 qmd 或本地文档搜索来匹配已有页面；涉及写入时必须先形成提案，按风险确认后再落盘。
确认落盘后，`devwiki-ingest` 只负责任务编排；Topic 正文交给 `devwiki-topic`，Workflow 正文交给 `devwiki-workflow`。

### 2. Wiki 健康维护类

命中条件：

- 用户说「维护 Wiki」「检查 Wiki 健康」「maintain」「体检一下」
- 用户指出 query 回答用了旧规则、旧机制或过期页面
- 用户要求检查 raw/wiki/code 是否一致，或检查 topic 是否遗漏关键设计
- 用户要求修正冲突、过期、断链、孤立页、引用缺失、index/glossary 或页面入口错误
- 用户要求把旧内容标记为历史、降低旧页面检索命中，或避免 query 继续命中过期结论

路由到：

```text
devwiki-maintain
```

默认需要读取目标 Wiki、source、index/glossary 和页面入口链接；涉及实现偏差时需要代码搜索。中高风险修正必须先输出 Maintain Proposal，再按确认落盘。

### 3. 代码修改类

命中条件：

- 用户说「修改」「改成」「实现」「开发」「修复」「修 bug」「不生效」「重构」
- 用户要求调整接口、配置、业务逻辑、持久化、调用链、页面行为或运行时行为
- 用户要求补测试、跑测试、提交代码，或说「按这个方案做」「开始改」
- 用户描述目标状态，例如「从全局改为按 group 配置」「新增字段」「删除旧字段」

路由到：

```text
devwiki-code
```

默认需要 DevWiki workflow 定位和代码核对。路径优先为：

```text
index → workflow card/core → 必要 topic core → 代码核对 → 测试 → 实现 → 验证
```

不要转入 `devwiki-query` 做完整解释；Workflow core 是代码修改的主入口。
修改后如果发现 Wiki 过期，只提示后续使用 `devwiki-code-to-doc` 更新文档，不在 code 流程里默认写 Wiki。

### 4. 查询类

命中条件：

- 用户问「这个功能是什么」「某个逻辑在哪里」「代码入口在哪」
- 用户问「这个设计怎么理解」「这个功能和哪个文档有关」
- 用户要求「qmd 搜一下」「查一下 Wiki」「找相关设计 / 排障记录」
- 用户问当前实现、接口、函数、文件归属或运行时行为
- 用户只要求影响分析、代码定位或测试建议，但没有要求本轮修改代码

路由到：

```text
devwiki-query
```

默认需要项目知识检索。是否看代码取决于问题：

- 只问功能背景、设计意图、流程说明：先用 `wiki/` 和 `raw/`
- 问“怎么实现”、入口、模块职责、接口调用链但没有明确要求核对当前代码：仍走 `devwiki-query`，优先读 workflow card/core/explain，总结文档中的代码线索
- 明确要求查代码、核对当前实现、找文件函数/行号、确认真实调用链或运行态：改走 `devwiki-code`
- 文档证据足够时不要为了“更稳”默认展开代码
- 一旦用户转为要求修改、修复、开发或提交，改走 `devwiki-code`

### 5. 代码反向成文类

命中条件：

- 用户要求「从代码生成文档」「从接口反推功能说明」「代码梳理成 Wiki」
- 用户给出 API URL、关键文件、关键函数、路由、配置项或日志关键字，要求沉淀为文档
- 文档缺失或过期，需要以当前实现为准整理 workflow，或必要时补 topic / troubleshooting

路由到：

```text
devwiki-code-to-doc
```

默认需要代码搜索；默认写入 `wiki/workflows/`，写入 Wiki 前必须先给出证据摘要、拟写路径和待确认事项。
如果用户目标是修改生产代码，不走本 Skill，改走 `devwiki-code`。

### 6. qmd 检索层维护类

命中条件：

- 用户说「qmd 不可用」「qmd collection 没注册」「补接 qmd」「刷新索引」
- 用户要求检查 `config/search.yaml`、`zatools qmd sync/update/status/embed`
- 已有工作区需要补做或修复 qmd collection 注册、索引刷新与状态检查

处理方式：

```text
qmd maintenance commands
```

在本地 DevWiki 工作区直接执行 `zatools qmd sync/update/status` 等维护命令。只处理 qmd 检索层，不替代 ingest / query / code-to-doc；不再使用独立 `devwiki-qmd-sync` skill。

## Multi-Intent Priority

同一句请求可能命中多个意图时，按以下优先级收敛：
1. qmd maintenance commands
2. `devwiki-ingest`
3. `devwiki-maintain`
4. `devwiki-code`
5. `devwiki-code-to-doc`
6. `devwiki-query`

示例：
- 「qmd 搜不到这些 Wiki 页面」优先执行 `zatools qmd sync/update/status` 维护命令
- 「把这 20 个设计文档总结成 Wiki」优先是 `devwiki-ingest`
- 「query 总是答旧机制，帮我维护一下 Wiki」优先是 `devwiki-maintain`
- 「把防脑裂网关配置从全局改成按 HA group 配置」优先是 `devwiki-code`
- 「从这个接口反推出功能页」优先是 `devwiki-code-to-doc`
- 「这个功能是什么」优先是 `devwiki-query`

如果优先级仍无法判断，先追问用户目标产物：是回答问题、修改代码、写入 Wiki、维护已有 Wiki、从代码成文，还是修复 qmd 检索层。

## User Role

默认身份：

```text
developer
```

如果用户明确说「客户、外部用户、官网、对外页面」，视为：

```text
external_user
```

如果用户明确说「产品、测试、售前、支持、运维」，视为：

```text
internal_non_developer
```

访问范围：

| 身份 | 可访问范围 |
|---|---|
| developer | `wiki/` + `raw/` + skills + code |
| internal_non_developer | `wiki/` + 可公开或内部共享的 raw 摘要；默认不展开代码实现细节 |
| external_user | public only；不要暴露内部代码路径、未公开设计、排障细节或内部实现 |

如果用户身份与请求目标冲突，例如外部用户要求内部代码路径，先说明访问边界，再给可公开版本的回答或建议改走内部身份。

## Retrieval Rules

### 需要项目知识时

按 `references/zatools-qmd.md` 的“结构化入口优先，低置信升档”规则执行。Router 只判断是否需要项目知识、qmd 和代码搜索；具体分档、短词处理、fallback 和停止条件都由该 reference 统一维护。

### 必须查询项目知识的场景

- 查询功能说明、设计文档、工程定位、业务流程
- 查询排障经验、历史会议或已有 Wiki 页面
- 判断是否已有类似页面或类似 topic
- 准备把文档或代码证据写入 Wiki
- 判断请求应属于 ingest、maintain、query、code-to-doc 还是 qmd maintenance

### 不进入 DevWiki 检索的场景

- 纯通用 Linux 命令
- 纯语言语法或框架常识
- 用户只要求润色，且已经贴出完整上下文
- 用户明确要求不要查项目知识
- 请求与当前项目、Wiki、代码、文档沉淀无关

这类请求应简短说明「不进入 DevWiki 路由」，然后直接按普通任务处理。

## Code Search Rules

不要无边界全局搜索代码。只有以下情况才进入代码搜索；如果当前目标是 `devwiki-query`，遇到这些条件应转 `devwiki-code`，不要由 query 直接搜索代码：

- 用户明确问文件、函数、接口、调用链、实现现实、运行时行为
- `wiki/` / `raw/` 证据不足，且必须用代码才能回答
- 目标 Skill 要写入或修正 代码定位表
- 查询或讨论结果会影响开发实现，需要确认现有入口和边界

代码搜索顺序：

1. 先读取相关 workflow 页的 代码定位表
2. 如果已有明确代码锚点，直接定向读取或 `rg` 搜该锚点
3. 如果没有明确锚点，先按检索规则召回相关 topic / raw / code 候选
4. 只读取 top-K 候选文件中的关键入口、关键函数或关键边界
5. 代码与文档冲突时，分别标明「代码现状」和「文档描述」，不要混成一个结论

## Workflow

### Step 1: 判断是否属于 DevWiki

如果请求完全不涉及项目知识、DevWiki、代码定位、设计沉淀或知识维护，不进入本 Skill 的后续流程。

输出：

```text
判断：这是非 DevWiki 项目知识任务，不进入 DevWiki 路由。
```

然后按普通任务处理。

### Step 2: 判断用户身份与可见范围

1. 默认 `developer`
2. 识别是否为 `internal_non_developer` 或 `external_user`
3. 如果身份限制影响证据范围，先在路由判断后说明访问边界

### Step 3: 判断意图类型

按 Intent Detection 和 Multi-Intent Priority 分类。若命中不清，先问目标产物，不要硬路由。

### Step 4: 判断检索与代码需求

分别给出：

- 是否需要 qmd 或本地文档检索
- 是否需要代码搜索
- 是否涉及写入风险
- 是否需要用户确认

### Step 5: 输出路由判断

格式必须简短、稳定：

```text
判断：这是 [意图类型]，命中 [目标 Skill]，需要/不需要 qmd，需要/不需要代码搜索。
```

如果有身份边界或写入风险，在下一句补充。

### Step 6: 执行目标 Skill 流程

如果目标 Skill 已安装且可用，进入目标 Skill 的流程。

如果目标 Skill 尚未安装或尚未实现：

1. 明确说明应进入的目标 Skill
2. 按该目标的职责给出最小下一步，例如需要用户提供文档路径、问题锚点、对话记录或维护范围

## Constraints

- **先路由再执行**：不要跳过判断直接回答项目知识问题
- **qmd 是召回，不是事实源**：关键结论必须落到 `wiki/`、`raw/` 或已核对代码
- **代码搜索必须有边界**：不要为了保险全仓铺开
- **写入必须控风险**：凡是会创建、重挂、删除或修正中高风险关系的动作，都必须先提案并等确认
- **身份边界要明确**：外部用户不能看到内部代码路径和未公开资料
- **低置信必须追问**：几轮检索仍无法判断时，问具体锚点或目标产物
- **不要虚构目标 Skill 能力**：如果目标 Skill 尚未实现，说明路由结果和下一步，不要假装已经执行完整流程

## Error Handling

- **Wiki 基本为空**：若用户目标是整体建库或导入资料，路由到 `devwiki-ingest`
- **Wiki 有旧结论污染 query**：若目标是修正已有知识健康，路由到 `devwiki-maintain`
- **raw 为空**：提示先准备原始资料；若用户问题要求从代码反推，先确认是否允许基于代码生成草稿
- **qmd 不可用**：降级为本地搜索，并在回答中说明「本轮 qmd 不可用，已降级」
- **代码目录未配置**：先提示可用 `zatools devwiki repo info <project>` 检查 `code_repos`；本轮只基于 `wiki/` / `raw/` 处理，并说明未核对代码
- **目标 Skill 未安装**：输出路由结果、缺失 Skill 名称和需要的下一步输入
- **身份不允许访问**：提供可见范围内的回答版本，或要求切换到内部开发身份
