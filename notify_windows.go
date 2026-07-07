//go:build windows

package main

import (
	"os/exec"
	"syscall"
	"unsafe"
)

// msgBox 调用 Windows API 弹系统默认对话框。
func msgBox(title, text string) {
	mb := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	mb.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(0))
}

// msgBoxErr 弹错误框（带错误图标）。
func msgBoxErr(title, text string) {
	mb := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	mb.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(0x10)) // MB_ICONERROR
}

// notify 在 Win10/11 右下角 Action Center 弹出通知；Win7 静默。
// 来源显示为 "FileVersion"（由注册表 AppUserModelId=com.fileversion.app 控制）。
func notify(title, msg string, isError bool) {
	_ = isError
	ps := `$eap = [System.Threading.ThreadPool]::QueueUserWorkItem({ ` +
		`try { ` +
		`$t = [Windows.UI.Notifications.ToastTemplateType]::ToastText02; ` +
		`$xml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($t); ` +
		`$txt = $xml.GetElementsByTagName('text'); ` +
		`$txt.Item(0).AppendChild($xml.CreateTextNode('` + title + `')); ` +
		`$txt.Item(1).AppendChild($xml.CreateTextNode('` + msg + `')); ` +
		`$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('com.fileversion.app'); ` +
		`$notifier.Show($xml); ` +
		`} catch {} ` +
		`});`
	_ = runPS(ps)
}

// runPS 执行 PowerShell 命令（静默失败，Win7 无 Action Center 时不影响主流程）。
func runPS(script string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	return cmd.Run()
}
