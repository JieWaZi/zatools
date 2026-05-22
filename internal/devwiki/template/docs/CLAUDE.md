# DevWiki — 运行时规则

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，只定义项目级目录、链接规范、检索顺序与 workflow 约束。

## 目录结构

```text
./
├── README.md
├── {{RUNTIME_FILE}}
├── config/
│   ├── project.yaml
│   └── search.yaml
├── raw/
│   ├── requirements/
│   ├── designs/
│   ├── features/
│   └── tests/
└── wiki/
    ├── index.md
    ├── glossary.md
    ├── log.md
    ├── capabilities/
    ├── features/
    ├── workflows/
    ├── troubleshooting/
    └── outputs/
```

本目录就是 DevWiki 文档库根目录。代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录。Agent 在代码库内使用 `devwiki-query` 或 `devwiki-code-to-doc` 前，必须先阅读本文件；查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源，生成的新 Wiki 文件也必须写回本目录。

关键路径：

- `wiki/capabilities/{slug}.md`
- `wiki/features/{slug}.md`
- `wiki/workflows/{slug}.md`
- `wiki/troubleshooting/{slug}.md`
- `wiki/log.md`
- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

## Wiki 目录

| 目录 | 文件名 |
|------|--------|
| `wiki/capabilities/` | `{slug}.md` |
| `wiki/features/` | `{slug}.md` |
| `wiki/workflows/` | `{slug}.md` |
| `wiki/troubleshooting/` | `{slug}.md` |

具体页面模板、字段结构、页面边界和证据写入规则由对应 DevWiki skill 和 `references/` 维护。本文不重复定义各类页面边界、页面模板、frontmatter 字段、`sources` 与 `code_refs` 结构，避免和 skill 内的模板规则冲突。

## 链接语法

内部 Wiki 页面使用 Obsidian wikilink：

```markdown
[[slug]]
[[user-management]]
```

slug 使用小写连字符，发布后应尽量保持稳定。外部 raw 文件、代码文件、接口入口和测试入口不是 Wiki 页面，不使用 wikilink。

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| capability/A 写入 `features: [[feature-F]]` | feature/F 的 `capabilities` 追加 A |
| capability/A 写入 `related_capabilities: [[capability-B]]` | capability/B 的 `related_capabilities` 追加 A |
| feature/F 写入 `workflow: [[workflow-W]]` | workflow/W 的 `features` 追加 F |
| feature/F 写入 `related_features: [[feature-G]]` | feature/G 的 `related_features` 追加 F |
| troubleshooting/T 写入 `features: [[feature-F]]` | feature/F 的排障入口摘要追加 T |

外部 raw 文件、代码文件、接口入口和测试入口没有反向页面要求。

## 检索顺序

- 能力问题：`capabilities → features`
- 功能问题：`features → capabilities`
- 代码问题：`workflows → features → rg`
- 排障问题：`troubleshooting → workflows → features`

检索通道按成本升档：本地搜索、`zatools qmd search`、`zatools qmd query`。`qmd` 只是召回工具，不是真相源。

## Workflow 约束

- `raw/` 只读
- facts 与 inference 必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据
- 页面写入和证据字段更新必须遵守对应 DevWiki skill 的模板和引用规则
- 项目知识任务先由 `devwiki-project-router` 判断意图、身份、证据需求和检索边界，再路由到 `devwiki-ingest`、`devwiki-maintain`、`devwiki-query`、`devwiki-code-to-doc` 或 `devwiki-qmd-sync`
- 待确认问题优先通过对话和用户对齐
- 中高风险写入必须先给 proposal，再落盘
- 多个候选归属、拆分边界不清、页面合并/重命名、代码证据不足时必须先问用户

## 操作说明

- 使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库并安装 skills
- 使用 `zatools devwiki link` 将已有 DevWiki 文档库关联到代码库
- 使用 `devwiki-project-router` 作为项目知识任务的默认总入口
- 使用 `devwiki-ingest` 吸收 raw 文档并生成或更新 Wiki 页面、术语和入口导航
- 使用 `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失、入口错误和 query 污染
- 使用 `devwiki-query` 查询 Wiki、raw、代码线索、设计意图和排障知识
- 使用 `devwiki-code-to-doc` 从代码、接口、配置项、日志或路由反向生成或更新 workflow 页面

这份运行时规则应保持稳定，只在 DevWiki 的项目目录、链接规范或工作流约束发生真实变化时修改。
