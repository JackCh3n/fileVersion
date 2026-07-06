//go:build windows

// fileversion —— 文件“发送到”小工具
//
// 功能：
//   copy  复制文件，并在文件名后追加版本后缀 V.YYYY_MMDD_HHMM
//   move  重命名文件，追加版本后缀（若已存在则更新该后缀）
//   install  安装到当前用户，并在“发送到”菜单创建两个快捷方式
//
// 版本后缀示例：V.2026_0706_1135
// 时间戳精确到“分”（与示例一致）。如需精确到“秒”，把 versionSuffix 中的
// time.Format("2006_0102_1504") 改为 time.Format("2006_0102_150405") 即可。
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"
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

// ---- Windows 消息框 ----

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32.NewProc("MessageBoxW")
)

const (
	mbOKOnly    = 0x00000000
	mbIconInfo  = 0x00000040
	mbIconError = 0x00000010
)

func msgBox(title, text string) {
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(mbOKOnly|mbIconInfo),
	)
}

func msgBoxErr(title, text string) {
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(mbOKOnly|mbIconError),
	)
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

	if err := createShortcut(sendTo, "复制并加版本号(FileVersion).lnk", destExe, "copy"); err != nil {
		return err
	}
	if err := createShortcut(sendTo, "重命名加版本号(FileVersion).lnk", destExe, "move"); err != nil {
		return err
	}

	msgBox("FileVersion 安装完成",
		"已安装到：\n"+destExe+
			"\n\n右键文件 → 发送到 中已出现：\n"+
			"• 复制并加版本号(FileVersion)\n"+
			"• 重命名加版本号(FileVersion)")
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

// ---- 入口 ----

func main() {
	defer func() {
		if r := recover(); r != nil {
			msgBoxErr("FileVersion 错误", fmt.Sprintf("发生异常：%v", r))
		}
	}()

	if len(os.Args) < 2 {
		msgBox("FileVersion 使用说明",
			"用法：\n"+
				"  fileversion.exe install        安装到本用户并创建“发送到”菜单\n"+
				"  fileversion.exe copy  <文件>   复制文件并加版本后缀\n"+
				"  fileversion.exe move  <文件>   重命名文件加版本后缀（已存在则更新）\n\n"+
				"安装后，右键文件 → 发送到 即可使用，支持一次选中多个文件。")
		return
	}

	mode := strings.ToLower(os.Args[1])
	files := os.Args[2:]

	switch mode {
	case "install":
		if err := doInstall(); err != nil {
			msgBoxErr("FileVersion 安装失败", err.Error())
		}
		return
	case "copy", "move":
		if len(files) == 0 {
			msgBox("FileVersion", "未提供文件。请右键文件 → 发送到 使用。")
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
		if fail == 0 {
			msgBox("FileVersion 完成", fmt.Sprintf("成功处理 %d 个文件。", ok))
		} else {
			msgBoxErr("FileVersion 完成（有失败）",
				fmt.Sprintf("成功 %d，失败 %d：\n\n%s", ok, fail, strings.Join(errs, "\n")))
		}
	default:
		msgBox("FileVersion", "未知命令："+mode+"\n请使用 install / copy / move。")
	}
}
