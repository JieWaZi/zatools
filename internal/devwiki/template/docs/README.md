# DevWiki

面向研发场景的结构化 Wiki 与代码检索工作流。

DevWiki 的目标不是“让 Agent 临时回答一次问题”，而是把需求文档、设计文档、特性说明、代码总结、复盘等材料持续沉淀成可维护、可检索、可修正的知识层。新需求到来时，Agent 先利用已有 Wiki 和代码线索做范围收敛，再决定这是新增能力、已有能力改造，还是仍需继续追问。

## 什么是 DevWiki

DevWiki 是一个单产品、单仓库场景下的研发知识底座，核心思路是：

- 把原始资料放在 `raw/`
- 把结构化知识沉淀到 `wiki/`
- 用 `zatools qmd ...` 加速 `raw / wiki / code` 三层召回
- 用 `capabilities` 和 `changes` 回答“这是什么能力”和“这是新增还是改造”
- 用 `devwiki-refresh`、`devwiki-check`、`devwiki-edit` 处理知识漂移和人工纠错

它参考了 OmegaWiki 的结构化工作流，但实体模型换成了研发场景的一等对象：

- `documents`
- `capabilities`
- `changes`

## 为什么不是临时 RAG

| 维度 | 临时 RAG | DevWiki |
|------|----------|---------|
| 知识持久化 | 每次查询重新拼接 | 文档持续沉淀为结构化 Wiki |
| 研发实体 | 通常只有文本块 | 有 `documents / capabilities / changes` 三类中心对象 |
| 代码关联 | 只靠临时关键词命中 | 用 `code_refs`、接口 URL、函数名做结构化关联 |
| 变更定性 | Agent 容易乱猜 | 用 `devwiki-scope` 判断 `new / modify / unclear` |
| 知识修正 | 上次答错，下次还可能错 | 可用 `refresh / check / edit` 持续修正 |
| 无文档场景 | 只能瞎搜代码 | 可用 `devwiki-feature-doc` 从代码反向梳理特性文档 |

## 快速开始

### 环境要求

- 已安装 `zatools`
- 如果你要使用 Claude 侧运行时：需要先安装 Claude Code
- 如果你要使用 `zatools qmd ...` 检索加速：需要确保当前环境可以成功执行 `zatools qmd status`
- 如果你要在 Codex、Claude Code 这类沙箱 agent 中执行 `zatools qmd ...`：还要确保 agent 具备对应执行权限，以及项目根 `.cache` 目录写权限

### 第一步：执行 `zatools devwiki init`

不携带参数时会进入交互式流程：

```bash
zatools devwiki init
```

如果你已经明确项目名称、agent、语言和代码目录，也可以一次性传完：

```bash
zatools devwiki init "{{PROJECT_NAME}}" --agent {{AGENT}} --lang {{LANG}} --code-dir "{{PRIMARY_CODE_DIR}}" --yes
```

如果你希望把 DevWiki skills 安装到全局级，而不是当前项目根，可以追加：

```bash
--global
```

`zatools devwiki init` 会完成以下动作：

- 检测当前项目根，并在项目根下创建 `{{WORKSPACE_DIR}}/`
- 生成 `README.md`
- 生成当前 agent 对应的运行时文件 `{{RUNTIME_FILE}}`
- 生成 `config/project.yaml` 与 `config/search.yaml`
- 安装 DevWiki skills
- 预热一次 `zatools qmd` 所需模型，并尽量提前完成首次下载
- 在当前项目根创建或更新桥接用运行时文件，引导 agent 先读取 `./{{WORKSPACE_DIR}}/{{RUNTIME_FILE}}`

如果是项目级安装，项目级 skill 安装状态、桥接用运行时文件和 `.zatools-lock.json` 都写在当前检测到的项目根，不写进 `{{WORKSPACE_DIR}}/`。

### 第二步：同步 `zatools qmd` 检索层

初始化会在创建完成后自动做一次 `zatools qmd` 预热，尽量提前完成 collection 注册、索引刷新和模型下载；如果你是在已有工作区里补接 qmd，仍然优先使用 `zatools qmd sync` 或 `devwiki-qmd-sync`。

通过 `zatools qmd ...` 执行检索与维护命令时，会把以下 flags 映射成环境变量：

- `--embed-model` -> `QMD_EMBED_MODEL`
- `--rerank-model` -> `QMD_RERANK_MODEL`
- `--generate-model` -> `QMD_GENERATE_MODEL`

如果不显式传参，则默认值直接使用内置模型：

- `hf:Qwen/Qwen3-Embedding-0.6B-GGUF/Qwen3-Embedding-0.6B-Q8_0.gguf`
- `hf:ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF/qwen3-reranker-0.6b-q8_0.gguf`
- `hf:tobil/qmd-query-expansion-1.7B-gguf/qmd-query-expansion-1.7B-q4_k_m.gguf`

如果 agent 当前就在某个 DevWiki 工作区里，推荐先读取 `config/search.yaml`，再把其中的三个模型显式带到 `zatools qmd ...` 命令里，这样执行目录不在 DevWiki 根时也不会丢配置。

`zatools qmd ...` 会自动注入 `XDG_CACHE_HOME`，并把它指到检测到的项目根目录下的 `.cache`。

如果你在 Codex 或 Claude Code 里运行，`zatools qmd ...` 常见失败原因不是命令本身，而是沙箱拦住了执行或阻止写入 cache 目录。遇到这种情况时，需要先给对应 agent 放开 `zatools qmd ...` 的执行权限，并确保项目根下 `.cache/` 可写。

先预览将要执行的注册命令：

```bash
zatools qmd sync --root .
```

确认无误后再执行：

```bash
zatools qmd sync --root . --apply
```

如果你希望单独提前把模型下载好，也可以直接执行：

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

如果工作区已经存在，只是想补做或修复 `zatools qmd ...` 注册与索引，可以直接执行 `devwiki-qmd-sync`。

生成后的 `config/search.yaml` 会为同一个项目注册一组 namespaced collections，例如：

```yaml
qmd:
  collections:
    - name: devwiki-{{PROJECT_SLUG}}-raw
      path: raw
    - name: devwiki-{{PROJECT_SLUG}}-wiki
      path: wiki
    - name: devwiki-{{PROJECT_SLUG}}-code-{{PRIMARY_CODE_SLUG}}
      path: {{PRIMARY_CODE_DIR}}
  embed_model: hf:Qwen/Qwen3-Embedding-0.6B-GGUF/Qwen3-Embedding-0.6B-Q8_0.gguf
  rerank_model: hf:ggml-org/Qwen3-Reranker-0.6B-Q8_0-GGUF/qwen3-reranker-0.6b-q8_0.gguf
  generate_model: hf:tobil/qmd-query-expansion-1.7B-gguf/qmd-query-expansion-1.7B-q4_k_m.gguf
```

如果同一个项目下挂了多个代码目录，会继续追加更多 `devwiki-{{PROJECT_SLUG}}-code-*` collection；但它们仍然都在同一个项目命名空间下，不会和别的项目冲突。

### 第三步：准备原始资料

把原始文档放到 `raw/` 对应目录：

- `raw/requirements/`
- `raw/designs/`
- `raw/features/`
- `raw/code-summaries/`
- `raw/postmortems/`
- `raw/api/`
- `raw/tests/`

当前以 Markdown 为主。后续再扩展 `docx / pdf -> md` 转换工具。

### 第四步：开始构建 Wiki

完成初始化后，就可以进入 Agent 运行时执行：

- `devwiki-setup`
- `devwiki-qmd-sync`
- `devwiki-init`
- `devwiki-ingest`
- `devwiki-scope`
- `devwiki-ask`
- `devwiki-feature-doc`
- `devwiki-refresh`
- `devwiki-check`
- `devwiki-edit`
- `devwiki-reset`

## 典型使用闭环

### 场景一：从原始资料启动第一版 Wiki

1. 把需求、设计、特性说明、代码总结、复盘等文件放入 `raw/`
2. 执行 `devwiki-init`
3. Agent 生成第一版 `documents / capabilities / changes` 骨架
4. 对中高风险写入做确认后落盘

### 场景二：增量吸收一批新文档

1. 向 `raw/` 补充新文档
2. 执行 `devwiki-ingest`
3. Agent 识别文档类型、匹配已有 capability / change、补代码线索
4. 提案确认后更新 Wiki

### 场景三：开发前先判断需求属于新增还是改造

1. 执行 `devwiki-scope`
2. Agent 在 `wiki / raw / code` 三层召回相关证据
3. 输出 `new / modify / unclear`
4. 给出相关文档、历史 change、候选代码文件和函数

### 场景四：缺少现成文档，只能从代码反向梳理

1. 执行 `devwiki-feature-doc`
2. 提供至少一个明确特性名称
3. 最好再提供一个入口锚点：接口 URL、关键文件、关键函数、页面路径 / 路由
4. Agent 按标准步骤逐层阅读调用链，输出 `raw/features/<特性名称>特性文档.md`

### 场景五：知识漂移或用户纠错

1. 执行 `devwiki-refresh`
2. 检查 `source_hash`、失效路径、失效 symbol、归类漂移
3. 生成修正提案
4. 用户确认后回写 Wiki

### 场景六：工作区已存在，但 `zatools qmd ...` 检索层还没真正接上

1. 执行 `devwiki-qmd-sync`
2. Agent 校验 `config/search.yaml` 与当前 `zatools qmd` collection 状态
3. 按需执行 `zatools qmd sync --root . --apply`
4. 如果还没做过模型预热，执行 `zatools qmd download --root .`
5. 继续执行 `zatools qmd update`，必要时补 `zatools qmd embed`
6. 用 `zatools qmd status` 与实际检索结果确认 `zatools qmd` 检索层已生效

## 核心能力

当前已落地的能力主要通过以下技能提供：

| 技能 | 作用 |
|------|------|
| `devwiki-setup` | 解释初始化约束、安装范围与运行时使用方式 |
| `devwiki-qmd-sync` | 为已有工作区补做或修复 `zatools qmd` collection 注册、索引刷新与状态检查 |
| `devwiki-init` | 从现有 `raw/` 启动第一版知识骨架 |
| `devwiki-ingest` | 增量吸收新的原始文档 |
| `devwiki-scope` | 在开发前归并上下文并判断 `new / modify / unclear` |
| `devwiki-ask` | 对已有 Wiki 与代码线索做通用问答 |
| `devwiki-feature-doc` | 在缺少文档时从代码反向梳理特性文档 |
| `devwiki-refresh` | 修复知识漂移与错误关联 |
| `devwiki-check` | 做确定性健康检查 |
| `devwiki-edit` | 对 Wiki 或 raw 入口做定点编辑 |
| `devwiki-reset` | 按 scope 重置工作区生成内容 |

## 目录结构

```text
{{WORKSPACE_DIR}}/
├── README.md                    ← 项目说明、使用方式与工作流入口
├── {{RUNTIME_FILE}}             ← 当前 agent 对应的运行时规则副本
├── config/
│   ├── project.yaml             ← 当前项目默认配置
│   └── search.yaml              ← `zatools qmd` collection 配置
├── raw/
│   ├── requirements/
│   ├── designs/
│   ├── features/
│   ├── code-summaries/
│   ├── postmortems/
│   ├── api/
│   └── tests/
└── wiki/
    ├── documents/
    │   ├── requirements/
    │   ├── designs/
    │   ├── features/
    │   ├── code-summaries/
    │   ├── postmortems/
    │   ├── api/
    │   └── tests/
    ├── capabilities/
    ├── changes/
    ├── index.md
    └── log.md
```

当前生成骨架只保留用户会直接使用的目录和配置；安装脚本、内部模板目录以及额外的派生层/工具层目录，不再出现在 fresh init 产物中。

其中：

- `raw/` 是只读原始资料层
- `wiki/` 是结构化知识层
- `config/search.yaml` 只保存检索配置，不会自动执行 `zatools qmd` 注册
- 当前项目根还会额外持有项目级 skills、桥接用运行时文件和 `.zatools-lock.json`

## 设计原则

- `raw/` 只读，不直接改写源材料
- 中高风险动作必须先提案、后确认
- `zatools qmd ...` 只是检索加速层，不是真相源
- 代码关联必须能落到文件、函数、接口 URL 等可追踪对象
- 检索几轮仍低置信时，Agent 必须停止乱搜并向用户提问

## 常用维护命令

`zatools qmd ...` 会自动注入 qmd 所需模型环境变量，并把 `XDG_CACHE_HOME` 指向检测到的项目根目录下 `.cache/`，这里不需要手动 `export`。

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

当前版本优先保证单产品、单仓库场景可演示、可闭环。重点覆盖：

- 原始文档进入 `raw/`
- `init / ingest` 生成结构化 Wiki
- `devwiki-scope` 辅助开发前收敛上下文
- `devwiki-feature-doc` 从代码补文档
- `refresh / check / edit / reset` 保持知识可维护

多仓库、多产品、复杂格式转换与更强自动化，会在后续版本继续扩展。
