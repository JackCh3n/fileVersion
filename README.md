# FileVersion —— 文件批量改名工具（Go + Wails GUI）

右键文件即可批量重命名：支持**整体模板命名、查找替换、前后缀/位置增删、clean 整理、版本号追加**，
并带有**实时预览、冲突处理、一键撤销**。

> 由原先的“发送到”命令行小工具升级为带 GUI 的批量改名器（Go + Wails），核心重命名引擎仍为纯标准库实现。

---

## 功能一览

| 模式 | 说明 |
|------|------|
| **整体（模板）** | 用模板 `A_#` / `*V.#` 批量命名；支持编号起始/增量/位数、随机编号、扩展名改写、自动冲突解决 |
| **替换** | 查找/替换文件名内容，支持正则、忽略大小写 |
| **添加/删除** | 文件名前后缀、指定位置插入、按子串或位置删除字符 |
| **clean 整理** | 去 `(N)` 计数，把日期规整为 `V.YYYY_MM_DD` 并移到末尾（含完整时间戳只取日期） |
| **版本号** | 追加/更新 `V.时间戳` 后缀（copy=复制并加版本号，move=原地改名） |

通用能力：
- **实时预览**：在「预览」列即时显示每个文件改名后的结果，不实际执行。
- **冲突处理**：可配置 `跳过 / 覆盖 / 自动编号`，结果列用颜色标识。
- **撤销（Revert）**：记录每次批量操作的 `原名→新名` 到日志，一键还原上一批。
- **多文件**：支持拖拽 / 添加按钮，可排序、移除、全部清空。

---

## 安装（每个用户只需一次）

以**要使用的那个用户账号**登录系统，运行：

```bat
fileversion.exe install
```

它会：
1. 把 `fileversion.exe` 复制到 `%LOCALAPPDATA%\FileVersion\`；
2. 在 `%APPDATA%\Microsoft\Windows\SendTo` 下创建快捷方式（FileCopy / FileMove / FileClean / GUI）；
3. 在**右键主菜单**注册（无需管理员，写 `HKCU\Software\Classes\*\shell` 与 `Directory\shell`），带图标；
4. 注册 Action Center 通知来源显示为「FileVersion」。

卸载：

```bat
fileversion.exe uninstall
```

> 右键文件 / 文件夹即可看到「FileVersion 批量改名」「复制加版本号」「重命名加版本号」「整理文件名」等菜单项。

---

## 配置（JSON）

配置文件位于 `%APPDATA%\FileVersion\config.json`，记录用户偏好：

```json
{
  "conflictStrategy": "auto",   // skip / overwrite / auto
  "templateStart": 1,
  "templateIncrement": 1,
  "templateDigits": 3,
  "copyOutputDir": "",         // copy 模式输出目录，空=同目录
  "cleanKeepTime": false,      // clean 是否保留完整时间戳
  "windowWidth": 900,
  "windowHeight": 640
}
```

---

## 命令行用法（兼容旧脚本 / 自动化）

```bat
fileversion.exe install                 # 安装到本用户 + 创建右键菜单
fileversion.exe uninstall               # 卸载
fileversion.exe copy  "报告.docx"       # 复制并加版本后缀
fileversion.exe move  "报告.docx"       # 重命名并加版本后缀（已存在则更新）
fileversion.exe clean "报告 - 20260604(1).docx"  # 整理：去 (N)、日期规整为 V.YYYY_MM_DD
```

> 不带任何参数运行 `fileversion.exe` 则启动 GUI。

---

## 模板占位符

| 占位符 | 含义 |
|--------|------|
| `*` | 原文件名（插入位置） |
| `#` | 序号（按 起始/增量/位数 生成，可随机） |

示例：
- `A_#` → `A_001`, `A_002` …
- `*V.#` → `原文件V.01`, `原文件V.02` …
- `IMG_#_原图` → `IMG_001_原图` …

---

## 开发 / 构建

环境：Go 1.22+，Windows 10/11（GUI 依赖 WebView2，Win10/11 自带）。

```bat
# 安装 Wails CLI（一次性）
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 开发模式（热重载前端）
build.bat dev
# 或
wails dev

# 构建发布版
build.bat
# 或
wails build -platform windows/amd64
wails build -platform windows/386 -o fileversion-386.exe   # 兼容 32 位 Win7
```

项目结构：

```
fileversion/
├── main.go              # Wails 入口（无参数=GUI；有子命令=CLI）
├── app.go               # 前端绑定（AddFiles/ComputePreview/ExecuteRename/Revert/Install…）
├── cli.go               # 命令行子命令处理
├── assets.go            # 内嵌前端
├── notify_windows.go    # Win10/11 Action Center 通知（Win7 静默）
├── renamer/             # 核心重命名引擎（纯标准库，可单测）
│   ├── types.go         # 类型与计划定义
│   ├── version.go       # 版本号后缀
│   ├── clean.go         # clean 整理
│   ├── template.go      # 模板命名
│   ├── addremove.go     # 前后缀/增删
│   ├── plan.go          # 计划编排 + 冲突处理 + 执行
│   └── revert.go        # 撤销日志
├── config/             # JSON 配置
├── shell/               # 右键菜单 / SendTo 注册（Windows）
├── frontend/            # HTML/CSS/JS 界面
└── wails.json
```

核心引擎 `renamer` 包不依赖 Wails，可在任意平台 `go test ./renamer/` 验证。

---

## 关于时间戳精度

默认精确到**秒**，后缀形如 `V.2026_0707_114232`（YYYY_MMDD_HHMMSS）。
识别正则同时兼容两种历史精度：`V.YYYY_MMDD_HHMMSS`（6 位）与 `V.YYYY_MMDD_HHMM`（4 位），
因此对历史文件执行 copy / move 不会叠加第二个 `V.`，而是更新为最新值。
