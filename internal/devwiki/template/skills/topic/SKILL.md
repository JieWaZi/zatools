---
name: "devwiki-topic"
description: "当已确认需要创建或维护 DevWiki Topic 页面，或需要整理主题边界、功能规则、关键状态和关联 Workflow 时使用。"
argument-hint: "<TopicTask、topic slug、原始资料范围或已有 topic 路径>"
---

# /devwiki-topic

## 定位

`devwiki-topic` 只负责生成和维护 Topic。

Topic 是主题页，用来回答：

- 这个主题解决什么问题；
- 功能边界和核心行为是什么；
- 哪些规则、配置、状态会影响行为；
- 需要看实现时应该跳转哪个 Workflow。

Topic 不负责代码实现。代码路径、函数名、handler、调用链、状态读写实现和修改影响必须写入 Workflow 或 Troubleshooting。

## 输入

常见输入：

1. `devwiki-ingest` 输出的 TopicTask；
2. raw 资料、设计片段或用户补充说明；
3. 已有 `wiki/topics/<slug>.md`；
4. 关联 Workflow 的 card/core 摘要；
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

1. 只写主题边界、功能行为、关键规则、关键配置和状态。
2. 不写代码路径、函数名、handler、调用链。
3. 实现相关内容只写关联 Workflow 链接和一句话入口说明，不写代码路径。
4. 高频高价值内容进入 `card` / `core`。
5. 低频背景、验收、冲突、来源进入 `explain`。
6. 低频、低价值、高体积内容不进入 Wiki，保留 raw。
7. 不复制 raw 原文。
8. 不为了填模板强行写空内容。
9. 关系字段使用字符串数组，不维护 relation/reason 对象。
10. 页面必须包含 `card`、`core`、`explain` 三个 `devwiki:section`。

## 写入流程

1. 明确 TopicTask 或目标 topic slug。
2. 检查 `wiki/index.md`、`wiki/glossary.md`、`wiki/topics/`，避免重复建页。
3. 读取相关 raw 和已有 topic。
4. 如有关联 Workflow，只读取其 card/core 摘要，不复制完整实现说明。
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

## 入口和术语更新建议

## 冲突与不确定内容

## 需要你确认的问题
```

## 质量检查

- 是否没有代码路径、函数名、handler、调用链。
- 是否能用 card 判断命中。
- core 是否覆盖主题边界、核心行为、关键规则、关键配置和状态。
- explain 是否只放背景、验收、冲突、来源和低频细节。
- workflows / related_topics 是否是字符串数组。
- 来源、冲突和不确定内容是否清楚标注。
- 是否已执行 `zatools devwiki check document` 校验 `devwiki:section` 分块。
- 是否已执行 `zatools devwiki check graph`。
