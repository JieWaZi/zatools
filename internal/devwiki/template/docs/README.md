# DevWiki

面向研发场景的结构化 Wiki 与代码检索工作流。

DevWiki 的目标不是临时 RAG，而是把需求、设计、功能说明、测试资料和代码定位线索持续沉淀成可维护、可检索、可修正的知识层。新需求到来时，Agent 先利用已有 Wiki、raw 资料和必要代码线索做范围收敛，再判断这是新增能力、已有能力改造，还是需要继续向用户确认。

## 什么是 DevWiki

DevWiki 是一个单产品知识底座。文档库可以独立存在，并通过 `devwiki link` 关联一个或多个代码库。核心思路是：

- 把原始资料放在 `raw/`
- 把结构化知识沉淀到 `wiki/`
- 用 `zatools qmd ...` 加速 `wiki` 结构化知识召回；raw 和代码目录只在用户手动加入 collection 时进入 qmd
- DevWiki 的人工知识主模型围绕 capabilities、features、workflows 和 troubleshooting
- 用 `capabilities` 说明系统能力、业务边界、能力效果和覆盖功能
- 用 `features` 说明具体功能、参数、取值、联动、设计思想和功能流转
- 用 `workflows` 合并工程流程和模块定位，说明入口、调用链、关键逻辑、代码引用和修改影响
- 用 `troubleshooting` 沉淀故障现象、诊断路径、证据和修复建议
- 用 `devwiki-project-router` 统一判断用户意图，再分流到摄入、查询、Wiki 健康维护、代码反向成文或 qmd 检索层维护流程

同时保留辅助根文件和报告目录：

- `outputs`
- `relations.yml`
- `glossary.md`

## 为什么不是临时 RAG

| 维度 | 临时 RAG | DevWiki |
|------|----------|---------|
| 知识持久化 | 每次查询重新拼接 | 文档持续沉淀为结构化 Wiki |
| 业务理解 | 容易只返回片段 | `capabilities` 维护能力地图和边界 |
| 功能讨论 | 功能说明容易碎片化 | `features` 统一承接功能设计、参数和联动 |
| 代码定位 | 容易盲搜代码 | `workflows` 面向编程维护入口、调用链和关键文件 |
| 意图路由 | Agent 容易直接乱答 | `devwiki-project-router` 先判断任务类型、证据需求和下一步 Skill |
| 知识修正 | 上次答错，下次还可能错 | 用 `devwiki-maintain` 维护现有 Wiki 健康，用 `devwiki-ingest` 补来源，或用 `devwiki-code-to-doc` 从代码补证据 |

## 快速开始

### 第一步：执行 `zatools devwiki init`

不携带参数时会进入交互式流程：

```bash
zatools devwiki init
```

如果你已经明确项目名称、agent、语言和代码目录，也可以一次性传完：

```bash
zatools devwiki init "{{PROJECT_NAME}}" --agent {{AGENT}} --lang {{LANG}} --code-dir "{{PRIMARY_CODE_DIR}}" --yes
```

`zatools devwiki init` 会完成以下动作：

- 直接把当前工作目录作为 DevWiki 文档库根目录，不再额外创建单独的 DevWiki 子目录
- 在当前目录生成 `README.md`
- 在当前目录生成当前 agent 对应的运行时文件 `{{RUNTIME_FILE}}`
- 生成 `config/project.yaml` 与 `config/search.yaml`
- 在当前目录安装完整 DevWiki skills
- 在初始化结束后提示用户：如需，可手动执行 `zatools qmd download --root .`

如果已经有一个 DevWiki 文档库，只想补做代码库关联，可以执行：

```bash
zatools devwiki link --root . --code-dir "{{PRIMARY_CODE_DIR}}"
```

`zatools devwiki link` 会把代码库与本 DevWiki 文档库关联起来。代码库中的关联说明会告诉 Agent：在代码库内使用 `devwiki-query` 或 `devwiki-code-to-doc` 时，必须先读取本 DevWiki 文档库的 `{{RUNTIME_FILE}}`，查询从本目录的 `wiki/`、`raw/`、`config/search.yaml` 取证，生成的新 Wiki 文件也写回本目录。

### 第二步：同步 `zatools qmd` 检索层

先预览将要执行的注册命令：

```bash
zatools qmd sync --root .
```

确认无误后再执行：

```bash
zatools qmd sync --root . --apply
```

如果你想手动提前把模型下载好，可以在 DevWiki 工作区内执行：

```bash
zatools qmd download --root .
```

完成 collection 注册后，建议立刻刷新索引并查看状态：

```bash
zatools qmd update
zatools qmd status
```

如果你接下来依赖 `zatools qmd query` 做更高质量的语义召回，再按需执行：

```bash
zatools qmd embed
```

生成后的 `config/search.yaml` 默认只注册 Wiki collection，避免把 raw 或代码库里的大量无关文件自动纳入 qmd。需要额外注册 raw 或代码目录时，手动编辑 `config/search.yaml` 后再执行 sync。

```yaml
qmd:
  collections:
    - name: devwiki-{{PROJECT_SLUG}}-wiki
      path: wiki
```

### 第三步：准备原始资料

把原始文档放到 `raw/` 对应目录：

- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/tests/`

### 第四步：开始构建 Wiki

完成初始化后，就可以进入 Agent 运行时执行：

- `devwiki-qmd-sync`
- `devwiki-project-router`
- `devwiki-ingest`
- `devwiki-maintain`
- `devwiki-query`
- `devwiki-code-to-doc`

## 典型使用闭环

### 场景一：从原始资料启动第一版 Wiki

1. 把需求、设计、功能说明、测试资料放入 `raw/`
2. 执行 `devwiki-project-router`，由它路由到 `devwiki-ingest`
3. Agent 先输出 proposal，说明准备创建或更新哪些 capability、feature、workflow、troubleshooting
4. 对拆分边界、命名冲突和中高风险写入做确认后落盘

### 场景二：查询功能、设计或代码位置

1. 用 `devwiki-project-router` 提出问题
2. Router 路由到 `devwiki-query`
3. Agent 在 `capabilities / features / workflows / troubleshooting` 中按意图召回证据
4. 只有问题涉及当前实现、代码位置、修改影响或排障时，才继续核对代码

### 场景三：从代码反向生成 Wiki 文档

1. 用 `devwiki-project-router` 提供接口、文件、函数、路由、配置项或日志锚点
2. Router 路由到 `devwiki-code-to-doc`
3. Agent 从真实代码追踪入口和关键边界，默认生成或更新 `workflows`
4. 只有需要补功能设计时，才同步更新 `features`

### 场景四：维护已有 Wiki 健康

1. 用 `devwiki-project-router` 描述错误回答、旧机制、冲突页面或待维护范围
2. Router 路由到 `devwiki-maintain`
3. Agent 审计 Wiki、source、relations、index、glossary，必要时核对代码
4. 对中高风险知识修正先输出 Maintain Proposal，确认后再落盘并刷新 qmd 索引

## 核心能力

| 技能 | 作用 |
|------|------|
| `devwiki-qmd-sync` | 为已有工作区补做或修复 `zatools qmd` collection 注册、索引刷新与状态检查 |
| `devwiki-project-router` | 总入口，先判断意图、身份、证据需求、qmd / 代码检索边界，再路由到具体 Skill |
| `devwiki-ingest` | 增量消化需求、设计、测试或会议资料，更新三层 Wiki 页面和关系 |
| `devwiki-maintain` | 维护已有 Wiki 的证据一致性、过期内容、引用缺失、关系错误和 query 污染 |
| `devwiki-query` | 查询已有 Wiki、raw 和必要的代码线索，回答功能、设计、流程、代码定位和排障问题 |
| `devwiki-code-to-doc` | 从代码、接口、配置项、日志或路由反向生成或更新 workflow / 代码定位页面 |

## 目录结构

```text
./
├── README.md                    ← 项目说明、使用方式与工作流入口
├── {{RUNTIME_FILE}}             ← 当前 agent 对应的运行时规则副本
├── config/
│   ├── project.yaml             ← 当前项目默认配置
│   └── search.yaml              ← `zatools qmd` collection 配置
├── raw/
│   ├── requirements/
│   ├── designs/
│   ├── features/
│   └── tests/
└── wiki/
    ├── index.md
    ├── glossary.md
    ├── relations.yml
    ├── log.md
    ├── capabilities/
    ├── features/
    ├── workflows/
    ├── troubleshooting/
    └── outputs/
```

其中：

- `raw/` 是只读原始资料层
- `wiki/capabilities/` 是业务能力地图
- `wiki/features/` 是功能设计页
- `wiki/workflows/` 是工程定位页，合并流程和模块定位
- `wiki/troubleshooting/` 是排障知识页
- `wiki/outputs/` 保存 ingest / maintain / query / code-to-doc / qmd-sync 报告
- `wiki/relations.yml` 保存能力、功能、工程定位和排障关系
- `wiki/glossary.md` 保存术语表
- 当前目录还会持有项目级 DevWiki skills、`.cache/` 和 `.zatools-lock.json`

## 设计原则

- `raw/` 只读，不直接改写源材料
- `capabilities` 只写业务能力、作用效果和边界，不写代码引用
- `features` 只写功能设计、参数、取值、联动和功能流转，不写代码引用
- `workflows` 合并原 workflow/module，负责调用链、关键文件、关键函数、状态落点、修改影响和测试入口
- 待确认问题优先通过对话和用户对齐，不默认写入文件
- `relations.yml` 负责跨 capability / feature / workflow / troubleshooting 的结构化关系
- DevWiki 任务先由 `devwiki-project-router` 统一判断意图和证据边界
- 已有 Wiki 出现旧结论、缺证据、断链、关系错误或 query 污染时，使用 `devwiki-maintain` 做证据一致性维护
- 中高风险动作必须先提案、后确认
- `zatools qmd ...` 只是检索加速层，不是真相源
- 代码关联必须能落到文件、函数、接口入口、测试入口等可追踪对象

## 常用维护命令

预览 `zatools qmd` 注册命令：

```bash
zatools qmd sync --root .
```

更新当前作用域下已变化的 DevWiki runtime skills：

```bash
zatools devwiki update
```

执行 `zatools qmd` 注册：

```bash
zatools qmd sync --root . --apply
```

刷新 `zatools qmd` 索引：

```bash
zatools qmd update
```

查看 `zatools qmd status`：

```bash
zatools qmd status
```

按需刷新向量：

```bash
zatools qmd embed
```

预览 reset 计划：

```bash
zatools devwiki tool reset --scope wiki --project-root .
```

真正执行 reset：

```bash
zatools devwiki tool reset --scope wiki --project-root . --yes
```

向 `wiki/log.md` 追加日志：

```bash
zatools devwiki tool log --wiki-root wiki --message "init | note"
```

## 当前边界

当前版本优先保证单产品文档库与一个或多个代码库的关联场景可演示、可闭环。重点覆盖：

- 原始文档进入 `raw/`
- `devwiki-project-router` 作为总入口
- `devwiki-ingest` 从原始资料生成和更新结构化 Wiki
- `devwiki-maintain` 维护已有 Wiki 的证据一致性、过期内容、引用缺失和 query 污染
- `devwiki-query` 覆盖知识查询、设计理解、代码定位和排障检索
- `devwiki-code-to-doc` 从代码反向补 workflow，必要时补 feature
- `devwiki-qmd-sync` 维护 qmd collection 注册、索引和状态
