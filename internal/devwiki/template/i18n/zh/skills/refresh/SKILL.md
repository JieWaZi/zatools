---
name: "devwiki-refresh"
description: "当 DevWiki 已有知识与当前 raw 文档、代码路径、符号位置、能力归类发生漂移时使用，尤其适用于用户纠错、source_hash 失配、code refs 失效、symbol 消失和分类判断需要修正的场景。"
argument-hint: "[漂移范围或问题描述]"
---

# /devwiki-refresh

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 处理 DevWiki 的知识漂移与用户纠错。`refresh` 默认先做检测和修正提案，不直接改写高影响内容。

## Inputs

- `scope`（可选）：漂移范围或问题描述，例如“用户权限相关 capability 漂移”或“全部失效 code refs”
- `wiki/documents/**/*.md`
- `wiki/capabilities/*.md`
- `wiki/changes/*.md`
- `raw/*/*.md`
- `config/project.yaml`

## Outputs

- 一份 `refresh proposal`
- 待修正的 documents / capabilities / changes 列表
- 待修正的 `code refs`、坏链、失效 `source_hash`、缺失 `symbol` 列表
- 用户确认后写入的修正结果
- 更新后的 `wiki/index.md`
- 追加到 `wiki/log.md` 的 refresh 记录

## DevWiki Interaction

### Reads

- `config/project.yaml` — 获取主代码目录
- `wiki/documents/**/*.md` — 检查 `source_path`、`source_hash`
- `wiki/capabilities/*.md` — 检查能力归类、documents、code refs
- `wiki/changes/*.md` — 检查 change 分类和关联对象
- `wiki/index.md` — 交叉核对页面存在性
- `raw/*/*.md` — 回源对比当前原始文档
- 本地代码目录 — 检查 `code refs` 路径和 `symbol` 是否仍有效

### Writes

- 默认只生成提案，不直接写
- 用户确认后才允许：
  - EDIT `wiki/documents/**/*.md`
  - EDIT `wiki/capabilities/*.md`
  - EDIT `wiki/changes/*.md`
  - EDIT `wiki/index.md`
  - APPEND `wiki/log.md`


## Workflow

### Step 1: 识别 refresh 范围

1. 确定是全局 refresh，还是针对某个 capability / change / 文档路径
2. 区分问题类型：
   - `source_hash` 失配
   - `source_path` 失效
   - `code refs` 路径不存在
   - `symbol` 消失
   - capability 归类错误
   - change 的 `new / modify` 判断错误
3. 如果用户给的是具体纠错意见，优先围绕该意见收敛，而不是全库乱扫

### Step 2: 运行确定性漂移检查

1. 对 document 镜像页重新核对 `source_path` 与 `source_hash`
2. 对 capability 与 change 中的 `code refs` 做路径检查
3. 若代码目录存在，再核对 `symbol` 是否还能在对应文件中找到
4. 先执行 `zatools qmd status`
5. 若 `zatools qmd status` 正常，优先用 `zatools qmd query` 重召回 `wiki / raw / code`，再用 `zatools qmd get` / `zatools qmd multi-get` 压缩修正范围
6. 记录三类结果：
   - 确定性坏链
   - 高概率漂移
   - 低置信推断

### Step 3: 重新召回候选并生成修正方向

1. 对失效 document 回到 `raw/` 查同源文档是否改名、迁移或内容已更新
2. 对失效 code refs 在代码目录中重新搜索候选文件和 symbol
3. 对 capability 归类可疑的页面，重新比对其 documents、changes、code refs
4. 对 change 分类可疑的页面，重新评估它更像 `new`、`modify` 还是 `unclear`
5. 如果文档描述已明显落后于代码现状，可在建议中分流到 `/devwiki-feature-doc`

### Step 4: 生成 refresh proposal

`refresh proposal` 至少包含：

- 哪些问题是确定性坏链
- 哪些问题是高概率漂移
- 哪些问题仍是低置信推断
- 建议如何修：
  - 更新 `source_hash`
  - 更换 `source_path`
  - 替换 `code refs.path`
  - 移除不存在的 `symbol`
  - 修正 capability 归类
  - 修正 change 分类

并按风险分层：

- 低风险：刷新 `source_hash`、删除明显失效的辅助 code clue、更新索引
- 中风险：替换 `code refs`、追加更可信的候选路径
- 高风险：修改 capability 归类、修改 change 分类、改写主 code refs

### Step 5: 等待用户确认

1. 中高风险修正必须等待用户确认
2. 如果重新召回后仍然存在多个候选路径、多个能力归属、多个 change 分类，就不要自行拍板
3. 这时向用户提 1 到 3 个具体问题
4. 只有提案确认后，才能正式写回 wiki

### Step 6: 应用修正并记录结果

确认后：

1. 更新对应的 documents / capabilities / changes
2. 更新 `wiki/index.md`
3. 在 `wiki/log.md` 追加 `refresh | proposal-applied`
4. 落盘成功后执行：

```bash
zatools qmd update
zatools qmd status
```

5. 若当前任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

6. 若修正过程中暴露出更多确定性坏链，建议用户后续执行 `/devwiki-check`

## Constraints

- **raw/ 只读**：不得改动原始文档
- **默认先提案**：`refresh` 先出修正提案，不直接落盘高影响修改
- **不得虚构 symbol**：找不到的函数、类、文件不能硬写回去
- **中高风险必须确认**：尤其是 capability 归类、change 分类、主 `code refs`
- **事实和推断分离**：确定性坏链、高概率漂移、低置信推断必须分开报告
- **可承认不确定**：证据不足时不要硬修

## Error Handling

- **wiki 基本为空**：提示先执行 `/devwiki-init` 或 `/devwiki-ingest`
- **代码目录未配置**：允许只修文档层漂移，但要说明未核对代码
- **`zatools qmd ...` 不可用**：回退到本地搜索，不中断 refresh
- **source_path 指向文件已删除**：报告为确定性坏链，并回源查是否有替代文档
- **多个候选都像**：停止扩散，转而向用户提问
