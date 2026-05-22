# DevWiki — 运行时规则

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，只定义项目级目录、链接规范、检索顺序与 workflow 约束。

---

## 目录结构

```text
./
├── README.md                    ← 项目说明、使用方式与工作流入口
├── {{RUNTIME_FILE}}             ← 当前 agent 对应的运行时规则副本
├── config/
│   ├── project.yaml             ← 当前项目默认配置，如项目 slug、代码仓列表、语言与 agent
│   └── search.yaml              ← qmd collection 配置
├── raw/
│   ├── requirements/            ← 原始需求文档
│   ├── designs/                 ← 原始设计文档
│   ├── features/                ← 原始功能说明
│   └── tests/                   ← 测试方案与测试记录
└── wiki/
    ├── index.md                 ← Wiki 总目录与导航入口
    ├── glossary.md              ← 术语表
    ├── log.md                   ← 追加式操作日志
    ├── capabilities/            ← 系统能力
    ├── features/                ← 具体功能设计
    ├── workflows/               ← 工程定位、调用链与代码线索
    ├── troubleshooting/         ← 排障知识
    └── outputs/                 ← ingest / maintain / query / code-to-doc / qmd-sync 报告
```

本目录就是 DevWiki 文档库根目录。代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录。Agent 在代码库内使用 `devwiki-query` 或 `devwiki-code-to-doc` 前，必须先阅读本文件；查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源，生成的新 Wiki 文件也必须写回本目录。

`raw/` 是事实来源层，`wiki/` 是结构化知识层。`config/search.yaml` 只保存检索配置，不替代事实内容。

---

## Wiki 目录

DevWiki 的人工维护知识页按目录组织。具体页面模板、字段结构、页面边界和证据写入规则由对应 DevWiki skill 和 `references/` 维护。

| 目录 | 文件名 |
|------|--------|
| `wiki/capabilities/` | `{slug}.md` |
| `wiki/features/` | `{slug}.md` |
| `wiki/workflows/` | `{slug}.md` |
| `wiki/troubleshooting/` | `{slug}.md` |

生成目录不是人工知识实体：

- `wiki/outputs/`

本文不重复定义各类页面边界、页面模板、frontmatter 字段、`sources` 与 `code_refs` 结构，避免和 skill 内的模板规则冲突。

关键显式路径：

- `wiki/capabilities/{slug}.md`
- `wiki/features/{slug}.md`
- `wiki/workflows/{slug}.md`
- `wiki/troubleshooting/{slug}.md`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `wiki/capabilities/`
- `wiki/features/`
- `wiki/workflows/`
- `wiki/troubleshooting/`
- `wiki/outputs/`
- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

---

## 链接语法

所有内部链接使用 Obsidian wikilink：

```markdown
[[slug]]
[[user-management]]
[[permission-assignment]]
[[user-group-management]]
```

命名规范：

- 全小写
- 使用连字符分隔
- 不带空格
- slug 一旦发布应尽量保持稳定

外部 raw 文件、代码文件、接口入口和测试入口不是 Wiki 页面，不使用 wikilink。它们的证据记录方式以对应 skill 的写入规则为准。

---

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| capability/A 写入 `features: [[feature-F]]` | feature/F 的 `capabilities` 追加 A |
| capability/A 写入 `related_capabilities: [[capability-B]]` | capability/B 的 `related_capabilities` 追加 A |
| feature/F 写入 `workflow: [[workflow-W]]` | workflow/W 的 `features` 追加 F |
| feature/F 写入 `related_features: [[feature-G]]` | feature/G 的 `related_features` 追加 F |
| troubleshooting/T 写入 `features: [[feature-F]]` | feature/F 的排障入口摘要追加 T |

以下外部对象没有反向页面要求：

- `raw/*`
- 代码文件和 symbol
- 接口入口
- 测试入口

---

## 检索顺序

按用户意图选择起点：

- 能力问题：`capabilities → features`
- 功能问题：`features → capabilities`
- 代码问题：`workflows → features → rg`
- 排障问题：`troubleshooting → workflows → features`

检索通道按「成本 / 速度由低到高」阶梯升档，详见 skill 内的 `references/zatools-qmd.md`：

1. 本地 `grep` / 文件搜索
2. `zatools qmd search`
3. `zatools qmd query`

任一档命中 top-K 且置信足够即停；只有当问题落到实现现实、代码归属，或文档证据仍不足时，才对 top-K 代码候选做一次定向本地排查。

---

## Workflow 约束

### 事实约束

- `raw/` 是只读原始资料
- facts 与 inference 必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据
- 页面写入和证据字段更新必须遵守对应 DevWiki skill 的模板和引用规则

### 路由约束

- 项目知识任务先由 `devwiki-project-router` 判断意图、身份、证据需求和检索边界，再路由到 `devwiki-ingest`、`devwiki-maintain`、`devwiki-query`、`devwiki-code-to-doc` 或 `devwiki-qmd-sync`

### 确认策略

- 待确认问题优先通过对话和用户对齐
- 中高风险写入必须先给 proposal，再落盘
- 多个候选归属、拆分边界不清、页面合并/重命名、代码证据不足时必须先问用户

---

## 操作说明

- 使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库并安装 skills
- 使用 `zatools devwiki link` 将已有 DevWiki 文档库关联到代码库，并在代码库中写入 AGENTS/CLAUDE 关联规则
- 如需下载 qmd models，初始化完成后可在 DevWiki 工作区内手动执行 `zatools qmd download --root .`
- 对已有工作区补做或修复 qmd collection 注册、索引刷新与状态检查时，使用 `devwiki-qmd-sync`
- 使用 `devwiki-project-router` 作为项目知识任务的默认总入口
- 使用 `devwiki-ingest` 吸收 raw 文档并生成或更新 Wiki 页面、术语和关系
- 使用 `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失、关系错误和 query 污染
- 使用 `devwiki-query` 查询 Wiki、raw、代码线索、设计意图和排障知识
- 使用 `devwiki-code-to-doc` 从代码、接口、配置项、日志或路由反向生成或更新 workflow 页面

这份运行时规则应保持稳定，只在 DevWiki 的项目目录、链接规范或工作流约束发生真实变化时修改。
