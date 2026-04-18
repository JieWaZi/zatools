---
name: "devwiki-edit"
description: "当用户已经明确给出要修改哪类 DevWiki 内容时使用，适用于定点更新 wiki 页面、补充元数据、添加新的 raw 资料入口，或执行已确认的结构化编辑。"
argument-hint: "[编辑请求]"
---

# /devwiki-edit

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - 涉及写入、重分类或破坏性操作时，再读 `references/mutation-safety.md`
> - 涉及代码追踪、代码归因或实现核对时，再读 `references/code-tracing.md`


> 对 DevWiki 做定点编辑。`edit` 适合用户已经知道要改什么的场景；如果是知识漂移或归类错误，优先考虑 `/devwiki-refresh`。

## Inputs

- `request`：明确的编辑请求
- 可选：本地文件路径、URL、目标 wiki 页面路径

## Outputs

- 更新后的 `wiki/` 页面或新增的 `raw/` 资料入口
- 更新后的 `wiki/index.md`
- 更新后的 `wiki/log.md`
- 如有必要，附带后续建议：`/devwiki-ingest` 或 `/devwiki-refresh`

## DevWiki Interaction

### Reads

- 用户指定的 `wiki/` 页面
- `wiki/index.md`
- `raw/` 目录中相关源资料
- `config/project.yaml`（若需要代码目录信息）

### Writes

- CREATE / EDIT `wiki/**/*.md`
- 可在确认后向 `raw/` 新增文件
- EDIT `wiki/index.md`
- APPEND `wiki/log.md`


## Workflow

### Step 1: 解析用户意图

把请求分成三类：

1. 编辑已有 wiki 页面
2. 往 `raw/` 新增资料入口
3. 删除或替换错误内容

如果请求本质上是在纠正漂移、失效路径、错误归类，应提醒用户这可能更适合 `/devwiki-refresh`。

### Step 2: 确定修改边界

1. 锁定具体目标页面或目标目录
2. 如果用户只给了模糊目标，先 ask for confirmation，不要盲改
3. 对涉及多个页面的改动，先列出将受影响的页面
4. 对会影响结构化关系的改动，先检查是否需要同步反向链接或索引

### Step 3: 执行编辑

1. **编辑 wiki 页面**：
   - 只修改用户要求的字段或章节
   - 保持模板结构，不随意重写整个页面
   - 若新增正向链接，要同步维护反向关系
2. **向 raw/ 新增资料**：
   - 允许新增文件或入口记录
   - 新资料不会自动进入 wiki，后续需建议 `/devwiki-ingest`
3. **删除或替换内容**：
   - 中高风险删除动作必须等待确认
   - 涉及 capability / change 主归属的编辑，应先确认再写

### Step 4: 更新导航与日志

1. 如页面标题、slug、分类发生变化，同步更新 `wiki/index.md`
2. 在 `wiki/log.md` 追加：
   - `edit | wiki-updated | <summary>`
   - 或 `edit | raw-added | <summary>`
3. 如果本次编辑写入了 `wiki/` 或向 `raw/` 新增了文件，继续执行：

```bash
zatools qmd update
zatools qmd status
```

4. 如果后续任务马上依赖 `zatools qmd query` 做更高质量语义召回，且 `status` 显示还有 pending embeddings，再询问用户是否继续执行：

```bash
zatools qmd embed
```

### Step 5: 给出后续建议

- 新增了 raw 资料：建议 `/devwiki-ingest`
- 发现旧知识和现状不一致：建议 `/devwiki-refresh`
- 发现需要更系统的能力文档：建议 `/devwiki-feature-doc`

## Constraints

- **不要扩写用户没要求的范围**
- **raw/ 中已有文件默认不改写**：新增可以，覆盖和删除必须确认
- **结构化页面保持模板边界**
- **涉及 capability / change 主归属的修改要确认**
- **链接关系要同步维护**

## Error Handling

- **目标页面不存在**：先告诉用户，再确认是否新建
- **请求过于模糊**：先 ask for confirmation，不要猜
- **需要跨多页高影响改动**：暂停直接编辑，建议改走 `/devwiki-refresh`
