# DevWiki 通用文件格式

本文件统一定义 `wiki/index.md`、`wiki/glossary.md`、`wiki/log.md` 的格式。任何 Skill 更新这三个文件前都先读取本文件，不要在 Skill 正文中重复定义这些格式规则。

## 共同规则
- 不使用 YAML frontmatter。
- 不使用 `devwiki:section`

## `wiki/index.md`

用途：DevWiki 的人工可读入口目录，帮助 Agent 快速找到权威页面。

固定标题：

```markdown
# Wiki Index
```

建议结构：

```markdown
# Wiki Index

## Topics

- [[<topic-slug>]]：<一句话说明>

## Workflows

- [[<workflow-slug>]]：<一句话说明>

## Troubleshooting

- [[<troubleshooting-slug>]]：<一句话说明>
```

写作规则：

- 只放当前推荐入口和高价值导航。
- 每条入口用 wiki link 加一句话说明。
- 不写完整页面摘要、实现细节、排障过程或维护报告。
- 不把 deprecated、report、outputs 页面作为主入口。

## `wiki/glossary.md`

用途：保存全局术语、稳定入口概念和常用别名，帮助 query 消歧和路由。

固定标题：

```markdown
# Glossary
```

建议结构：

```markdown
# 术语表

> 仅收录**跨页面、跨能力**的全局术语。配置项、状态文件、功能规则与局部概念见对应 Feature「关键概念」或 Workflow「数据与状态」；检索用词见各页 `search_terms`。

| 术语 | 说明 |
|------|------|
| 告警采集与外发 | 统一承接业务模块产生的告警消息，完成采集、去重、恢复、升级回滚保护、外发和节点告警接续 |
```

术语选择规则：
- 优先沉淀业务能力、系统能力、跨页面主题、稳定领域概念和常用别名。
- 不要把单个进程、文件名、字段名、配置键、CSV 列、动作码、函数名、日志关键字或检索词当作 glossary 术语。
- 一个术语说明要回答“这个能力/主题解决什么问题、覆盖哪些关键行为、和哪些场景相关”。

## `wiki/log.md`

用途：追加式记录 Wiki 维护动作，方便回溯写入历史。

固定标题：

```markdown
# Wiki Log

> Append-only chronological log.
```

追加格式：

```markdown
## YYYY-MM-DD

- <动作>：<路径或主题>；<一句话原因或影响>。
```

写作规则：

- 只追加，不重写历史记录。
- 每条记录写清动作、目标和原因。
- 不写长篇过程报告；维护过程报告写入 `wiki/outputs/`。
- 不把 log 当作功能事实来源。
