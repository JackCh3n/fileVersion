//go:build windows

// fileversion —— 文件“发送到”小工具
//
// 功能：
//   copy  复制文件，并在文件名后追加版本后缀 V.YYYY_MMDD_HHMMSS
//   move  重命名文件，追加版本后缀（若已存在则更新该后缀）
//   install  安装到当前用户，并在“发送到”菜单创建 FileCopy / FileMove 两个快捷方式
//   uninstall  卸载，移除快捷方式与安装目录
//
// 版本后缀示例：V.2026_0706_113500（精确到秒）
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// 已存在的版本后缀（用于 move 时更新）。例如 V.2026_0706_113500（精确到秒）
var reVersion = regexp.MustCompile(`(V\.\d{4}_\d{4}_\d{6})$`)

// versionSuffix 返回当前时间的版本后缀，如 V.2026_0706_113500
func versionSuffix(t time.Time) string {
	return "V." + t.Format("2006_0102_150405")
}

// newName 根据原路径计算带版本后缀的新路径。
// 若原文件名已含版本后缀，则先移除旧后缀再加新的（实现“更新”语义）。
func newName(path string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)
	name = reVersion.ReplaceAllString(name, "")
	return filepath.Join(filepath.Dir(path), name+versionSuffix(time.Now())+ext)
}

// copyFile 逐块复制文件内容（支持大文件）。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// avoidCollision 若目标已存在，则在末尾追加 _1/_2... 避免覆盖。
func avoidCollision(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); err != nil {
			return candidate
		}
	}
	return path
}

func doCopy(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("不支持目录：%s", path)
	}
	dest := newName(path)
	if dest == path {
		return fmt.Errorf("目标文件名与源文件相同：%s", path)
	}
	if _, err := os.Stat(dest); err == nil {
		dest = avoidCollision(dest)
	}
	if err := copyFile(path, dest); err != nil {
		return fmt.Errorf("复制失败 %s: %w", path, err)
	}
	return nil
}

func doMove(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("不支持目录：%s", path)
	}
	dest := newName(path)
	if dest == path {
		// 文件名已经是最新后缀，无需修改
		return nil
	}
	if _, err := os.Stat(dest); err == nil {
		dest = avoidCollision(dest)
	}
	if err := os.Rename(path, dest); err != nil {
		return fmt.Errorf("重命名失败 %s: %w", path, err)
	}
	return nil
}

// ---- Windows 通知（Win10/11 Action Center） ----

// “发送到”菜单中的快捷方式名称
const (
	linkCopy = "FileCopy.lnk"
	linkMove = "FileMove.lnk"
)

// notify 在 Win10/11 上通过右下角 Action Center 弹出通知；Win7 没有 Action Center，
// 直接静默（不弹窗、也不回退到其它提示）。非 Windows 环境同样静默。
//
// 实现：调用系统自带的 PowerShell 加载 WinRT 的 ToastNotificationManager 显示通知，
// 不依赖任何第三方库。Win7 上 WinRT 类型不可用，脚本会抛错，此处忽略即可（即静默）。
func notify(title, message string, isError bool) {
	if runtime.GOOS != "windows" {
		return
	}

	// 文案清洗：换行替换为空格，过长截断（Action Center 文本容量有限）
	msg := strings.TrimSpace(message)
	msg = strings.ReplaceAll(msg, "\r\n", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	if len([]rune(msg)) > 240 {
		msg = string([]rune(msg)[:240]) + "…"
	}

	const script = `param($Title, $Message)
try {
  [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
  $tpl = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
  $txt = $tpl.GetElementsByTagName('text')
  $txt.Item(0).AppendChild($tpl.CreateTextNode($Title)) | Out-Null
  $txt.Item(1).AppendChild($tpl.CreateTextNode($Message)) | Out-Null
  $appId = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
  [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($appId).Show($tpl)
  Start-Sleep -Milliseconds 300
} catch { exit 1 }`

	tmp, err := os.CreateTemp("", "fv-toast-*.ps1")
	if err != nil {
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		return
	}
	tmp.Close()

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-NoProfile", "-STA",
		"-File", tmpName, "-Title", title, "-Message", msg)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Run() // 忽略错误：Win7 或组策略禁用通知时静默
}

// ---- 安装（创建“发送到”快捷方式） ----

func doInstall() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	installDir := filepath.Join(localAppData, "FileVersion")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return err
	}
	destExe := filepath.Join(installDir, "fileversion.exe")
	if err := copyFile(exe, destExe); err != nil {
		return err
	}

	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	sendTo := filepath.Join(appData, "Microsoft", "Windows", "SendTo")
	if err := os.MkdirAll(sendTo, 0o755); err != nil {
		return err
	}

	if err := createShortcut(sendTo, linkCopy, destExe, "copy"); err != nil {
		return err
	}
	if err := createShortcut(sendTo, linkMove, destExe, "move"); err != nil {
		return err
	}

	notify("FileVersion 安装完成",
		"已安装到 "+destExe+"\n右键文件 → 发送到 中已出现：FileCopy（复制加版本号）、FileMove（重命名加版本号）",
		false)
	return nil
}

// createShortcut 通过 PowerShell 创建 .lnk 快捷方式。
// 使用临时 .ps1 文件（带 BOM）以正确处理中文路径。
func createShortcut(sendTo, name, target, args string) error {
	lnk := filepath.Join(sendTo, name)
	script := fmt.Sprintf(
		"$ws=New-Object -ComObject WScript.Shell\n"+
			"$s=$ws.CreateShortcut(\"%s\")\n"+
			"$s.TargetPath=\"%s\"\n"+
			"$s.Arguments=\"%s\"\n"+
			"$s.Description=\"FileVersion %s\"\n"+
			"$s.Save()\n",
		lnk, target, args, args,
	)

	tmp, err := os.CreateTemp("", "fv-*.ps1")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString("\ufeff" + script); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", tmpName)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// doUninstall 卸载：删除本用户 SendTo 中的快捷方式与安装目录。
func doUninstall() error {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
	}
	sendTo := filepath.Join(appData, "Microsoft", "Windows", "SendTo")

	// 同时清理新旧两种快捷方式名称，避免旧版本残留
	links := []string{
		filepath.Join(sendTo, linkCopy),
		filepath.Join(sendTo, linkMove),
		filepath.Join(sendTo, "复制并加版本号(FileVersion).lnk"),
		filepath.Join(sendTo, "重命名加版本号(FileVersion).lnk"),
	}
	var errs []string
	for _, l := range links {
		if err := os.Remove(l); err != nil && !os.IsNotExist(err) {
			errs = append(errs, l+": "+err.Error())
		}
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	installDir := filepath.Join(localAppData, "FileVersion")
	if err := os.RemoveAll(installDir); err != nil {
		errs = append(errs, installDir+": "+err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("部分清理失败：\n%s", strings.Join(errs, "\n"))
	}
	notify("FileVersion 卸载完成",
		"已移除 SendTo 中的 FileCopy / FileMove 快捷方式与安装目录 "+installDir, false)
	return nil
}

// ---- 入口 ----

func main() {
	defer func() {
		if r := recover(); r != nil {
			notify("FileVersion 错误", fmt.Sprintf("发生异常：%v", r), true)
		}
	}()

	if len(os.Args) < 2 {
		notify("FileVersion 使用说明",
			"用法：fileversion.exe install 安装（创建“发送到”菜单）；uninstall 卸载；copy <文件> 复制并加版本后缀；move <文件> 重命名加版本后缀（已存在则更新）。安装后右键文件 → 发送到 即可使用（FileCopy / FileMove），支持多文件。",
			false)
		return
	}

	mode := strings.ToLower(os.Args[1])
	files := os.Args[2:]

	switch mode {
	case "install":
		if err := doInstall(); err != nil {
			notify("FileVersion 安装失败", err.Error(), true)
		}
		return
	case "uninstall":
		if err := doUninstall(); err != nil {
			notify("FileVersion 卸载失败", err.Error(), true)
		}
		return
	case "copy", "move":
		if len(files) == 0 {
			notify("FileVersion", "未提供文件。请右键文件 → 发送到 使用。", false)
			return
		}
		ok, fail := 0, 0
		var errs []string
		for _, f := range files {
			var err error
			if mode == "copy" {
				err = doCopy(f)
			} else {
				err = doMove(f)
			}
			if err != nil {
				fail++
				errs = append(errs, f+"\n  "+err.Error())
			} else {
				ok++
			}
		}
		// Win10/11 走右下角 Action Center 通知；Win7 静默（无 Action Center，不回退）
		if fail > 0 {
			first := ""
			if len(errs) > 0 {
				first = strings.SplitN(errs[0], "\n", 2)[0]
			}
			notify("FileVersion 完成（有失败）",
				fmt.Sprintf("成功 %d，失败 %d。首个失败：%s", ok, fail, first), true)
		} else if ok > 0 {
			notify("FileVersion 完成", fmt.Sprintf("已成功处理 %d 个文件。", ok), false)
		}
	default:
		notify("FileVersion", "未知命令："+mode+"\n请使用 install / uninstall / copy / move。", false)
	}
}
