---
name: "devwiki-feature-doc"
description: "当需要围绕某个研发特性从现有 wiki 与代码反向梳理原始特性文档时使用，尤其适用于缺少文档、文档过期，或需要从接口 URL、关键文件、关键函数、页面路径向下追完整调用链并落到 raw/features 的场景。"
argument-hint: "<特性名称>"
---

# /devwiki-feature-doc

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


为指定特性生成或更新 `raw/features/<特性名称>特性文档.md`。

## 输入要求

- 必须先拿到明确的特性名称。
- 优先要求用户提供至少一个入口锚点：`接口 URL`、`关键文件`、`关键函数`、`页面路径/路由`。
- 如果只有特性名称，先自主检索 `wiki/`、`raw/` 与本地代码；检索几轮仍无法稳定定位时，立即向用户提问。


## 执行规则

1. 先阅读 `references/source-priority.md`，明确资料优先级。
2. 如果同名文件已存在，先询问用户是更新原文档，还是改名后新建；未经确认不要覆盖。
3. 开始正式排查前，阅读 `references/trace-playbook.md` 与 `references/question-rules.md`。
4. 先执行 `zatools qmd status`。若 `zatools qmd status` 正常，优先用 `zatools qmd query` 在 `wiki / raw / code` 三层召回，再用 `zatools qmd get` / `zatools qmd multi-get` 读取 top-K 命中。
5. 梳理时必须覆盖完整业务闭环，不允许只看入口主函数就停止。
6. 遇到查不到、调用链断裂、动态分发无法确认、外部依赖缺少上下文时，先自主尝试几轮，再按提问规则向用户追问。
7. 起草前先阅读 `references/doc-template.md` 与 `references/section-examples.md`，按模板产出，不要自由发挥章节结构。
8. 写入 `raw/features/` 前，先给出证据摘要、拟写路径与待确认事项，得到确认后再落盘。

## 禁止事项

- 禁止只根据函数名或目录名臆测功能。
- 禁止跳过关键分支、关键调用、关键数据结构。
- 禁止把未确认的推断写成既定事实。
