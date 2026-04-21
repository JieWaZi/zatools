---
name: "devwiki-setup"
description: "当需要用 zatools 初始化 DevWiki 工作区、选择 runtime、注册一个或多个代码目录，并安装 DevWiki skills 时使用。"
argument-hint: "[<project-name>] [--agent codex|cursor|claude --lang zh|en --code-dir <path>] [--global]"
---

# /devwiki-setup

> 先阅读通用约束：
> - `references/evidence-grounding.md`
> - `references/zatools-qmd.md`
> - `references/mutation-safety.md`

> 使用 zatools 初始化 DevWiki，统一走 `zatools devwiki init`。

## Inputs

- `project-name`：可选；缺省时在交互模式中询问，非交互模式下必须显式给出
- `--agent codex|cursor|claude`
- `--lang zh|en`
- `--code-dir <path>`：可重复传入，也可一次传逗号分隔值
- `--global`：可选；不传时默认安装到项目级
- 当前工作目录与其识别出的项目根

## Outputs

- 一个新的 `devwiki-<project-name>/` 目录
- `devwiki-<project-name>/README.md`
- `devwiki-<project-name>/raw/`
- `devwiki-<project-name>/wiki/`
- `devwiki-<project-name>/config/project.yaml`
- `devwiki-<project-name>/config/search.yaml`
- 选定的 DevWiki skills
- 项目级安装时：当前项目根下的 `.agents/` 与 `.zatools-lock.json`
- 全局安装时：用户主目录下的技能安装与锁文件
- 可选：`zatools qmd` 的注册结果或可手动执行的命令
- 可选：初始化结束后提示用户手动执行 `zatools qmd download --root .`
- 可选：若工作区已存在，可转给 `devwiki-qmd-sync`

## DevWiki Interaction

### Reads

- 当前工作目录，用于推断项目根
- 用户给定的代码目录
- 内置 DevWiki skills 与共享 references 模板
- `devwiki-<project-name>/config/search.yaml`，用于后续 `zatools qmd sync`

### Writes

- CREATE `devwiki-<project-name>/`
- CREATE `devwiki-<project-name>/raw/`
- CREATE `devwiki-<project-name>/wiki/`
- CREATE `devwiki-<project-name>/config/project.yaml`
- CREATE `devwiki-<project-name>/config/search.yaml`
- CREATE / UPDATE 已选 scope 下的 DevWiki skills
- CREATE / UPDATE `.zatools-lock.json`


## Workflow

### Step 1: 收集初始化参数

如果用户没有提供完整参数，则通过交互补齐：

- 项目名称
- runtime：`codex`、`cursor` 或 `claude`
- 语言：`zh` 或 `en`
- 一个或多个代码目录
- 安装范围：项目级或全局级

如果用户直接给了参数，则不重复追问。

### Step 2: 创建 DevWiki 项目并安装 skills

标准初始化命令：

```bash
zatools devwiki init <project-name> --agent <agent> --lang <lang> --code-dir <dir1> --code-dir <dir2>
```

非交互场景建议补上：

```bash
zatools devwiki init <project-name> --agent <agent> --lang <lang> --code-dir <dir1> --yes
```

如果用户明确要求全局安装 skills，追加：

```bash
--global
```

项目级安装时必须记住：

- `.agents/` 与 `.zatools-lock.json` 写在**当前项目根**
- 不是写进 `devwiki-<project-name>/`

### Step 3: 同步 `zatools qmd` 检索层

如果用户需要 `zatools qmd` collection，同步命令为：

```bash
zatools qmd sync --root <devwiki-root>
```

真正执行注册时再带：

```bash
zatools qmd sync --root <devwiki-root> --apply
```

注册完成后，继续执行：

```bash
zatools qmd update
zatools qmd status
```

如果用户还想手动下载 qmd models，可在 DevWiki 工作区内执行：

```bash
zatools qmd download --root .
```

如果 `status` 显示仍有大量 pending embeddings，且后续任务明确依赖更高质量语义召回，再询问用户是否继续执行：

```bash
zatools qmd embed
```

如果 `zatools qmd ...` 不可用，不要假装 setup 完成；应打印生成的命令并明确说明当前是 fallback 模式。

### Step 4: 输出 setup 报告

报告至少包含：

- 实际创建的 `devwiki-<project-name>` 路径
- 选定的 runtime 与语言
- 已注册的代码目录
- skills 安装范围是项目级还是全局级
- 项目级安装时 `.agents/` 与 `.zatools-lock.json` 的落点
- `zatools qmd` 是否已注册，还是只打印了命令
- 是否已执行 `zatools qmd update` / `zatools qmd status`
- 是否还需要后续执行 `devwiki-qmd-sync`

## Constraints

- **不再依赖旧的脚本式初始化链路**
- **`zatools devwiki init` 本身就会安装 skills**
- **不生成内部模板垃圾**：不要生成 `i18n/`、`tools/`、`setup.*`、`requirements.txt`、`config/*.example`
- **项目级安装状态写在当前项目根**：不是写进 `devwiki-<project-name>/`
- **代码目录必须真实可访问**
- **`zatools qmd` 失败不能伪装成功**：要么执行成功，要么明确回退

## Error Handling

- **项目名缺失且无法交互**：直接报错，不要猜测
- **代码目录不存在或不是目录**：停止并让用户修正 `--code-dir`
- **目标目录已存在**：直接报错，不要覆盖已有 `devwiki-<project-name>`
- **skills 安装失败**：报告失败并停止，不要宣称 setup 完成
- **`zatools qmd sync --apply` 失败**：退回到打印命令，并明确说明当前为 fallback 模式
- **用户取消交互**：明确说明 setup 未完成
