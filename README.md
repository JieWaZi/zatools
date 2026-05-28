# zatools

`zatools` 是一个 Go 编写的 CLI，用于安装和管理可复用的 `skill`、`rule` 资产，并提供 DevWiki 工程初始化、qmd 检索封装、多 Agent 安装目录同步和锁文件维护能力。

当前主要能力：

- 管理 `skill`
- 管理 `rule`
- 初始化和维护 DevWiki 工程
- 封装 qmd 检索命令
- 将资产安装到不同 Agent 的约定目录

当前支持的 Agent：

- `codex`
- `cursor`
- `claude`

---

## 支持范围

| 类型 | 状态 | 入口 | 说明 |
| --- | --- | --- | --- |
| `skill` | 已支持 | `zatools skill ...` | 支持发现、安装、删除、检查更新、更新、初始化模板 |
| `rule` | 已支持 | `zatools rule ...` | 支持发现、安装、删除、检查更新、更新 |
| `devwiki` | 已支持 | `zatools devwiki ...` | 支持初始化、项目配置、查询、图谱、只读服务、更新和内置维护工具 |
| `qmd` | 已支持 | `zatools qmd ...` | 封装 qmd，统一注入模型和缓存配置 |
| `agent` | 已支持 | `--agent` | 作为安装目标使用，不是独立资产类型 |
| `command` | 未实现 | 无 | 暂无安装逻辑 |
| `hook` | 未实现 | 无 | 暂无安装逻辑 |

限制说明：

- `skill` 支持项目级和全局级安装
- `rule` 目前只支持项目级安装
- `rule` 目前只支持 `cursor` 和 `claude`
- `codex` 暂不支持 `rule`

---

## 构建与验证

```bash
go build ./cmd/zatools
go run ./cmd/zatools --help
```

```bash
go test ./...
go test -race ./...
go vet ./...
```

---

## 安装

Linux / macOS 可通过 GitHub Release 中的安装脚本安装最新版本：

```bash
curl -fsSL https://github.com/JieWaZi/zatools/releases/latest/download/install.sh | bash
```

安装脚本会自动识别系统和架构，下载对应的 release 压缩包，校验 `checksums.txt` 后安装 `zatools`。默认安装到 `/usr/local/bin`；如果无写入权限且无法使用 `sudo`，会回退到 `~/.local/bin`。

指定版本安装：

```bash
VERSION=v0.1.0 curl -fsSL https://github.com/JieWaZi/zatools/releases/latest/download/install.sh | bash
```

指定安装目录：

```bash
ZATOOLS_INSTALL_DIR="$HOME/.local/bin" curl -fsSL https://github.com/JieWaZi/zatools/releases/latest/download/install.sh | bash
```

Windows PowerShell：

```powershell
iwr https://github.com/JieWaZi/zatools/releases/latest/download/install.ps1 -UseBasicParsing | iex
```

发布新版本时，创建并推送 `vX.Y.Z` tag 即可触发 GitHub Actions 自动构建和更新 Release：

```bash
git tag v0.1.0
git push origin v0.1.0
```

---

## 命令总览

```text
zatools
├── skill
│   ├── add
│   ├── list
│   ├── init
│   ├── remove
│   ├── check
│   └── update
├── rule
│   ├── add
│   ├── list
│   ├── remove
│   ├── check
│   └── update
├── devwiki
│   ├── init
│   ├── update
│   ├── repo
│   ├── read
│   ├── search
│   ├── check
│   ├── graph
│   ├── server
│   └── tool
│       ├── reset
│       └── log
├── qmd
└── completion
```

---

## skill 常用命令

```bash
zatools skill add <source> [--agent codex,cursor,claude] [--global] [--list] [--skill <name>] [--yes]
zatools skill list [--global]
zatools skill init [name]
zatools skill remove [skills...] [--skill <name>] [--all] [--global] [--yes]
zatools skill check [--global]
zatools skill update [skills...] [--global]
```

示例：

```bash
# 查看可发现的 skill
zatools skill add ./examples/skills --list

# 安装指定 skill
zatools skill add ./examples/skills --skill golang-pro --agent codex --agent cursor

# 从 GitHub 安装
zatools skill add owner/repo

# 从远端子目录安装
zatools skill add owner/repo/skills/backend

# 指定分支或标签
zatools skill add owner/repo#main

# 全局安装
zatools skill add ./examples/skills --global --yes

# 安装或更新 DevWiki 内置 skills
zatools skill add devwiki
zatools skill update devwiki
```

---

## rule 常用命令

```bash
zatools rule add <source> [--agent cursor,claude] [--list] [--rule <name>] [--yes]
zatools rule list
zatools rule remove [rules...] [--rule <name>] [--all] [--yes]
zatools rule check
zatools rule update
```

示例：

```bash
# 查看可发现的 rule
zatools rule add ./examples/rules --list

# 安装指定 rule
zatools rule add ./examples/rules --rule common/engineering --agent cursor

# 安装全部 rule
zatools rule add ./examples/rules --yes
```

---

## DevWiki 命令

```bash
zatools devwiki init [project-name] [--agent <codex|cursor|claude>] [--code-dir <dir>]... [--global] [--yes]
zatools devwiki update
zatools devwiki repo add <project> [path] [--remote <url>]
zatools devwiki repo link <project> <repo-slug> <path>
zatools devwiki repo info [project]
zatools devwiki read <topic|workflow> <slug> [--view <card|core|explain>] [--format text] [--root <dir>] [--project <project>]
zatools devwiki search <index|glossary|topic|workflow> <query...> [--root <dir>] [--project <project>]
zatools devwiki check [document|graph] [path...] [--root <dir>]
zatools devwiki graph [--root <dir>] [--project <project>] [--host 0.0.0.0] [--port 5696] [--no-open] [--force]
zatools devwiki server [--root <dir>] [--project <project>] [--host 0.0.0.0] [--port 5697]
zatools devwiki tool reset --scope <wiki|raw|log|checkpoints|all> [--project-root <dir>] [--yes]
zatools devwiki tool log --wiki-root <dir> --message "<text>"
```

说明：

- `devwiki init`：初始化 DevWiki 文档库，并安装运行时所需 skills
- `devwiki update`：更新当前作用域内的 DevWiki 内置 skills，并尽力执行 qmd 注册、索引和向量刷新；qmd 失败只提示告警
- `devwiki repo`：维护用户级 DevWiki 项目配置，支持本地文档库或远端 HTTP API；输出默认 JSON，`repo info` 无参数时只列出 project 名称，有 project 时同时包含已绑定代码仓路径
- `devwiki read`：按 topic / workflow 的 `card`、`core`、`explain` 视图读取结构化页面内容
- `devwiki search`：`index` / `glossary` 本地解析结构化表格并输出最小 JSON；`topic` / `workflow` 调用 `qmd search` 后过滤并输出 `file`、`slug`、`title` 和 `score` JSON
- `devwiki check`：校验 index/glossary/log 格式、Topic/Workflow 文档分块和图谱关系；未指定类型时检查 document 和 graph，未指定路径时检查 `wiki/`
- `devwiki graph`：从 topic / workflow 页面生成图谱数据并启动本地图谱页面，默认监听 `0.0.0.0:5696`，可通过 `--host` / `--port` 指定；自动打开浏览器失败时只提示，不影响服务运行
- `devwiki server`：启动只读 HTTP API，默认 `0.0.0.0:5697`，接口使用内置 Basic Auth
- `tool reset`：默认只输出 dry-run 计划，加 `--yes` 后才会执行
- `tool log`：向 `wiki/log.md` 追加操作记录

---

## qmd 命令

```bash
zatools qmd [--embed-model <model>] [--rerank-model <model>] [--generate-model <model>] <native-qmd-args...>
zatools qmd sync [--root <dir>] [--apply]
zatools qmd download [--root <dir>]
```

说明：

- Agent 应统一使用 `zatools qmd ...`
- 除 `sync` 和 `download` 外，其余参数会透传给底层 qmd
- `sync` 默认只打印建议命令，加 `--apply` 后才实际执行
- `download` 用于提前下载 qmd 所需模型
- qmd 相关模型参数会映射为环境变量：
    - `QMD_EMBED_MODEL`
    - `QMD_RERANK_MODEL`
    - `QMD_GENERATE_MODEL`
- `XDG_CACHE_HOME` 会自动指向项目根目录下的 `.cache`

示例：

```bash
zatools qmd sync --root .
zatools qmd sync --root . --apply
zatools qmd download --root .
zatools qmd update
zatools qmd status
zatools qmd query "payment retry policy"
```

---

## 来源格式

`skill add` 和 `rule add` 共用来源解析规则。

| 格式 | 示例 |
| --- | --- |
| 本地目录 | `./skills` |
| GitHub shorthand | `owner/repo` |
| GitHub 子目录 | `owner/repo/skills/demo` |
| GitHub URL | `https://github.com/owner/repo` |
| GitHub tree URL | `https://github.com/owner/repo/tree/main/skills/demo` |
| GitLab URL | `https://gitlab.com/group/repo` |
| GitLab tree URL | `https://gitlab.com/group/repo/-/tree/main/rules/demo` |
| 直接 git URL | `https://example.com/repo.git` |
| 指定 ref | `owner/repo#main` |
| 内置 DevWiki skills | `devwiki` 或 `zatools/devwiki` |

约束：

- 子路径不能是绝对路径
- 子路径不能包含 `..`
- `#` 后内容会作为分支、标签或提交哈希
- `github:owner/repo` 会转成 GitHub shorthand
- `gitlab:group/repo` 会转成 GitLab URL

---

## 安装路径

### skill

| Agent | 项目级 | 全局级 |
| --- | --- | --- |
| `codex` | `.agents/skills` | `~/.codex/skills` |
| `cursor` | `.cursor/skills` | `~/.cursor/skills` |
| `claude` | `.claude/skills` | `~/.claude/skills` |

### rule

| Agent | 项目级 | 全局级 |
| --- | --- | --- |
| `cursor` | `.cursor/rules` | 不支持 |
| `claude` | `.claude/rules` | 不支持 |
| `codex` | 不支持 | 不支持 |

### 锁文件

| 场景 | 路径 |
| --- | --- |
| 项目级 | `<project-root>/.zatools-lock.json` |
| 全局级 skill | `~/.agents/.zatools-lock.json` |

---

## 项目根目录判定

项目级操作会从当前目录向上查找项目根目录。命中以下任一标记时，即认为该目录是项目根：

```text
.git
.jj
.hg
go.mod
package.json
pyproject.toml
Cargo.toml
Gemfile
.agents
.cursor
.claude
```

如果没有命中任何标记，则使用当前工作目录。

---

## 安装、更新与删除

安装流程：

1. 解析来源
2. 本地目录校验，远端仓库 clone 到临时目录
3. 发现可安装资产
4. 选择目标资产
5. 解析目标 Agent 和作用域
6. 安装并同步到目标目录
7. 写入锁文件

更新检查：

- 不依赖 git commit
- 重新解析来源并计算内容哈希
- 哈希变化则标记为 `outdated`

删除保护：

- 只允许删除当前 Agent 安装根目录下的路径
- 路径越界会拒绝删除
- 安装根目录本身不会被删除

---
