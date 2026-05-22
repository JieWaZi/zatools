# DevWiki Graph 设计说明

## 背景

DevWiki 当前以 Markdown 页面作为知识事实源，核心页面分层为 capability、feature、workflow 和 troubleshooting。图谱能力的目标不是恢复已经移除的 `relations.yml`，也不是引入第二套关系事实源，而是从现有页面 frontmatter 派生出一个可视化导航索引。

第一版只实现一个入口：

```bash
zatools devwiki graph
```

该命令完整跑通 Markdown 解析、关系校验、缓存判断、静态页面生成、本地 HTTP 服务和浏览器打开流程。后续 ingest / maintain 落盘后，可以通过同一个命令的校验模式验证页面关系是否合法。

## 目标

- 从 `wiki/capabilities/*.md`、`wiki/features/*.md`、`wiki/workflows/*.md` 生成 graph 数据。
- 页面左侧展示 capability / feature / workflow 的关系图。
- 页面右侧按 capability / feature / workflow 展示选中节点涉及的文档。
- 支持按维度切换、搜索、节点邻居高亮和 warning 展示。
- Wiki 输入内容未变化时复用上次构建产物，不重复构建。
- 提供 `zatools devwiki graph --check`，供 Agent 在写入页面后校验关系结构。
- 保持 graph 为派生索引，页面 Markdown 仍是唯一权威事实源。

## 非目标

- 不恢复 `wiki/relations.yml`。
- 不新增独立 graph 数据库作为事实源。
- 不把 raw、sources、code_refs 或 troubleshooting 做成第一版图谱节点。
- 不在第一版加入 `depends_on`、`inherits`、`blocks`、`conflicts_with` 等新关系类型。
- 不让 query skill 依赖 graph 数据；query 仍按当前 DevWiki / qmd 检索规则工作。
- 不在前端解析 Markdown 全文，也不把右侧面板做成完整 Markdown 阅读器。
- 不实现后台守护进程、停止命令或多子命令形态。

## 命令行为

第一版只新增一个命令：

```bash
zatools devwiki graph
```

默认行为：

1. 识别 DevWiki 根目录，或使用 `--root <dir>` 指定。
2. 扫描 `wiki/capabilities/*.md`、`wiki/features/*.md`、`wiki/workflows/*.md`。
3. 计算输入内容哈希并读取 `.devwiki/graph/manifest.json`。
4. 如果 schema version、builder version 和 input hash 均未变化，复用上次构建产物。
5. 如果输入变化或用户指定 `--force`，重新生成 graph 数据和静态页面。
6. 启动本地 HTTP 静态服务。
7. 默认打开浏览器访问图谱页。
8. 命令保持前台运行，用户按 `Ctrl-C` 退出服务。

参数：

```bash
zatools devwiki graph --root .
zatools devwiki graph --host 127.0.0.1
zatools devwiki graph --port 0
zatools devwiki graph --no-open
zatools devwiki graph --force
zatools devwiki graph --check
```

- `--root`：指定 DevWiki 根目录，默认当前目录。
- `--host`：本地服务监听地址，默认 `127.0.0.1`。
- `--port`：本地服务端口，默认 `0`，表示自动选择可用端口。
- `--no-open`：只启动服务并打印 URL，不自动打开浏览器。
- `--force`：忽略缓存，强制重建 graph 产物。
- `--check`：只校验页面关系，不生成页面、不启动服务、不打开浏览器。

暂不新增 `build`、`serve`、`open` 子命令，避免第一版命令面过宽。

## 缓存设计

构建产物位于：

```text
.devwiki/graph/
  index.html
  graph.json
  manifest.json
  assets/
```

`manifest.json` 负责缓存判断，建议字段：

```json
{
  "schema_version": 1,
  "builder_version": 1,
  "input_hash": "sha256...",
  "built_at": "2026-05-22T12:00:00+08:00",
  "files": [
    {
      "path": "wiki/features/vip-failover.md",
      "size": 1234,
      "sha256": "..."
    }
  ],
  "outputs": {
    "graph": ".devwiki/graph/graph.json",
    "index": ".devwiki/graph/index.html"
  }
}
```

缓存判断以内容哈希为准：

- 输入文件列表、相对路径、文件大小和文件内容 SHA256 共同参与总体 input hash。
- mtime 不作为语义判断依据。
- 如果 schema version、builder version 或 input hash 任一项变化，必须重建。
- 如果本次构建失败，不打开旧产物，避免用户误以为当前 Wiki 已成功更新。

## Graph 数据结构

`graph.json` 是页面唯一数据输入。前端不直接读取 Markdown。

建议结构：

```json
{
  "schema_version": 1,
  "project": {
    "name": "项目名",
    "slug": "project-slug",
    "root": "/abs/path/to/devwiki"
  },
  "built_at": "2026-05-22T12:00:00+08:00",
  "nodes": [
    {
      "id": "feature:vip-failover",
      "type": "feature",
      "slug": "vip-failover",
      "title": "VIP 接管",
      "summary": "说明该功能解决的问题",
      "status": "active",
      "confidence": "medium",
      "path": "wiki/features/vip-failover.md",
      "search_terms": ["vip", "接管", "failover"]
    }
  ],
  "edges": [
    {
      "id": "capability:ha->feature:vip-failover",
      "type": "contains",
      "source": "capability:ha",
      "target": "feature:vip-failover",
      "label": "包含功能",
      "sources": [
        "wiki/capabilities/ha.md",
        "wiki/features/vip-failover.md"
      ]
    }
  ],
  "documents": {
    "feature:vip-failover": {
      "type": "feature",
      "path": "wiki/features/vip-failover.md",
      "title": "VIP 接管",
      "summary": "说明该功能解决的问题"
    }
  },
  "warnings": []
}
```

节点 ID 规则：

```text
capability:<slug>
feature:<slug>
workflow:<slug>
```

同一 slug 可在不同类型中存在；同一类型内 slug 必须唯一。

## 关系模型

第一版只支持当前 skill 模板已经定义的字段，不新增 graph 专用格式。

Capability 模板字段：

```yaml
features: []
related_capabilities: []
```

Feature 模板字段：

```yaml
capabilities: []
workflow: ""
related_features: []
```

Workflow 模板字段：

```yaml
features: []
related_workflows: []
```

输出边类型收敛为三类：

| 边类型 | 方向 | 含义 |
|---|---|---|
| `contains` | capability -> feature | 能力覆盖哪些功能 |
| `implemented_by` | feature -> workflow | 功能由哪个 workflow 解释实现路径 |
| `related` | 同层为主 | 需要互相参照的相关页面 |

来源字段映射：

| 来源 | 输出 |
|---|---|
| capability `features` | `contains` |
| feature `capabilities` | `contains` |
| feature `workflow` | `implemented_by` |
| workflow `features` | `implemented_by` |
| `related_capabilities` | `related` |
| `related_features` | `related` |
| `related_workflows` | `related` |

关系归一化规则：

- 页面可从任一方向声明关系。
- 输出 graph 时同一关系只生成一条边。
- 如果双向页面都声明了同一关系，edge 的 `sources` 记录两个来源文件。
- `related` 在 UI 中按无向关系展示；数据中仍保留稳定 source / target，便于调试。
- 正文中的 `[[...]]` 链接第一版不作为 graph 边来源，避免把导航、说明、历史引用误判为强关系。

引用格式必须兼容当前 skill 生成习惯：

```yaml
features:
  - "vip-failover"
  - "[[vip-failover]]"
  - "wiki/features/vip-failover.md"
```

`workflow` 支持字符串或列表；字符串按单个 workflow 处理。

## 校验规则

`zatools devwiki graph --check` 只解析和校验页面，不生成静态页面，不启动 HTTP 服务。

Error，必须返回非零退出码：

| 类型 | 示例 |
|---|---|
| YAML frontmatter 解析失败 | `---` 未闭合或 YAML 格式错误 |
| 同类型重复 slug | 两个 feature 都是 `slug: vip-failover` |
| 关系字段类型错误 | `features: abc` 而不是数组 |
| `workflow` 指向不存在 workflow | feature 写了不存在的 workflow |
| `features` 指向不存在 feature | capability 或 workflow 指向不存在 feature |
| `capabilities` 指向不存在 capability | feature 指向不存在 capability |

Warning，不阻断：

| 类型 | 处理 |
|---|---|
| 缺少 slug | 从文件名推导 slug，并提示 |
| 缺少 title | 使用 slug 展示，并提示 |
| 缺少 summary | 摘要为空，并提示 |
| 单向关系缺少反向字段 | 提示维护双向关系 |
| related 指向不存在页面 | 提示断链，不生成边 |
| 孤立节点 | 提示补充关系 |
| status 或 confidence 缺失 | 使用默认展示值，并提示 |

普通 `zatools devwiki graph` 会执行同一套校验：

- 遇到 error：构建失败，不启动页面。
- 遇到 warning：继续构建，写入 `graph.json.warnings`，页面顶部展示 warning 数量。

示例输出：

```text
DevWiki graph check failed

ERROR wiki/features/vip-failover.md
  workflow points to missing workflow: workflow-vip-failover

WARNING wiki/capabilities/ha.md
  feature relation is not declared back from wiki/features/vip-failover.md

Summary:
  capabilities: 3
  features: 12
  workflows: 8
  errors: 1
  warnings: 1
```

## 页面交互

页面通过本地 HTTP 服务访问：

```text
http://127.0.0.1:<port>/
```

布局：

```text
顶部工具栏：搜索 / 维度切换 / 布局切换 / 关系深度 / 状态提示
左侧 Graph 画布
右侧文档面板
```

第一版交互：

- 默认展示 capability、feature、workflow 三层节点。
- 支持维度切换：
  - 全部
  - Capability
  - Feature
  - Workflow
- 搜索框按 `title`、`slug`、`summary`、`search_terms` 过滤。
- 点击节点后：
  - 当前节点高亮；
  - 一跳邻居高亮；
  - 非相关节点淡化；
  - 右侧面板展示该节点和相关文档。
- 关系深度：
  - 直接关系
  - 二跳关系
- 布局：
  - 分层布局：Capability 在左或上，Feature 在中间，Workflow 在右或下；
  - 力导向布局：用于查看复杂关系。

视觉编码：

| 类型 | 样式 |
|---|---|
| Capability | 蓝色系节点 |
| Feature | 绿色系节点 |
| Workflow | 橙色系节点 |
| `contains` | 实线 |
| `implemented_by` | 实线箭头 |
| `related` | 虚线 |

右侧文档面板：

- 未选节点时展示项目总览：节点数量、边数量、构建时间、warning 数量。
- 选中 Capability 时展示覆盖 Feature、相关 Capability、通过 Feature 间接关联的 Workflow。
- 选中 Feature 时展示所属 Capability、实现 Workflow、相关 Feature。
- 选中 Workflow 时展示支撑 Feature、上层 Capability、相关 Workflow。
- 只展示标题、摘要、状态、置信度、路径和关系说明。
- 不渲染 Markdown 全文。
- 第一版不要求点击路径打开编辑器；路径展示和复制即可。

前端技术：

- 使用 Cytoscape.js。
- 不引入 React / Vue / Node 构建链。
- 静态资源由 Go embed 打包进 zatools，运行时写入 `.devwiki/graph/`。
- Cytoscape.js 以 vendored 静态资源进入仓库，并保留 license 说明，避免离线环境依赖 CDN。

## Go 包边界

建议新增内部包：

```text
internal/devwiki/graph/
```

职责拆分：

```text
internal/devwiki/graph/parser.go      # Markdown frontmatter 解析、slug 归一化
internal/devwiki/graph/model.go       # Node、Edge、Graph、Manifest、Issue 等结构
internal/devwiki/graph/builder.go     # 从页面集合构建 graph
internal/devwiki/graph/cache.go       # 输入 hash、manifest 读取和比较
internal/devwiki/graph/check.go       # 校验规则和 issue 分级
internal/devwiki/graph/assets.go      # 静态资源写入
internal/devwiki/graph/server.go      # 本地 HTTP 服务
```

CLI 只负责参数解析和调用 app/service 层。若实现时遵循现有层次，应优先添加：

```text
internal/app/devwikiapp/graph.go
internal/cli/devwiki/command.go
```

用户可见文案必须集中到 `internal/ui/i18n.go`，不要散落在 graph 包或 CLI 包中。

## Skill 规则更新

代码实现后，再更新 DevWiki skill 文档。

Ingest 规则：

- Proposal 阶段说明新增或更新页面会产生哪些 graph 关系。
- 落盘后，如果修改了 capability、feature 或 workflow 页面，必须执行：

```bash
zatools devwiki graph --check
```

- 校验失败时，不得直接宣称完成；需要修复错误，或回到 proposal 请求用户确认。

Maintain 规则：

- 维护范围涉及关系、重命名、拆分、合并、断链、入口错误时，必须执行：

```bash
zatools devwiki graph --check
```

- 单向关系 warning 可以作为维护建议；是否自动补齐仍遵守 maintain 的 proposal / confirmed_write 规则。

Query 规则：

- 第一版不要求 query 读取 graph 数据。
- 后续如果 graph.json 稳定，再评估 query 是否读取 graph 作为快速路由辅助。

## 测试策略

Go 单元测试：

- 使用当前 capability / feature / workflow 模板字段构建 graph 成功。
- `slug`、`[[slug]]`、`wiki/features/slug.md` 三种引用格式都能解析。
- `workflow` 字符串和列表都能解析。
- 双向声明同一关系时只生成一条边，并记录多个 source。
- 重复 slug 返回 error。
- 主链关系指向不存在页面返回 error。
- related 指向不存在页面返回 warning。
- 缺少反向字段返回 warning。
- `--check` 不写 `.devwiki/graph`，不启动 HTTP 服务。
- 输入内容未变化时复用 manifest，不重建。
- 输入内容变化、schema version 变化、builder version 变化或 `--force` 时重建。

CLI / app 测试：

- `zatools devwiki graph --check --root <fixture>` 成功和失败路径。
- `zatools devwiki graph --root <fixture> --no-open --port 0` 能生成产物并启动服务。
- 构建失败时不打开旧产物。
- 非 DevWiki root 或空三层页面时输出明确错误。

前端验证：

- `graph.json` 加载失败时页面提示重新执行 `zatools devwiki graph --force`。
- 空搜索结果不破坏原始 graph 数据。
- 点击 capability / feature / workflow 时右侧分组符合设计。
- warning 数量可见，并可展开查看。

## 验收标准

- 在一个 DevWiki fixture 中执行 `zatools devwiki graph`，能打开本地页面并看到三层图谱。
- 第二次执行且输入未变化时复用缓存，不重新构建。
- 修改一个 feature 页面后再次执行会重新构建。
- `zatools devwiki graph --check` 能发现断链、重复 slug 和 YAML 错误。
- Agent 按现有 ingest 模板生成的页面能被 graph builder 正确识别。
- 所有新增用户可见文案集中在 `internal/ui/i18n.go`。
- 相关 Go 测试通过。
- 实现完成后，ingest / maintain skill 增加落盘后调用 `zatools devwiki graph --check` 的规则。
