//go:build windows

package shell

import (
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// Install 安装：复制 exe、创建 SendTo 快捷方式、注册主右键菜单。
func Install() error {
	src, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(installDir(), 0o755); err != nil {
		return err
	}
	if err := copyFile(src, exePath()); err != nil {
		return err
	}
	// SendTo 快捷方式
	if err := createShortcut(exePath(), "copy", filepath.Join(SendToDir(), linkCopy)); err != nil {
		return err
	}
	if err := createShortcut(exePath(), "move", filepath.Join(SendToDir(), linkMove)); err != nil {
		return err
	}
	if err := createShortcut(exePath(), "clean", filepath.Join(SendToDir(), linkClean)); err != nil {
		return err
	}
	if err := createShortcut(exePath(), "", filepath.Join(SendToDir(), linkGUI)); err != nil {
		return err
	}
	// 主右键菜单（文件）
	if err := registerContextMenu("*", "FileVersionGUI", "FileVersion 批量改名", exePath(), `%1`); err != nil {
		return err
	}
	if err := registerContextMenu("*", "FileVersionCopy", "FileVersion - 复制加版本号", exePath(), `copy "%1"`); err != nil {
		return err
	}
	if err := registerContextMenu("*", "FileVersionMove", "FileVersion - 重命名加版本号", exePath(), `move "%1"`); err != nil {
		return err
	}
	if err := registerContextMenu("*", "FileVersionClean", "FileVersion - 整理文件名", exePath(), `clean "%1"`); err != nil {
		return err
	}
	// 主右键菜单（文件夹）
	if err := registerContextMenu("Directory", "FileVersionGUI", "FileVersion 批量改名", exePath(), `%1`); err != nil {
		return err
	}
	// 注册 Action Center 通知来源显示为 "FileVersion"
	if err := registerToastSource(exePath()); err != nil {
		return err
	}
	return nil
}

// registerToastSource 在 HKCU\Software\Classes\AppUserModelId\com.fileversion.app
// 写入 DisplayName=FileVersion，使右下角通知来源显示为“FileVersion”而非“Windows PowerShell”。
func registerToastSource(exePath string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Software\Classes\AppUserModelId\com.fileversion.app`, registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()
	k.SetStringValue("DisplayName", "FileVersion")
	k.SetStringValue("IconUri", "file:///"+exePath+",0")
	return nil
}

// Uninstall 卸载：删除安装目录、SendTo 快捷方式、注册表菜单项。
func Uninstall() error {
	os.Remove(exePath())
	os.Remove(installDir())
	os.Remove(filepath.Join(SendToDir(), linkCopy))
	os.Remove(filepath.Join(SendToDir(), linkMove))
	os.Remove(filepath.Join(SendToDir(), linkClean))
	os.Remove(filepath.Join(SendToDir(), linkGUI))
	unregisterContextMenu("*", "FileVersionGUI")
	unregisterContextMenu("*", "FileVersionCopy")
	unregisterContextMenu("*", "FileVersionMove")
	unregisterContextMenu("*", "FileVersionClean")
	unregisterContextMenu("Directory", "FileVersionGUI")
	return nil
}

// registerContextMenu 在 HKCU\Software\Classes\<key>\shell 下注册一个带图标的右键项。
// key 通常是 "*"（文件）或 "Directory"（文件夹）。
func registerContextMenu(key, sub, title, exePath, cmd string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Software\Classes\`+key+`\shell\`+sub, registry.WRITE)
	if err != nil {
		return err
	}
	defer k.Close()
	k.SetStringValue("", title)
	k.SetStringValue("Icon", exePath+",0")
	ck, _, err := registry.CreateKey(k, "command", registry.WRITE)
	if err != nil {
		return err
	}
	defer ck.Close()
	ck.SetStringValue("", `"`+exePath+`" `+cmd)
	return nil
}

// unregisterContextMenu 删除注册表菜单项。
func unregisterContextMenu(key, sub string) error {
	return registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\`+key+`\shell\`+sub)
}

// createShortcut 创建 Windows .lnk 快捷方式（目标 exe + 参数）。
func createShortcut(target, args, lnkPath string) error {
	// 使用 PowerShell 创建 lnk，避免引入 COM 依赖。
	ps := `$ws = New-Object -ComObject WScript.Shell; ` +
		`$s = $ws.CreateShortcut('` + lnkPath + `'); ` +
		`$s.TargetPath = '` + target + `'; ` +
		`$s.Arguments = '` + args + `'; ` +
		`$s.Save()`
	return runPowerShell(ps)
}

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
	buf := make([]byte, 32*1024)
	for {
		n, e := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if e != nil {
			break
		}
	}
	return nil
}

func runPowerShell(script string) error {
	// 通过 cmd 调用 powershell -Command
	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	return cmd.Run()
}
