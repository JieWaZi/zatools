---
name: "devwiki-code-to-doc"
description: "当需要从真实代码、接口 URL、关键文件、关键函数、页面路径、路由、配置项或日志反向整理 DevWiki 文档时使用。该 Skill 专注于指导 Agent 如何从代码入口逐级追踪、理解调用链、识别关键逻辑、状态读写、配置处理、异常路径和修改影响；最终页面结构复用 ingest 中定义的 Capability / Feature / Workflow / Troubleshooting 模板"
argument-hint: "<功能名称、接口 URL、关键文件、关键函数、路由、配置项或日志关键字>"
---

# /devwiki-code-to-doc

## 一、使用前置

开始前先读取通用约束：

- `references/evidence-grounding.md`
- `references/zatools-qmd.md`
- 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
- 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`

生成或更新页面时，不在本 Skill 内重新定义模板，直接复用 ingest 体系中的模板：

- Capability：复用 `references/capability_template.md`
- Feature：复用 `references/feature_template.md`
- Workflow：复用 `references/workflow_template.md`
---

## 二、核心定位

Code-to-doc 不是模板写作 Skill，而是代码理解与证据追踪 Skill。

它负责回答：

- 从哪个代码入口开始；
- 如何一级一级向下追踪；
- 追到什么深度可以停止；
- 哪些代码证据可以支撑 Wiki；
- 哪些结论应该进入 Workflow；
- 哪些功能语义需要同步到 Feature；
- 哪些能力边界变化需要提交 Capability proposal；
- 哪些异常、日志、恢复路径需要同步到 Troubleshooting；
- 当代码、raw、wiki 冲突时如何处理；
- 什么时候必须向用户提问。

最终落页结构由 ingest 的模板负责，本 Skill 只负责把代码理解结果整理成可写入的证据和 proposal。

---

## 三、默认产出

默认产出优先级：

1. 默认写入 `wiki/workflows/<slug>.md`
    - 代码入口；
    - 调用链；
    - 关键逻辑；
    - 状态读写；
    - 配置处理；
    - 异常与恢复实现；
    - 代码引用；
    - 测试引用；
    - 修改影响。

2. `wiki/features/<slug>.md`
    - 仅当代码追踪能够确认功能行为、参数语义、联动、边界或验收关注点时同步更新。
    - Feature 不写代码引用，只链接 Workflow。
    - Feature 的 sources 不写代码文件路径或 `kind: code`；代码证据只写入对应 Workflow 的 `code_refs`。

3. `wiki/troubleshooting/<slug>.md`
    - 仅当输入锚点是日志、错误码、异常现象，或代码追踪确认了诊断/恢复路径时更新。

4. `wiki/capabilities/<slug>.md`
    - 仅当代码追踪发现能力边界、覆盖功能或能力关系需要调整时，输出 proposal。
    - 不默认直接扩写 Capability。

---

## 四、输入锚点

用户至少应提供一个锚点：

- 功能名称；
- API URL；
- 路由；
- 页面路径；
- 关键文件；
- 关键函数；
- 配置项；
- 日志关键字；
- 错误码；
- 已知 Feature / Workflow slug。

如果只有功能名称，不要马上提问。先自主搜索 wiki、raw、代码。  
只有多轮搜索后仍无法稳定定位，才向用户提问。

---

## 五、来源优先级

默认按以下顺序调查：

1. `wiki/`
    - 先看已有 capability、feature、workflow、troubleshooting，避免重复整理。
2. `raw/`
    - 再看需求、设计、功能说明、接口说明、测试资料，理解历史意图。
3. 本地代码
    - 用当前代码确认真实实现，并纠正文档漂移。
4. 用户澄清
    - 只有无法继续确认或必须用户拍板时才提问。

关键原则：

- `wiki/` 和 `raw/` 是线索与历史，不是最终实现真相。
- 当前实现结论必须由当前代码支撑。
- 如果代码与 wiki/raw 冲突，必须在 proposal 或 Workflow 的“代码核对结论 / 来源说明”中明确写出。
- 不得把历史设计默认当成当前实现。
- 不得把代码现状写成产品设计，除非明确标注为“实现现状”。

---

## 六、代码追踪方法

追踪深度和停止条件遵守 `references/code-tracing.md`。本 Skill 只补充 code-to-doc 的入口优先级：

1. API URL / route；
2. controller / handler；
3. service 主方法；
4. 关键文件；
5. 配置项；
6. 日志关键字；
7. 调用方反查；
8. 全局搜索。

不要找到第一个同名函数就停止。  
要确认它是否真的是当前功能链路的一部分。

---

## 七、证据记录

代码证据必须明确记录。

建议在 proposal 中输出：

```markdown
## 代码证据摘要

| 证据 | 类型 | 说明 | 置信度 |
|---|---|---|---|
| `<path>` / `<symbol>` | file/function/config/test |  | high/medium/low |
```

需要标记：

- 已确认入口；
- 已确认调用链；
- 已确认关键逻辑；
- 已确认状态读写；
- 已确认配置处理；
- 已确认异常路径；
- 已确认测试入口；
- 未确认动态分支；
- 与 wiki/raw 的冲突；
- 需要用户确认的锚点。

证据要求：

- 不得编造代码路径、函数名、模块名、接口名。
- 不确定的代码线索不得标 high confidence。
- 只把已核对代码放进 `code_refs`。
- `code_refs` 只能进入 Workflow，不能进入 Feature 或 Capability。
- Feature 的 `sources` 只能记录 raw、已有 Wiki 或用户提供的非代码资料；即使 Feature 结论来自代码核对，也不要把代码文件路径、函数名或 `kind: code` 写进 Feature。
- 代码证据结构遵守 `references/evidence-grounding.md` 中的 `code_refs` 文件级规则。

---

## 八、归属判断

| 代码追踪发现 | 写入位置 |
|---|---|
| 入口、调用链、类、模块、函数、handler | Workflow |
| 状态读写、配置读取、校验、同步、下发 | Workflow |
| 异常、失败、重试、回滚、恢复实现 | Workflow |
| 代码与 raw/wiki 的实现差异 | Workflow |
| 修改影响、测试引用、验证建议 | Workflow |
| 能从代码确认的功能行为、参数语义、边界规则 | Feature |
| 能从代码确认的功能联动、异常边界、验收关注点 | Feature |
| 日志、错误码、诊断路径、修复/恢复路径 | Troubleshooting |
| 能力边界、覆盖功能、能力关系变化 | Capability proposal |

不要在 code-to-doc 中重新定义这些页面的小节模板。落页时读取并遵循 ingest 对应模板。

---

## 九、冲突处理

如果代码与 wiki/raw 冲突：

1. 不要静默选择一方；
2. 不要直接把 raw 改掉；
3. 不要把历史设计写成当前实现；
4. 不要把代码现状写成产品设计，除非明确标注；
5. 在 proposal 中输出差异；
6. 确认后：
    - 功能语义差异写 Feature；
    - 实现差异写 Workflow；
    - 排障差异写 Troubleshooting；
    - 能力边界差异写 Capability proposal。

如果无法判断谁更新，必须提问确认。

---

## 十、提问规则

允许先自主尝试几轮；仍无法确认时，必须停下来向用户提问。

必须提问的场景：

- 只有特性名称，没有稳定入口，且多轮检索后仍出现多个候选实现；
- API URL、页面路径、关键函数在代码库里完全找不到；
- 代码通过动态注册、反射、配置下发、模板生成等方式间接生效，无法静态确认；
- 关键外部系统、网关、三方接口没有本地实现或缺少文档；
- 同名文件已存在，且需要决定“更新”还是“新建”；
- 代码与现有 wiki/raw 明显冲突，无法判断哪一份更新；
- 需要新增、拆分、合并或重命名页面。

提问方式：

- 问题要短；
- 只问缺失锚点；
- 不要一次问一串大问题；
- 不要泛泛地问“能多给点信息吗”。

示例：

```text
我在代码里找到两个候选入口：`a` 和 `b`。你希望我从哪条链路开始？
```

```text
我没在本地代码中找到接口 `<URL>`。这个接口是否在网关、子仓库或外部服务里？
```

---

## 十一、写入 Proposal

落盘前必须按 `references/mutation-safety.md` 输出 proposal。

```markdown
# Code-to-Doc Proposal

## 输入锚点

## 已检查资料

| 来源 | 结果 |
|---|---|

## 代码追踪摘要

| 层级 | 发现 | 证据 |
|---|---|---|
| 入口层 |  |  |
| 分发层 |  |  |
| 服务层 |  |  |
| 状态层 |  |  |
| 副作用层 |  |  |
| 异常层 |  |  |
| 测试层 |  |  |

## 建议写入

| 页面 | 类型 | 动作 | 原因 | 置信度 |
|---|---|---|---|---|

## 与 Wiki / Raw 的差异

## 需要用户确认的问题

## 暂不写入的内容
```

任何 Wiki 文件创建、修改或删除都必须等用户在 proposal 之后明确确认。风险分层和确认口径遵守 `references/mutation-safety.md`；以下 code-to-doc 场景必须在 proposal 中单独列出风险、候选方案和待确认问题：

- 新建页面；
- 拆分页面；
- 合并页面；
- 重命名页面；
- 改变 Feature / Workflow 主关系；
- 写入与 raw/wiki 冲突的当前实现结论；
- 删除或降级旧内容。

---

## 十二、确认后落盘

只有在用户明确确认 Code-to-Doc Proposal 后，才进入 `confirmed_write` 并执行：

1. 读取目标页面类型对应的 ingest 模板；
2. 创建或更新 `wiki/workflows/<slug>.md`；
3. 必要时更新 `wiki/features/<slug>.md`；
4. 必要时更新 `wiki/troubleshooting/<slug>.md`；
5. 必要时输出 Capability 调整 proposal；
6. 更新 `wiki/index.md`；
7. 更新 `wiki/glossary.md`；
8. 追加 `wiki/log.md`；
9. 执行或提示执行：

```bash
zatools qmd update
zatools qmd status
```

---

## 十三、禁止事项

### 13.1 调查禁止

- 不要只凭名称猜行为。
- 不要没查 wiki/raw 就直接读代码。
- 不要找到第一个同名函数就停止。
- 不要跳过关键分支、helper 调用或数据流转。
- 不要在还能继续自主排查时频繁打断用户。
- 不要泛泛地问“能多给点信息吗”。

### 13.2 写作禁止

- 不要在本 Skill 里重新定义 Capability / Feature / Workflow 模板。
- 不要把未核实推断写成事实。
- 不要编造代码路径、函数、接口、模块名。
- 不要把代码引用写进 Capability 或 Feature。
- 不要把代码文件路径、函数名或 `kind: code` 写进 Feature 的 `sources`。
- 不要为同一个代码文件生成多条 `code_refs`。
- 不要把一个文件中的每个方法都写成独立代码引用。
- 不要复制 Feature 的完整业务说明到 Workflow。
- 不要把 Workflow 写成逐行代码解释。
- 不要把历史 raw 设计默认写成当前实现。
- 不要静默处理代码与文档冲突。
- 不要在未确认情况下新建重复页面。
- 不要跳过 proposal 直接落盘。

### 13.3 粒度禁止

- 不要因为出现多个接口、多个 helper、多个分支就拆多个 workflow。
- 不要把一个完整功能链路拆成零散代码片段页。
- 不要把外部系统缺失实现强行写成本地实现。
- 不要在无法确认动态分发时写确定结论。

---
