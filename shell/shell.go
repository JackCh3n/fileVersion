// Package shell 负责注册/卸载右键上下文菜单与“发送到”快捷方式。
// 主菜单通过 HKCU\Software\Classes\*\shell 注册（无需管理员权限），
// 同时保留 SendTo 快捷方式作为备选。
package shell

import (
	"os"
	"path/filepath"
)

const (
	appName   = "FileVersion"
	linkCopy  = "FileCopy.lnk"
	linkMove  = "FileMove.lnk"
	linkClean = "FileClean.lnk"
	linkGUI   = "FileVersion 批量改名.lnk"
	exeName   = "fileversion.exe"
	iconKey   = `{6D7C7F5B-CE8A-4F8B-B9E2-1B3E9C8A7D6F}`
)

// localAppData 返回 %LOCALAPPDATA%。
func localAppData() string {
	if v := os.Getenv("LOCALAPPDATA"); v != "" {
		return v
	}
	return os.TempDir()
}

// appData 返回 %APPDATA%。
func appData() string {
	if v := os.Getenv("APPDATA"); v != "" {
		return v
	}
	return os.TempDir()
}

// installDir 安装目录。
func installDir() string {
	return filepath.Join(localAppData(), appName)
}

// exePath 安装后的 exe 完整路径。
func exePath() string {
	return filepath.Join(installDir(), exeName)
}

// SendToDir 发送到目录。
func SendToDir() string {
	return filepath.Join(appData(), "Microsoft", "Windows", "SendTo")
}
