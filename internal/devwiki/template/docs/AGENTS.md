# DevWiki — 运行时规则

> 面向研发场景的 DevWiki，由 Claude Code 与 Codex 共同驱动。
> 本文件是 wiki 的运行时入口，只定义项目级目录、链接规范、检索顺序与 workflow 约束。

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
    ├── topics/                  ← 主题边界、功能规则与实现入口
    ├── workflows/               ← 工程定位、调用链与代码线索
    ├── troubleshooting/         ← 排障知识
    └── outputs/                 ← ingest / maintain / query / code-to-doc 报告
```

本目录就是 DevWiki 文档库根目录。代码库通过 AGENTS/CLAUDE 中的托管关联块指向本目录。Agent 在代码库内使用 `devwiki-query`、`devwiki-code` 或 `devwiki-code-to-doc` 前，必须先阅读本文件；查询以本目录的 `wiki/`、`raw/`、`config/search.yaml` 为知识来源，生成的新 Wiki 文件也必须写回本目录。

`raw/` 是事实来源层，`wiki/` 是结构化知识层。`config/search.yaml` 只保存检索配置，不替代事实内容。

## Wiki 目录

DevWiki 的人工维护知识页按目录组织。具体页面模板、字段结构、页面边界和证据写入规则由对应 DevWiki skill 和 `references/` 维护。

| 目录 | 文件名 |
|------|--------|
| `wiki/topics/` | `{slug}.md` |
| `wiki/workflows/` | `{slug}.md` |
| `wiki/troubleshooting/` | `{slug}.md` |

生成目录不是人工知识实体：

- `wiki/outputs/`

关键显式路径：

- `wiki/topics/{slug}.md`
- `wiki/workflows/{slug}.md`
- `wiki/troubleshooting/{slug}.md`
- `wiki/index.md`
- `wiki/glossary.md`
- `wiki/log.md`
- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

## 链接语法

所有内部链接使用 Obsidian wikilink：

```markdown
[[slug]]
[[user-management]]
```

slug 使用小写连字符，发布后应尽量保持稳定。外部 raw 文件、代码文件、接口入口和测试入口不是 Wiki 页面，不使用 wikilink。

## Cross Reference 规则

当写入正向关系时，只要反向对象仍是一等 wiki 页面，就必须在同一次编辑中同步写入反向关系。

| 正向操作 | 必须同步的反向操作 |
|----------|-------------------|
| topic/T 写入 `workflows: [[workflow-W]]` | workflow/W 的 `topics` 追加 T |
| topic/T 写入 `related_topics: [[topic-R]]` | topic/R 的 `related_topics` 追加 T |
| workflow/W 写入 `topics: [[topic-T]]` | topic/T 的 `workflows` 追加 W |
| workflow/W 写入 `related_workflows: [[workflow-R]]` | workflow/R 的 `related_workflows` 追加 W |
| troubleshooting/T 写入 `topics: [[topic-A]]` | topic/A 的排障入口摘要追加 T |

以下外部对象没有反向页面要求：

- `raw/*`
- 代码文件和 symbol
- 接口入口
- 测试入口

## View 读取协议

Topic 和 Workflow 页面必须使用 `devwiki:section` 标记：

```markdown
<!-- devwiki:section id=card -->
## 导航卡
<!-- /devwiki:section -->
```

支持的 view：

- Topic：`card`、`core`、`explain`
- Workflow：`card`、`core`、`explain`

读取命令：

```bash
zatools devwiki read topic <slug> --view card --project <project>
zatools devwiki read topic <slug> --view core --project <project>
zatools devwiki read workflow <slug> --view core --project <project>
zatools devwiki search topic <query...> --project <project>
zatools devwiki search workflow <query...> --project <project>
zatools devwiki repo info
zatools devwiki repo info <project>
zatools devwiki search index <query...> --project <project>
zatools devwiki search glossary <query...> --project <project>
zatools devwiki search workflow <query...> --project <project>
zatools devwiki read workflow <slug> --view core --project <project>
zatools devwiki server --project <project> --host 0.0.0.0 --port 5697
zatools devwiki graph --project <project> --host 127.0.0.1 --port 5696
```

代码仓 `AGENTS.md` / `CLAUDE.md` 的 DevWiki link block 会写入 `DevWiki project`。如果无法从 link block 判断 project，可执行 `zatools devwiki repo info`；无参数时仅输出已配置 project 名称 JSON 数组。

`zatools devwiki repo info <project>` 默认输出 JSON，包含 DevWiki `source` 和所有关联代码仓 `code_repos[].path`。

## 检索顺序

按用户意图选择起点：

- 主题、能力、功能、规则问题：`topics`
- 代码问题：`workflows → topics → rg`
- 排障问题：先找相关 workflow/topic；troubleshooting 是排障知识目录，v1 统一 CLI 读取类型仍只使用 `topic|workflow`

检索通道按成本升档：`zatools devwiki search index/glossary`、`zatools devwiki search topic/workflow`、`zatools qmd query`。`qmd` 只是召回工具，不是真相源。

## Workflow 约束

- `raw/` 是只读原始资料
- facts 与 inference 必须分开表达
- 仅凭检索输出或其他派生内容，不能单独作为证据
- Topic 只写主题边界、功能行为、关键规则和实现入口，不写代码路径、函数名、handler、调用链
- Topic 的 `module` frontmatter 字段只用于 graph 聚合展示，不创建独立模块页面
- Workflow 只写工程实现知识，代码路径、函数、类、配置文件必须有证据
- 页面写入和证据字段更新必须遵守对应 DevWiki skill 的模板和引用规则
- 新建 Topic 或 Workflow 后必须同步检查 `wiki/glossary.md`；先查是否已有关键术语或等价别名，不存在才添加
- 项目知识任务先由 `devwiki-project-router` 判断意图、身份、证据需求和检索边界，再路由到 `devwiki-ingest`、`devwiki-topic`、`devwiki-workflow`、`devwiki-maintain`、`devwiki-code`、`devwiki-query`、`devwiki-code-to-doc` 或 qmd 维护命令
- 中高风险写入必须先给 proposal，再落盘

## 操作说明

- 使用 `zatools devwiki init` 在当前目录初始化 DevWiki 文档库并安装 skills
- 使用 `zatools devwiki repo link <project> <repo-slug> <path>` 将代码库路径写入用户级 DevWiki 项目配置
- 如需下载 qmd models，初始化完成后可在 DevWiki 工作区内手动执行 `zatools qmd download --root .`
- 对已有工作区补做或修复 qmd collection 注册、索引刷新与状态检查时，直接执行 `zatools qmd sync/update/status`
- 使用 `devwiki-project-router` 作为项目知识任务的默认总入口
- 使用 `devwiki-ingest` 吸收 raw 文档并生成 TopicTask / WorkflowTask；TopicTask 需要带 module 建议，Topic 正文交给 `devwiki-topic`，Workflow 正文交给 `devwiki-workflow`
- 使用 `devwiki-topic` 或 `devwiki-workflow` 新建页面后，必须先查 `wiki/glossary.md`，缺少关键术语时按通用格式补充
- 使用 `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失、关系错误和 query 污染
- 使用 `devwiki-code` 基于关联 DevWiki workflow 定位并修改当前代码仓
- 使用 `devwiki-query` 只读查询 Wiki、raw、文档内代码线索、设计意图和排障知识；真实代码核查交给 `devwiki-code`
- 使用 `devwiki-code-to-doc` 从代码、接口、配置项、日志或路由反向生成或更新 workflow 页面

这份运行时规则应保持稳定，只在 DevWiki 的项目目录、链接规范或工作流约束发生真实变化时修改。
