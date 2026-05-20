# 证据落地约束

> 供 `devwiki-query`、`devwiki-ingest`、`devwiki-maintain`、`devwiki-code-to-doc`、`devwiki-project-router` 等技能共享使用。
> DevWiki 的输出必须落到真实来源、已核对代码证据，或被明确标注为推断。

## 核心规则

DevWiki 中每一个重要结论，至少要能回溯到以下三类之一：

1. `raw/` 原始资料
2. `wiki/capabilities/`、`wiki/features/`、`wiki/workflows/` 或 `wiki/troubleshooting/`
3. 配置代码目录中的已核对代码证据

`qmd` 只是召回加速器，不是真相源。

## 页面层级边界

DevWiki 只在对应页面维护对应事实，其他页面只做摘要和链接。

```text
Capability = 系统具备什么能力、能力边界是什么
Feature = 具体功能的行为和规则是什么
Workflow = 功能在代码中的实现路径怎么走
Troubleshooting = 故障如何识别、诊断和恢复
```

| 层级 | 权威内容 | 不应该写什么 |
|---|---|---|
| Capability | 能力定义、业务价值、能力范围、覆盖 Feature、能力间关系、能力级约束 | 具体功能规则、状态机、决策表、代码路径 |
| Feature | 功能目标、用户场景、触发条件、核心行为、关键规则、关键概念、重要配置、边界异常、验收关注点 | 代码入口、函数名、调用链、实现分支、完整排障步骤 |
| Workflow | 代码入口、调用链、类/模块/函数、状态读写、配置处理、异常实现、测试引用、修改影响 | 完整业务背景、完整 Feature 规则复述、能力价值说明 |
| Troubleshooting | 故障现象、日志、错误码、诊断路径、恢复步骤、相关 Feature / Workflow 链接 | 完整功能全貌、大段实现背景、未确认现场经验 |

归属原则：

- 能力定义、业务价值、能力边界写入 Capability。
- 功能目标、功能行为、功能规则、配置对行为的影响写入 Feature。
- 代码路径、函数名、调用链、配置读取/校验/下发代码、修改影响和测试文件写入 Workflow。
- 故障现象、日志、诊断和修复步骤写入 Troubleshooting。
- Feature 的 `sources` 不写代码文件路径、函数名、handler、调用链或 `kind: code`；代码证据统一写入 Workflow 或 Troubleshooting 的 `code_refs`。

## 来源优先级

### 原始资料层

`raw/` 是最强的来源层，适合承载原始需求意图、原始设计决策、原始功能说明、测试方案与测试记录。

页面应通过内联 `sources.path` 与 `sources.hash` 记录这些来源。

### 结构化 Wiki 层

`wiki/` 是维护后的知识层，不是事实起点。

适合承载：

- capability 的业务能力总结
- feature 的功能设计、参数、约束和联动
- workflow 的工程定位、调用链、代码引用和修改影响
- troubleshooting 的故障现象、诊断路径和修复建议

如果 `wiki/` 与 `raw/` 冲突，理论上按照wiki最新的为准，因为raw可能是过时文件，可在回答或 proposal 中显式列出冲突；

### 代码证据层

当问题涉及真实实现时，代码证据是必须的。

应使用：

- workflow 或 troubleshooting 中的 `code_refs`
- workflow 或 troubleshooting 中的 `api_entries`
- workflow 或 troubleshooting 中的 `test_refs`
- 直接文件 / symbol 核对

没有读过或核对过的文件、函数、路由、接口，不得硬说相关。

## 事实与推断

必须把事实和推断拆开表达。

事实包括：raw 路径存在、hash 匹配、文件存在、symbol 能找到、某个 wiki 页面明确写了某条关系。

推断包括：需求更像新增还是改造、某 capability 很可能覆盖该 feature、某代码路径大概率是主实现。推断可以写，但必须显式标注为推断，并给出可见证据。

## 检索顺序

建议顺序：

1. 先按问题意图查对应 wiki 层。
2. 再查 `raw/` 来源。
3. 先判断页面是否已经足够支撑答案。
4. 只有在页面证据不足，或问题明确要求实现现实时，再查 code。
5. 只有在代码证据真的必要时，才对 top-K 代码候选做本地核对。

如果文档已经足够支撑答案，默认不要为了“保险”再做一轮代码展开。

## 如何使用 `sources.hash`

`sources.hash` 不是装饰字段。它用于回答原始文件是否变化、页面是否过期、refresh 提案是否属于确定性修正。只要 raw 内容变了，就不能静默沿用旧 hash。

## 如何使用 `code_refs`

`code_refs` 应被当作结构化证据，而不是关键词残留。

`code_refs` 以代码文件 `path` 为唯一粒度。同一个 `path` 在同一页面中只能出现一条 `code_refs`。

每条 `code_refs` 至少回答：

- 哪个文件相关；
- 该文件在当前 workflow / troubleshooting 中承担什么文件级职责；
- 当前置信度是多少；
- 哪些关键入口 symbol 可作为后续追踪起点。

结构约束：

- 顶层 `kind` 固定使用 `file`，不要把方法、函数、类拆成多条 `code_refs`。
- 顶层 `note` 只写文件级职责，不写每个方法的说明。
- `symbols` 是关键入口索引，不是文件内方法清单。
- `symbols` 默认最多 4 个，只列主入口、关键状态读写、配置处理、外发、副作用、恢复或排障入口。
- 不得为了完整性列出文件内所有方法。
- `symbols` 使用 map：key 格式为 `<symbol>#<kind>`，value 是该关键入口的短说明。
- `<kind>` 可取 `method`、`class`、`function`、`constant`、`handler`、`config`、`task`。
- `symbols` 的说明只写入口职责或风险点，不写完整方法解释；不再维护独立的 symbol 说明 map。

推荐格式：

```yaml
code_refs:
  - path: "services/user/service.ts"
    kind: file
    note: "用户资料读写和状态同步的主服务实现。"
    confidence: high
    symbols:
      UserService#class: "用户资料服务主类。"
      UserService#updateProfile#method: "状态写入入口，修改时需要同步检查缓存刷新。"
```

`code_refs` 只属于 workflow 或 troubleshooting，不属于 capability 或 feature。

## 低置信处理协议

如果经过几轮有边界的检索后仍然证据不足：

- 停止继续扩散搜索
- 先总结已经找到的证据
- 再向用户提问 1 到 3 个关键问题

不要用空泛描述掩盖不确定性，应明确向用户提问。

## 不该做什么

- 不要把 `qmd` 命中当成事实
- 不要让过期的 wiki 页面压过已变化的 raw 来源
- 不要虚构 `code_refs`
- 不要把事实与推断写在同一句里混淆
- 不要在证据持续薄弱时无限搜索，应向用户提问
