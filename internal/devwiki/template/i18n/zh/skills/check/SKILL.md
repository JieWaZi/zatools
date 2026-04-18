---
name: "devwiki-check"
description: "当需要对 DevWiki 的 documents、capabilities、changes、链接、source_hash、code refs、symbol 与索引状态做确定性健康检查时使用，尤其适用于批量验收、refresh 前后核对和周期性巡检。"
argument-hint: "[检查范围]"
---

# /devwiki-check

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 对 DevWiki 执行确定性健康检查，并输出分级报告。默认只报告，不直接修。

## Inputs

- `scope`（可选）：检查范围，如“全部 wiki”或“用户权限相关页面”
- `wiki/` 目录
- 可选：`--fix` — 自动修复确定性且低风险的问题
- 可选：`--fix --dry-run` — 预览修复但不执行
- 可选：`--json` — 以 JSON 形式输出结果

## Outputs

- 检查报告
- 可选：修复预览或修复结果
- 可选：写入 `wiki/outputs/check-report-<date>.md`
- `wiki/log.md` 中的 check 摘要记录

## DevWiki Interaction

### Reads

- `wiki/documents/**/*.md` — 检查 `source_path`、`source_hash`、关联字段
- `wiki/capabilities/*.md` — 检查 `documents`、`changes`、`code refs`
- `wiki/changes/*.md` — 检查 change 关联与分类
- `wiki/index.md` — 检查页面目录完整性
- `raw/*/*.md` — 回核 source 是否仍存在
- 本地代码目录 — 检查 `code_refs.path` 和 `symbol`

### Writes

- 默认不写任何页面
- 仅在用户指定 `--fix` 且问题属于确定性低风险时，允许修复相关字段
- APPEND `wiki/log.md`


## Workflow

### Step 1: 运行基础检查

至少检查以下项目：

1. 必填字段缺失
2. `source_path` 不存在
3. `source_hash` 失配
4. `code_refs.path` 不存在
5. `symbol` 在对应文件中找不到
6. 索引条目缺失或失效
7. 反向链接缺失
8. 孤儿页面
9. `qmd` 索引状态落后

### Step 2: 读取结果并分级

将问题分成三层：

- 🔴 立即修复：确定性坏链、坏路径、坏 hash、坏 symbol
- 🟡 建议修复：归类异常、反向链接缺失、索引陈旧
- 🔵 可选优化：孤儿页面、低价值冗余线索

### Step 3: fix 模式处理

1. 默认只报告
2. 若用户指定 `--fix`：
   - 只修低风险且确定性的问题
   - 如：刷新 `source_hash`、删除不存在的辅助 code clue、补简单索引项
3. 若用户指定 `--fix --dry-run`：
   - 只预览拟修复项，不真正写入
4. 任何需要重新判断 capability / change 归属的问题，都不要在 `--fix` 中自动处理，应分流到 `/devwiki-refresh`

### Step 4: 输出报告

报告至少包含：

- 检查范围
- 问题总数
- 各等级问题明细
- 若启用 fix：哪些问题已修、哪些只预览、哪些必须人工处理
- 后续建议：
  - 能自动修但未执行：提醒带 `--fix`
  - 归类漂移：建议 `/devwiki-refresh`
  - 缺少结构化文档：建议 `/devwiki-feature-doc`

### Step 5: 记录日志

在 `wiki/log.md` 中追加：

- `check | report-only | <summary>`
- 或 `check | fix-applied | <summary>`

## Constraints

- **默认只报告**：不带 `--fix` 时不得修改 wiki
- **`--fix` 只修确定性低风险问题**：不能自动修归类和高影响判断
- **raw/ 只读**：不修改源文档
- **不得虚构 symbol**：找不到就是找不到，不能靠猜测补
- **检查结果可重复**：相同输入下结果应基本稳定

## Error Handling

- **wiki/ 不存在**：提示先运行 `/devwiki-init`
- **代码目录未配置**：跳过代码检查并说明范围受限
- **`zatools qmd ...` 不可用**：报告索引检查受限，但继续执行其他检查
- **fix 遇到中高风险项**：停止自动修复这些项，并建议 `/devwiki-refresh`
