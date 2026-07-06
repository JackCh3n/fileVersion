# FileVersion —— 文件“发送到”版本后缀工具（Go）

给任意文件一键追加版本后缀 `V.YYYY_MMDD_HHMMSS`（精确到秒），例如：

- 复制：`厦门一体化政务服务平台系统运行服务月度报告-6月.docx`
  → `厦门一体化政务服务平台系统运行服务月度报告-6月V.2026_0706_113500.docx`
- 重命名（move）：同上，但原地改名；若文件名**已含**版本后缀，则**更新**该后缀部分。

支持 Win7 / Win10 / Win11，支持中文路径，支持中文/英文/数字用户名（gm、zhangshan、张德顺……），
按当前用户 `%APPDATA%` / `%LOCALAPPDATA%` 自动解析，无需管理员权限。

> 文件操作（copy / move）**成功后不弹窗**（静默完成），仅在**失败**时弹出错误提示。

---

## 安装（每个用户只需一次）

以**要使用的那个用户账号**登录系统，运行：

```bat
fileversion.exe install
```

它会：

1. 把 `fileversion.exe` 复制到 `%LOCALAPPDATA%\FileVersion\`（当前用户专属，无需管理员）。
2. 在 `%APPDATA%\Microsoft\Windows\SendTo` 下创建两个快捷方式：
   - `FileCopy.lnk` → 执行 `copy`（复制并加版本号）
   - `FileMove.lnk` → 执行 `move`（重命名加版本号）

> 不同用户各自运行一次 `install` 即可，互不干扰。

## 使用

1. 在资源管理器里**右键文件**（可多选）→ **发送到**
2. 选择：
   - **FileCopy**：生成一份带版本号的新副本，原文件不动。
   - **FileMove**：直接把原文件改名为带版本号的名字；若原本已有 `V.2026_0706_113500`，则替换为新的时间戳。

## 卸载

```bat
fileversion.exe uninstall
```

会自动移除：
- `%LOCALAPPDATA%\FileVersion\` 安装目录
- `%APPDATA%\Microsoft\Windows\SendTo\` 下的 `FileCopy.lnk` 与 `FileMove.lnk`
  （同时兼容清理旧版中文名称的快捷方式）

## 手动编译

```bat
build.bat
```

生成：
- `fileversion.exe` —— 64 位（Win7/10/11 主流）
- `fileversion-386.exe` —— 32 位（兼容老 32 位 Win7）

## 命令行用法

```bat
fileversion.exe install                 # 安装到本用户 + 创建“发送到”菜单
fileversion.exe uninstall               # 卸载，移快捷方式与安装目录
fileversion.exe copy  "报告.docx"       # 复制并加版本后缀
fileversion.exe move  "报告.docx"       # 重命名并加版本后缀（已存在则更新）
```

可一次传入多个文件，会逐个处理；仅当有文件失败时才弹窗汇总。

## 关于时间戳精度

默认精确到**秒**，后缀形如 `V.2026_0706_113500`（YYYY_MMDD_HHMMSS）。
如需只精确到**分**，把 `main.go` 中 `versionSuffix` 的
`time.Format("2006_0102_150405")` 改回 `time.Format("2006_0102_1504")`，
并把 `reVersion` 正则末尾的 `\d{6}` 改回 `\d{4}`，重新 `build.bat` 即可。
