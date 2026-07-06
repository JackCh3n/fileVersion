# FileVersion —— 文件“发送到”版本后缀工具（Go）

给任意文件一键追加版本后缀 `V.YYYY_MMDD_HHMMSS`（精确到秒），例如：

- 复制：`厦门一体化政务服务平台系统运行服务月度报告-6月.docx`
  → `厦门一体化政务服务平台系统运行服务月度报告-6月V.2026_0706_113500.docx`
- 重命名（move）：同上，但原地改名；若文件名**已含**版本后缀，则**更新**该后缀部分。

> **基于 `V.` 版本标记的通用识别（非 `- 副本` 补丁）**：识别的是文件名中任意位置的 `V.YYYY_MMDD_HHMMSS` 标记本身，
> 从**首个** `V.` 处整体截断并替换为最新时间戳。因此其后面的 ` - 副本`、旧版本号等都会一并被覆盖掉，
> 例如 Windows「复制到副本」生成的 `新建文本文档V.2026_0706_163543 - 副本.txt`，再次 move 会更新为
> `新建文本文档V.2026_0706_163550.txt`。这同时也自愈了此前错误产生的「双版本」文件名（如 `…V.xxx - 副本V.yyy`）。

支持 Win7 / Win10 / Win11，支持中文路径，支持中文/英文/数字用户名（gm、zhangshan、张德顺……），
按当前用户 `%APPDATA%` / `%LOCALAPPDATA%` 自动解析，无需管理员权限。

> **通知方式**：安装 / 卸载 / 用法说明仍使用系统默认弹窗；**仅文件改名（copy / move）的成功与失败**走 Win10 / Win11 右下角 **Action Center（操作中心）** 通知（Win7 静默）。通知来源显示为“FileVersion”，无需下载任何控件（通过系统自带 PowerShell + 注册表 AppUserModelId 实现）。

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

可一次传入多个文件，会逐个处理；处理结束后在 Win10/11 右下角 Action Center 弹出一条通知（成功/失败），Win7 下静默无提示。

## 关于时间戳精度

默认精确到**秒**，后缀形如 `V.2026_0706_113500`（YYYY_MMDD_HHMMSS）。

识别正则同时兼容两种历史精度：
`V.YYYY_MMDD_HHMMSS`（6 位，当前版本）与 `V.YYYY_MMDD_HHMM`（4 位，早期版本生成的旧文件），
因此对任意历史文件执行 copy / move 都不会叠加出第二个 `V.`，而是把旧时间戳更新为最新值。
