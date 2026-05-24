---
name: "devwiki-topic"
description: "当已确认需要创建或维护 DevWiki Topic 页面，或需要整理主题边界、功能机制、关键规则、关键状态和关联 Workflow 时使用。"
argument-hint: "<TopicTask、topic slug、原始资料范围或已有 topic 路径>"
---

# /devwiki-topic

## 定位

`devwiki-topic` 只负责生成和维护 Topic。

Topic 是主题页，用来回答：

- 这个主题解决什么问题；
- 功能边界和核心机制是什么；
- 哪些规则、配置、状态会影响行为；
- 哪些边界、误区和联动需要注意；
- 需要看实现时应该跳转哪个 Workflow。

Topic 不负责代码实现，也不负责测试设计。代码路径、函数名、handler、调用链、状态读写实现、修改影响、测试位置和测试方法必须写入其他页面或由 Workflow/代码追踪处理。

一句话边界：

```text
Topic 解释功能机制和规则；Workflow 解释代码如何实现；Troubleshooting 解释故障如何处理。
```

## 输入

常见输入：

1. `devwiki-ingest` 输出的 TopicTask；
2. raw 资料、设计片段、产品帮助文案或用户补充说明；
3. 已有 `wiki/topics/<slug>.md`；
4. 关联 Workflow 的 card 摘要；
5. 需要同步的 glossary/index 入口。

如果没有 TopicTask，也可以从用户给出的主题、资料或已有页面开始，但写入前仍必须按 `references/mutation-safety.md` 输出 proposal 并等待确认。

## 使用前置

开始前按需读取：

- `references/topic_template.md`：创建或重写 Topic 时必须读取；
- `references/evidence-grounding.md`：判断事实来源、推断和证据放置时读取；
- `references/knowledge-placement.md`：判断内容应该进入 card、core、explain 还是保留 raw 时读取；
- `references/mutation-safety.md`：任何写入、重命名、拆分、合并或关系调整前读取；
- `references/zatools-qmd.md`：本地 Wiki 命中低置信或需要 qmd 检索时读取。

不要读取 `workflow_template.md`。如果任务需要写 Workflow 正文，转交或加载 `devwiki-workflow`。

关联 Workflow 只允许读取 `card` 摘要来确认链接关系；不要为了写 Topic 去读取 Workflow 的 `code` 或 `explain`，避免把实现细节带入 Topic。

## 输出

默认输出：

```text
wiki/topics/<slug>.md
```

必要时附带：

- `wiki/index.md` 入口更新建议；
- `wiki/glossary.md` 术语更新建议；
- 关联 Workflow 的链接更新建议；
- `wiki/log.md` 追加内容。

## 核心规则
1. 只写主题边界、功能机制、关键规则、关键配置和状态。
2. 不写代码路径、函数名、handler、调用链、状态读写实现和修改影响。
3. 不写测试位置、测试方法、测试用例、验收清单或回归建议。
4. 不写 UI 字段清单、接口字段清单、完整操作列表或产品帮助全文。
5. 实现相关内容只写关联 Workflow 链接和一句话入口说明，不写代码路径。
6. 低频、低价值、高体积内容不进入 Wiki，保留 raw。
7. 不复制 raw 原文，不为了填模板强行写空内容。
8. 关系字段使用字符串数组，不维护 relation/reason 对象。
9. 页面必须包含 `card`、`core`、`explain` 三个 `devwiki:section`。

## 知识放置规则

Topic 编写时先判断知识的查询频率、决策价值、稳定性和体积成本。

| 内容类型 | 放置位置 |
|---|---|
| 一句话主题定位、适合/不适合回答、核心规则摘要 | `card` |
| 主题摘要、功能边界、核心机制、关键规则、关键配置与状态 | `core` |
| 典型场景、边界误区、低频规则补充、来源与待确认 | `explain` |
| 接口细节、字段清单、完整操作步骤、测试设计、代码细节 | 不进入 Topic |
| 低频、低价值、高体积原文 | `raw` |

## 写入流程

1. 明确 TopicTask 或目标 topic slug。
2. 检查 `wiki/index.md`、`wiki/glossary.md`、`wiki/topics/`，避免重复建页。
3. 读取相关 raw 和已有 topic。
4. 如有关联 Workflow，只读取其 card 摘要，不复制实现说明。
5. 读取 `references/topic_template.md`。
6. 按 `references/knowledge-placement.md` 分配 card/core/explain/raw。
7. 输出 Topic Proposal，列出拟写路径、动作、来源、风险和待确认问题。
8. 用户明确确认后写入。
9. 执行：

```bash
zatools devwiki check document
zatools devwiki check graph
zatools qmd update
zatools qmd status
```

## Topic Proposal

```markdown
# Topic Proposal

## 目标 Topic

## 输入来源

## 已检查页面

## 拟写入文件

| 路径 | 动作 | 写入重点 | 不写入内容 | 风险 |
|---|---|---|---|---|

## 关联 Workflow

## 内容放置判断

| 内容 | 放置位置 | 原因 |
|---|---|---|

## 入口和术语更新建议

## 冲突与不确定内容

## 需要你确认的问题
```

## 质量检查

- 是否没有代码路径、函数名、handler、调用链。
- 是否没有测试设计、验收清单、UI 字段清单或接口字段清单。
- 是否能用 card 判断命中。
- core 是否覆盖主题边界、核心机制、关键规则、关键配置和状态。
- explain 是否只是 core 的低频补充，而不是详细说明或产品帮助文档，是否控制篇幅，没有为了完整而堆内容。
- workflows / related_topics 是否是字符串数组。
- 来源、冲突和不确定内容是否清楚标注。
- 是否已执行 `zatools devwiki check document` 校验 `devwiki:section` 分块。
- 是否已执行 `zatools devwiki check graph`。
