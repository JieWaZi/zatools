---
name: "devwiki-feature-doc"
description: "当需要围绕某个研发功能，从现有 wiki 与代码反向整理或刷新结构化 feature 页面时使用，尤其适用于缺少文档、文档过期，或需要从接口 URL、关键文件、关键函数、页面路径向下追入口与关键调用边界的场景。"
argument-hint: "<特性名称>"
---

# /devwiki-feature-doc

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


为指定功能生成或更新 `wiki/features/<feature-slug>.md`。

## Input requirements

- 必须提供一个明确的 feature 名称。
- 强烈建议再给至少一个入口锚点：`API URL`、`关键文件`、`关键函数`、`页面路径 / 路由`。
- 如果只有 feature 名称，先在 `wiki/`、`raw/` 与本地代码中自行搜索；若几轮后位置仍不稳定，再向用户提问。

## Execution rules

1. 调查前先读 `references/source-priority.md`。
2. 如果目标页面已存在，默认原地更新；除非用户明确要求另起一页。
3. 代码追踪前先读 `references/trace-playbook.md` 与 `references/question-rules.md`。
4. 按 `references/zatools-qmd.md` 的阶梯召回规则执行，默认 local-first：
   - 先用本地 `grep` / 文件搜索解析已知锚点
   - 本地命中不足时再升档 `zatools qmd search`
   - 只有概念级召回需求且环境支持加速时才考虑 `zatools qmd query`；不支持时按共享 fallback 处理
5. 至少追到足以解释业务流程、约束、入口和实现锚点为止；如果只看 controller / handler 还不够，就继续往下走。
6. 如果代码找不到、调用链断裂、或动态分发导致无法确认，先自行尝试几轮，再向用户提问。
7. 落笔前阅读 `references/doc-template.md` 与 `references/section-examples.md`，按其结构组织内容。
8. 写入 `wiki/features/` 前，先给出证据摘要、拟写路径与待确认事项，得到确认后再落盘。
9. 页面应聚焦：支撑哪些 capability、业务流程是什么、约束是什么、入口在哪里、代码线索和测试入口是什么；不要写成完整实现大论文。

## Prohibited shortcuts

- 不要只凭命名猜行为。
- 不要跳过关键分支、helper 调用或数据流转。
- 不要把未核实的推断写成事实。
