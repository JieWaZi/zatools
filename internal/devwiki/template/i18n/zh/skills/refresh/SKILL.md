---
name: "devwiki-refresh"
description: "当现有 DevWiki 知识与 raw 文档、capability-feature 关系、代码路径、symbol 或测试/接口入口发生漂移时使用。"
argument-hint: "[漂移范围或问题描述]"
---

# /devwiki-refresh

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 修复 DevWiki 知识漂移与用户反馈错误。`refresh` 默认 proposal-first，不应静默改写高影响知识。

## Inputs

- `scope`（可选）：漂移范围或问题描述，例如“用户权限 feature 漂移”或“所有失效 code refs”
- `wiki/capabilities/*.md`
- `wiki/features/*.md`
- `raw/*/*.md`
- `config/project.yaml`

## Outputs

- 一份 `refresh proposal`
- 待修正的 capabilities / features 列表
- 失效的 `code_refs`、过期的 `sources.hash`、缺失的 `symbol`、错误路径、失效 capability-feature 关系
- 用户确认后的修复结果
- 更新后的 `wiki/index.md`
- 追加到 `wiki/log.md` 的 refresh 日志

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录
- `wiki/capabilities/*.md` — 检查能力总结、feature 关联与边界
- `wiki/features/*.md` — 检查 `sources`、`api_entries`、`code_refs`、`test_refs`
- `wiki/index.md` — 校验页面存在性
- `raw/*/*.md` — 与当前 raw 来源对比
- 本地代码目录 — 检查 code-ref 路径和 `symbol` 是否仍存在

### Writes

- 默认 proposal-first，不立即写入
- 只有用户确认后才允许：
  - EDIT `wiki/capabilities/*.md`
  - EDIT `wiki/features/*.md`
  - EDIT `wiki/index.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: 识别 refresh 范围

1. 判断这是全局 refresh，还是只修一个 capability、一个 feature 或一组 source path
2. 将问题归类为：
   - `sources.hash` 过期
   - raw 来源缺失
   - `code_refs.path` 失效
   - `symbol` 缺失
   - `api_entries` 或 `test_refs` 过期
   - capability-feature 关系错误
   - capability 总结漂移
3. 如果用户已经指出了具体错误，优先围绕该错误收敛，不要盲扫全库

### Step 2: 做确定性漂移检查

按 `references/zatools-qmd.md` 的阶梯召回规则执行，**默认 local-first**：

1. 重新核对 feature 页里的 `sources.path` 与 `sources.hash`
2. 检查 feature 页的 `code_refs`、`api_entries`、`test_refs` 是否仍指向有效入口
3. 若配置了代码目录，检查每个 `symbol` 是否仍存在
4. 核对 capability-feature 的双向链接
5. 需要重新召回时，按阶梯升档：
   - 先本地 `grep` / 文件搜索
   - 不足再升档 `zatools qmd search`
   - 只有概念级召回需求时才考虑 `zatools qmd query`；无 GPU/加速时按共享规则走硬性 fallback
6. 将发现分成三类：
   - 确定性坏链
   - 高概率漂移
   - 低置信推断

### Step 3: 重新召回候选修复方案

1. 对缺失的 raw 来源，在 `raw/` 下搜索是否被改名、移动或替换
2. 对失效 code refs，在代码目录里搜索替代路径和 symbol
3. 对 capability 漂移，可回看关联 feature 和 raw 资料
4. 对 feature 漂移，可回看 source、入口和有限代码证据
5. 如果某个 feature 明显已经无法增量修复，应建议 `/devwiki-feature-doc`

### Step 4: 生成 refresh proposal

`refresh proposal` 必须包含：

- 哪些问题是确定性坏链
- 哪些问题是高概率漂移
- 哪些问题仍是低置信推断
- 建议动作：
  - 更新 `sources.hash`
  - 替换 `sources.path`
  - 替换 `code_refs.path`
  - 移除失效 `symbol`
  - 修复 capability-feature 关系
  - 收紧或简化过时总结

并按风险分层：

- 低风险：刷新 `sources.hash`、移除明显失效的辅助线索、更新索引
- 中风险：替换 `code_refs` / `api_entries` / `test_refs`，修复明显失效的 capability-feature 链接
- 高风险：改变 capability 边界、新建或删除 feature、替换主 code refs

### Step 5: 等待用户确认

1. 中高风险修复必须等待确认
2. 如果多个候选路径或 capability 映射都合理，不要静默决定
3. 改为向用户提 1 到 3 个具体问题
4. 只有提案得到确认后，才允许落盘

### Step 6: 应用修复并记录结果

在用户确认后：

1. 更新对应的 capabilities / features
2. 更新 `wiki/index.md`
3. 在 `wiki/log.md` 追加 `refresh | proposal-applied`
4. 落盘成功后执行：

```bash
zatools qmd update
zatools qmd status
```

5. 如果当前任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

6. 如果暴露出更多确定性坏链，建议继续执行 `/devwiki-check`

## Constraints

- **raw/ 只读**：不得修改原始资料
- **默认 proposal-first**：不要静默落盘高影响修复
- **不得虚构 symbol**：没有核实过的函数、文件、symbol 不得改写
- **中高风险必须确认**：尤其是 capability 边界、feature 改挂、主 `code_refs`
- **事实与推断要分开**：确定性坏链、高概率漂移、低置信推断必须分层报告
- **允许承认不确定**：证据弱时不要强修

## Error Handling

- **wiki 基本为空**：提示用户先执行 `/devwiki-init` 或 `/devwiki-ingest`
- **代码目录缺失**：允许只做 wiki/raw 层 refresh，但要说明未核对代码
- **`zatools qmd ...` 不可用**：回退到本地搜索，不中断 refresh
- **source 目标已删除**：报告为确定性坏链，并在 `raw/` 下搜索替代文件
- **多个候选都合理**：停止扩散并向用户提问
