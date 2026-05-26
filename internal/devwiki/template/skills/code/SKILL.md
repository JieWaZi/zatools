---
name: "devwiki-code"
description: "当用户要求修改代码、开发功能、修复 bug、调整接口/配置/业务逻辑、重构实现、补测试或提交代码时使用。它以 DevWiki Workflow 作为代码入口和规则证据，但目标是改当前代码仓，不是回答知识问题，也不是生成 DevWiki 文档。"
argument-hint: "<要开发、修改、修复或提交的功能/问题>"
---

# /devwiki-code

## 核心定位

`devwiki-code` 负责把 DevWiki 中的 Workflow 锚点转化为当前代码仓的真实代码改动，并用最小搜索、最小改动、最小验证完成开发任务。

- `devwiki-query`：只读回答知识、规则、现有实现和排障线索，不改代码。
- `devwiki-code-to-doc`：从真实代码反向生成或更新 DevWiki 文档，不改生产代码。

Code 不是 Query，也不是 Code-to-Doc：本 Skill 的目标是修改当前代码仓，不能只停在知识回答，也不能把代码修改任务改写成文档生成任务。

如果用户目标是修改代码，即使问题里包含“在哪里、怎么改、影响面”，也优先使用本 Skill，不转回 `devwiki-query`。

---

## 模式判断

开始前先判断当前属于哪种模式。不要所有场景都重复走完整 DevWiki 定位流程。

| 场景 | 模式 | 行为 |
|---|---|---|
| 用户首次提出开发 / 修复 / 修改需求 | 首次开发模式 | 走 DevWiki 定位流程 |
| 用户说“刚才不对”“按我的想法改”“不要这样写”“换一种实现” | 续改模式 | 不重新读 Wiki，优先看当前 diff 并定向 patch |
| 用户要求“重新分析”“换个方案”“这个方向废掉” | 重做模式 | 先看当前 diff，再决定是否重新读取 Workflow |
| 用户新增另一个功能点或模块 | 新任务模式 | 可以重新走 DevWiki 定位流程 |
| 用户指出具体文件 / 函数 / 代码块 | 定点修改模式 | 直接查看对应文件和 diff，不全局搜索 |

---

## 首次开发模式

首次开发模式用于用户第一次提出本轮开发、修复或修改需求。

默认路径：

压缩流程：理解用户问题 → 结构化搜索 → card 验证 → 读取正文 → 代码核对 → 测试 → 实现 → 验证。

读取原则：

- Workflow core 是代码修改的主入口。
- 不能因为要改代码就跳过 DevWiki。
- 已有 Workflow core 锚点时，不先全仓搜索。
- `zatools devwiki search` 返回的是候选排序，不是真相源；必须通过 `read ... --view card` 验证后再继续。
- `zatools devwiki read workflow` 的页面参数必须使用 search 结果里的 `slug` 字段值，不要使用 `file` 文件名。

---

## 续改模式

如果用户是在当前任务后续轮次中，对刚才的代码改动表示不满意，或要求“按我的想法改 / 换一种实现 / 不要这样写 / 只改这里”，必须进入续改模式。

续改模式下：

1. 不重新执行完整 DevWiki 读取流程。
2. 不重新搜索 `wiki/index.md`。
3. 不重复读取 workflow card/core，除非用户的新要求改变了功能范围、模块边界或原 Workflow 锚点明显不适用。
4. 优先查看当前工作区改动：

   ```bash
   git status --short
   git diff
   ```

5. 基于用户反馈定位需要调整的已改文件和已改代码块。
6. 只做针对性 patch，不做无关重构。
7. 如果用户的新要求与原业务规则冲突，`workflow explain` 核对。
8. 如果两轮 patch 后仍无法对齐用户意图，停止继续猜测，向用户提出具体问题。

### 不重复流程规则

如果当前对话中已经完成过一次 DevWiki 定位，并且后续用户反馈只是调整同一任务的实现方式，不得重复执行完整读取流程。应复用已有上下文、当前 diff 和用户最新指令，进行最小范围修改。

---

## 读取规则

开始前必须读取：

- 关联代码仓 `AGENTS.md` / `CLAUDE.md` 中的 DevWiki link block；
- `config/project.yaml`。

首次开发模式下，按以下顺序定位知识。

### Phase 1：理解用户问题

先判断问题类型：

| 问题类型 | 优先搜索类型 | 说明 |
|---|---|---|
| topic | topic | 功能规则、业务语义、配置含义、功能边界 |
| workflow | workflow | 代码定位、接口链路、实现入口、改动影响 |
| troubleshooting | workflow | 错误日志、故障现象、运行时异常；确认要改代码后落到 workflow |

再提取关键词和同义词：
- 功能名、模块名、页面名、接口路径、配置项、错误码、日志关键字；
- 中文业务名和英文代码名；
- 用户原词和常见同义词；
- 如果关键词过短，例如 `vip`、`auth`、`sync`，必须组合模块名、错误码、接口路径或日志片段一起搜索。

### Phase 2：结构化搜索

使用 `zatools devwiki search <kind> <query...>` 召回候选。`<kind>` 只使用 `topic` 或 `workflow`。多个关键词应作为多个参数传入，不要合并成一个带空格的字符串。

```bash
zatools devwiki search workflow "防脑裂" "网关" "ha-group" "gateway"
```

`search` 返回 JSON 数组，每个元素包含 `file`、`slug`、`title`、`score`：

```json
[{"file":"workflow-ha-brain-split-protection.md","slug":"workflow-ha-brain-split-protection","title":"HA 脑裂监控与防护实现定位","score":"83%"}]
```

字段含义：

- `file`：页面文件名，只用于人类核对来源，不作为 `read` 参数；
- `slug`：页面唯一标识，必须作为后续 `read workflow/topic` 的参数；
- `title`：候选标题，用于判断语义是否匹配；
- `score`：召回排序分数，只代表候选优先级，不代表结论正确。

根据 `score`、`title`、`slug` 判断 top candidates。优先处理 top 5，不要一次展开所有候选。

### Phase 3：card 验证

对高分候选逐个读取 card，判断是否真的匹配用户问题。

Workflow 候选：

```bash
zatools devwiki read workflow <slug> --view card
```

注意：`read workflow/topic` 后面的参数必须使用 `slug` 字段值。例如 search 返回：

```json
[{"file":"workflow-ha-brain-split-protection.md","slug":"workflow-ha-brain-split-protection","title":"HA 脑裂监控与防护实现定位","score":"83%"}]
```

正确读取方式是：

```bash
zatools devwiki read workflow workflow-ha-brain-split-protection --view card
```

不要写成：

```bash
zatools devwiki read workflow workflow-ha-brain-split-protection.md --view card
```

card 匹配后再进入正文读取；card 不匹配就换下一个候选，不读取 core/explain。

### Phase 4：读取正文

按问题类型选择正文层级：

| 问题类型 | 读取方式 | 说明 |
|---|---|---|
| 代码定位问题，实现机制问题 | `zatools devwiki read workflow <slug> --view core` | 获取文件、函数、接口、配置、日志锚点 |
| 深入背景问题 | `zatools devwiki read workflow <slug> --view explain` | 只有用户要求背景、设计意图或 core 不足时读取 |
| 排障后修复 | 先读最相关 workflow card/core，再结合错误日志和真实代码 | troubleshooting 作为问题类型，不作为 `devwiki search` 的 kind |

读取正文时继续使用 card 验证通过的 `slug`。

注意： 第一轮只读取workflow core，后续如果workflow内容无法支持代码定位，在第二轮会使用 workflow explan。

### Phase 5：兜底

top 5 不满足，再 grep `wiki/index.md`：

```bash
grep -iE '<关键词>' wiki/index.md
```

如果后续存在 `zatools devwiki index`，可用它替代直接 grep `wiki/index.md`。

两轮搜索没有新增有效候选，停止盲搜并向用户提问。不要不断换关键词、不断全仓搜索。

查到结果后也是走card 验证， 由于目前index.md没有区分workflow和topic，存在多个结果时结合内容判断 top candidates。优先处理 top 5，不要一次展开所有候选。然后继续走Phase 3，没获取到内容就继续，不要因为zatools报错就停止。除非所有的结果都没获取到，就得提示用户未从文档中获取到有效内容。

### 补充
不到万不得已静止使用topic查询。
---

## 任务分型

开始后先判断任务类型，并按类型选择读取内容。

| 任务类型 | 优先读取 | 说明 |
|---|---|---|
| 修 bug | workflow core + 错误日志 + 真实代码 | 先定位问题路径，不急着读 explain |
| 新增功能 | topic core + workflow core | 先确认规则，再找实现入口 |
| 修改业务规则 | topic core + workflow core | Topic 确认规则，Workflow 定位代码 |
| 修改实现机制 | workflow core + workflow explain | 需要理解模块协作和副作用 |
| 重构 | workflow core + workflow explain | 先确认边界和影响面 |
| 调整接口 / 配置 | workflow core + topic core | 同时核对入口、存储/下发、消费 |
| 补测试 | 真实代码 + 现有测试目录 | Workflow 不提供测试入口时，不强行依赖 Workflow |
| 排障后修复 | troubleshooting + workflow core | 确认要改代码后进入实现 |
| 续改 / 用户不满意 | git diff + 用户最新指令 | 不重新走完整 DevWiki 流程 |

---

## Bash / 搜索范围纪律

目标：从 Workflow core 锚点逐级扩大范围，减少 token、噪声和误判。

搜索范围优先级：

1. 明确文件 + 精确 symbol / 字段 / 常量；
2. 明确文件 + 相关关键词；
3. Workflow core 给出的同目录；
4. Workflow core 给出的同模块；
5. 代码仓根目录。

已有明确文件时：

```bash
rg -n '<关键字段|函数|常量>' <明确文件>
sed -n '<start>,<end>p' <明确文件>
```

自动执行 `rg -n '<关键字段|函数|常量>' <明确文件>` 时，只读取必要上下文，不整文件展开。

已有明确目录时：

```bash
rg -n '<关键字段|函数|常量>' <明确目录>
rg --files <明确目录> | rg '<文件名关键词>'
```

只有没有文件锚点时，才允许从模块目录扩大：

```bash
rg -n '<接口路径|配置项|日志关键字>' <明确模块目录>
```

最后才允许代码仓根目录：

```bash
rg -n '<高置信唯一关键词>' .
```

- 已经明确要读取的多个文件、多个短片段，可以在一次 shell 调用中批量读取，减少工具往返。
- 优先使用 `rg` / `grep -R` 精准定位，再用 `sed -n` / `awk` 读取必要范围。

限制：

- 禁止在已有明确文件锚点时直接 `rg -n '<关键词>' .`。
- 禁止同时对 `wiki/`、`raw/`、代码仓根目录并发大范围搜索。
- 禁止用宽泛词全仓搜，例如 `config`、`status`、`gateway`、`update`，除非加上明确文件或目录。
- 禁止用 `find .` 替代 `rg --files`。
- 同一范围内同义关键词搜索 2 次仍无结果，记录“该范围未命中”，再扩大一层。
- 读取文件优先 `sed -n` 小范围查看，不要整文件 `cat` 大文件。
- 必须先在 workflow core 给出的文件内搜索：
     ```bash
     rg -n '<函数|字段|配置项|接口路径>' <workflow-core-file>
   ```
- 如果一个文件内命中，再只读命中点附近 40-120 行：

     sed -n '<start>,<end>p' <workflow-core-file>

- 如果需要追调用链，只能扩大到：
      - 同文件；
      - workflow core 给出的其他文件；
      - 当前文件同目录。
- 只有满足以下全部条件，才允许全仓搜索：
      - workflow core 没有给出任何可用文件锚点；或所有锚点文件都不存在；
      - 已在 workflow core 文件、同目录、同模块各搜索至少 1 轮仍无有效命中；
      - 搜索词是高置信唯一词，例如完整接口路径、完整配置字段、完整错误码、完整类名；
      - 执行前在回复中说明为什么必须扩大到仓库根目录。
---

## 防 Bash 死循环规则

不要陷入“不断换关键词、不断找目录、不断全仓搜索”的循环。

必须停止盲搜并向用户提问的情况：

1. 已按搜索范围扩大到同模块或全仓，但连续 2 轮没有新增有效锚点；
2. Workflow core 提供的文件路径与当前仓库不一致，且无法确认迁移位置；
3. 用户需求依赖业务判断，但 Topic core 没有明确规则；
4. 同一个概念出现多个相似实现，无法判断应该改哪一个；
5. 当前代码存在多套版本、插件、适配层或产品线分支，无法确定目标环境；
6. 修改可能影响数据结构、配置格式或兼容行为，但缺少版本/范围约束。

提问要求：

- 问具体问题，不问泛泛的“你想怎么改”。
- 带上已经确认的信息和卡住点。
- 一次最多问 2 个关键问题。

示例：

```text
我已经确认当前仓库里有两套 forward-zone 实现：A 用于旧 DNS 服务，B 用于 v3 本地服务。你这次要改的是哪一套？如果不确定，我可以先按当前接口路径继续定位。
```

---

## Workflow 锚点展开

读完 workflow core 后，先提取：

- 文件路径：第一搜索范围；
- 函数 / 类 / 常量：第一关键词；
- 接口 URL / 配置项 / 日志关键字：调用链定位关键词；
- 修改影响：风险判断入口。

执行顺序：

1. 对 workflow core 中列出的关键文件，用 `rg -n '<函数|字段|配置项>' <file>` 定位行号。
2. 只读取命中行上下文，例如 `sed -n '120,180p' <file>`。
3. 如果函数内部调用其他本地函数，再用 `rg -n 'def <callee>|<callee>\(' <同文件或同目录>` 追一层。
4. 只有调用链断裂时，才扩大到同模块目录。
5. 如果两轮展开仍无新增有效锚点，停止盲搜并向用户提问。

---

## 实现纪律

- 修改前先看 `git status --short`，不要覆盖用户已有改动。
- 先理解现有实现边界，再做最小必要改动。
- 不做无关重构、格式化或大范围风格调整。
- 涉及接口、配置结构或持久化格式时，至少核对：
  - 入口接收；
  - 存储 / 下发；
  - 运行时消费。
- 涉及状态、副作用或跨模块联动时，必须读取 Workflow explain 或真实代码确认机制。
- 不要默认写入 DevWiki 文档；只在结果中提示可能过期的页面。

---

## 续改实现纪律

当用户对已完成改动提出调整时：

1. 先看 `git diff`，理解当前已改内容。
2. 明确哪些改动保留、哪些撤销、哪些替换。
3. 优先修改已变更文件，不扩大到新模块。
4. 如果用户指定方案，以用户方案为准；除非明显会破坏功能或与已知规则冲突。
5. 如果需要撤销局部改动，不要直接重置整个工作区。
6. 修改后说明“按你的反馈调整了哪里”，不要重新解释完整需求。

---

## 最终输出

完成后回答：

- 改了哪些文件和关键行为；
- 跑了哪些测试或验证命令，结果是什么；
- 哪些 Wiki 页面可能过期，需要后续用 `devwiki-maintain` 更新；
- 修改后导致 Wiki 过期时，只提示后续使用 `devwiki-code-to-doc`，不要在本轮默认写 DevWiki；
- 如果无法验证，说明缺失的环境、命令或依赖；
- 如果是续改模式，说明按用户反馈调整了哪些点。

不要输出大段代码全文。必要时只引用关键片段或说明路径。
