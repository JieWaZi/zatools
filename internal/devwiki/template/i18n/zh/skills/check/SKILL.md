---
name: "devwiki-check"
description: "当需要对 DevWiki 的 capabilities、features、链接、source hash、code refs、symbol 与索引状态做确定性健康检查时使用。"
argument-hint: "[check-scope]"
---

# /devwiki-check

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 对 DevWiki 做确定性健康检查并输出分层报告。默认只读。

## Inputs

- `scope`（可选）：检查范围，例如“整个 wiki”或“用户权限相关页面”
- `wiki/` 目录
- 可选 `--fix` — 自动修复确定性的低风险问题
- 可选 `--fix --dry-run` — 只预览修复，不落盘
- 可选 `--json` — 输出 JSON 结果

## Outputs

- health report
- 可选 repair preview 或 repair result
- 可选写入 `wiki/outputs/check-report-<date>.md`
- 追加到 `wiki/log.md` 的 check 摘要

## DevWiki Interaction

### Reads

- `wiki/capabilities/*.md` — 检查 feature 关联与必填字段
- `wiki/features/*.md` — 检查 `sources`、`code_refs`、`api_entries`、`test_refs`
- `wiki/index.md` — 检查目录完整性
- `raw/*/*.md` — 检查来源是否仍存在
- 本地代码目录 — 检查 `code_refs.path` 和 `symbol`

### Writes

- 默认不写页面
- 只有在显式传入 `--fix` 且问题确定、低风险时，才允许自动修复
- APPEND `wiki/log.md`


## Workflow

### Step 1: 做基线检查

至少要检查：

1. 必填字段缺失
2. `sources.path` 缺失
3. `sources.hash` 失配
4. `code_refs.path` 缺失
5. 被引用文件里 `symbol` 缺失
6. `index.md` 过期或缺项
7. capability-feature 反向链接缺失
8. 孤儿页
9. `qmd` 索引状态过期

### Step 2: 给发现分层

将问题分成：

- 🔴 立即处理：确定性坏链、坏路径、坏 hash、坏 symbol
- 🟡 建议处理：capability-feature 关系异常、索引陈旧、反向链接缺失
- 🔵 可选改进：孤儿页、低价值冗余线索

### Step 3: 处理 fix 模式

1. 默认模式是只读报告
2. 如果用户传了 `--fix`：
   - 只修确定性的低风险问题
   - 例如：刷新 `sources.hash`、移除失效的辅助 code clue、修简单索引项
3. 如果用户传了 `--fix --dry-run`：
   - 只预览修复候选，不写入
4. 凡是需要重新判断 capability 边界或 feature 归属的，都不能自动修；应分流到 `/devwiki-refresh`

### Step 4: 输出结果

报告必须包含：

- 检查范围
- 问题数量
- 按严重级别分组的详情
- 若执行了 fix：修了什么、预览了什么、哪些仍需人工处理
- 下一步建议：
  - 对确定性问题可继续用 `--fix`
  - 对结构漂移使用 `/devwiki-refresh`
  - feature 文档缺失时使用 `/devwiki-feature-doc`

### Step 5: 记录执行

在 `wiki/log.md` 追加：

- `check | report-only | <summary>`
- 或 `check | fix-applied | <summary>`

## Constraints

- **默认只读**：没传 `--fix` 时，不得修改 wiki 页面
- **`--fix` 只修确定性低风险问题**：不得自动修 capability 边界或 feature 归属
- **raw/ 只读**：不得修改来源层
- **不得虚构 symbol**：找不到就是找不到，必须如实报告
- **结果应稳定**：同一状态下重复执行，结果应基本一致

## Error Handling

- **缺少 wiki/**：提示用户先执行 `/devwiki-init`
- **代码目录缺失**：跳过代码检查，并说明覆盖范围下降
- **`zatools qmd ...` 不可用**：报告索引校验覆盖下降，但继续做其他检查
- **fix 模式遇到中高风险项**：停止自动修复该项，并建议转到 `/devwiki-refresh`
