package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func isVersioned(name string) bool {
	return reVersion.MatchString(strings.TrimSuffix(name, filepath.Ext(name)))
}

func TestNewNameBasic(t *testing.T) {
	p := `C:\tmp\厦门一体化政务服务平台系统运行服务月度报告-6月.docx`
	n := newName(p)
	if !strings.HasSuffix(n, ".docx") {
		t.Fatalf("扩展名丢失: %s", n)
	}
	if !isVersioned(filepath.Base(n)) {
		t.Fatalf("缺少版本后缀: %s", n)
	}
	if !strings.Contains(n, "厦门一体化政务服务平台系统运行服务月度报告-6月V.") {
		t.Fatalf("原文件名被破坏: %s", n)
	}
}

func TestNewNameUpdateExisting(t *testing.T) {
	// 已带后缀的文件，再次 newName 应保持后缀（幂等），不被叠加
	p := `C:\tmp\报告-6月V.2026_0706_113500.docx`
	a1 := newName(p)
	a2 := newName(a1)
	if a1 != a2 {
		t.Fatalf("已带后缀应幂等: %s -> %s", a1, a2)
	}
	if !isVersioned(filepath.Base(a1)) {
		t.Fatalf("应仍含版本后缀: %s", a1)
	}
	// 确保不是 报告-6月V.xxxV.yyy 的叠加
	idx := strings.Index(filepath.Base(a1), "V.")
	if strings.Contains(filepath.Base(a1)[idx+1:], "V.") {
		t.Fatalf("出现后缀叠加: %s", a1)
	}
}

func TestCopyAndMove(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "月度报告-6月.docx")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	// copy：原文件保留，新增带后缀副本
	if err := doCopy(src); err != nil {
		t.Fatalf("doCopy: %v", err)
	}
	afterCopy, _ := os.ReadDir(dir)
	var versioned, origin int
	for _, e := range afterCopy {
		if e.Name() == "月度报告-6月.docx" {
			origin++
		}
		if isVersioned(e.Name()) {
			versioned++
		}
	}
	if origin != 1 || versioned != 1 {
		t.Fatalf("copy 后应有 1 原 + 1 副本，实际: %v", afterCopy)
	}

	// move：把原文件改名加后缀
	if err := doMove(src); err != nil {
		t.Fatalf("doMove: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("move 后原文件应不存在: %s", src)
	}
	// 目录里应有一个带后缀的文件
	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if isVersioned(e.Name()) {
			found = true
		}
	}
	if !found {
		t.Fatalf("move 后未找到带后缀文件: %v", entries)
	}

	// 再 move 一次（此时已带后缀）：应保持单一带后缀文件，不叠加
	for _, e := range entries {
		if isVersioned(e.Name()) {
			if err := doMove(filepath.Join(dir, e.Name())); err != nil {
				t.Fatalf("二次 doMove: %v", err)
			}
		}
	}
	final, _ := os.ReadDir(dir)
	cnt := 0
	for _, e := range final {
		if isVersioned(e.Name()) {
			cnt++
		}
	}
	if cnt != 1 {
		t.Fatalf("二次 move 后带后缀文件应为 1，实际 %d: %v", cnt, final)
	}
}

func TestCreateShortcut(t *testing.T) {
	sendTo := t.TempDir()
	target := `C:\Program Files\FileVersion\fileversion.exe`
	if err := createShortcut(sendTo, "复制并加版本号(FileVersion).lnk", target, "copy"); err != nil {
		t.Fatalf("createShortcut: %v", err)
	}
	lnk := filepath.Join(sendTo, "复制并加版本号(FileVersion).lnk")
	info, err := os.Stat(lnk)
	if err != nil {
		t.Fatalf("快捷方式未生成: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("快捷方式为空")
	}
	// 用 PowerShell 读回，确认 TargetPath 与参数正确（中文名称也应无乱码）
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`$s=(New-Object -ComObject WScript.Shell).CreateShortcut('%s'); Write-Output ($s.TargetPath+';'+$s.Arguments)`, lnk)).Output()
	if err != nil {
		t.Fatalf("读回快捷方式失败: %v", err)
	}
	got := strings.TrimSpace(string(out))
	want := target + ";copy"
	if got != want {
		t.Fatalf("快捷方式内容不匹配:\n got  %q\n want %q", got, want)
	}
}

