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
	// 已带后缀的文件，再次 newName 应更新为单一最新后缀（不叠加、不破坏前缀）
	p := `C:\tmp\报告-6月V.2026_0706_113500.docx`
	a1 := newName(p)
	base1 := filepath.Base(a1)
	if !strings.Contains(base1, "报告-6月V.") {
		t.Fatalf("前缀或后缀被破坏: %s", base1)
	}
	// 确保不会出现 报告-6月V.xxxV.yyy 的叠加
	first := strings.Index(base1, "V.")
	if strings.Contains(base1[first+1:], "V.") {
		t.Fatalf("出现后缀叠加: %s", base1)
	}
	// 再处理一次（模拟再次 move），仍不应叠加、不应破坏前缀
	a2 := newName(a1)
	base2 := filepath.Base(a2)
	first2 := strings.Index(base2, "V.")
	if strings.Contains(base2[first2+1:], "V.") {
		t.Fatalf("二次处理出现叠加: %s", base2)
	}
	if !strings.Contains(base2, "报告-6月V.") {
		t.Fatalf("二次处理后前缀或后缀被破坏: %s", base2)
	}
}

// TestNewNameUpdateWithCopySuffix 复现并验证 bug：
// Windows“复制到副本”会产生 “xxxV.2026_0706_163543 - 副本.txt”，
// 再次移动应更新版本号且去掉 “ - 副本”，得到 “xxxV.<新时间戳>.txt”。
func TestNewNameUpdateWithCopySuffix(t *testing.T) {
	p := `C:\tmp\新建文本文档V.2026_0706_163543 - 副本.txt`
	n := newName(p)
	if !strings.HasSuffix(n, ".txt") {
		t.Fatalf("扩展名丢失: %s", n)
	}
	base := filepath.Base(n)
	if !strings.Contains(base, "新建文本文档V.") {
		t.Fatalf("前缀被破坏: %s", base)
	}
	if strings.Contains(base, "副本") {
		t.Fatalf("不应残留“ - 副本”: %s", base)
	}
	// 旧时间戳应被更新掉（不能还含 163543）
	if strings.Contains(base, "163543") {
		t.Fatalf("旧时间戳应被更新: %s", base)
	}
	// 不应出现后缀叠加（如 V.xxxV.yyy）
	first := strings.Index(base, "V.")
	if strings.Contains(base[first+1:], "V.") {
		t.Fatalf("出现后缀叠加: %s", base)
	}
}

// TestNewNameCopyAlreadyVersioned 复现并验证 copy 的同类 bug：
// 源文件已带版本号时，copy 不应再叠加出第二个 V.，而应更新为单一最新版本号。
func TestNewNameCopyAlreadyVersioned(t *testing.T) {
	p := `C:\tmp\一体化平台项目重保报告-0619-端午V.2026_0622_1651.docx`
	n := newName(p)
	if !strings.HasSuffix(n, ".docx") {
		t.Fatalf("扩展名丢失: %s", n)
	}
	base := filepath.Base(n)
	if !strings.Contains(base, "一体化平台项目重保报告-0619-端午V.") {
		t.Fatalf("前缀或版本标记被破坏: %s", base)
	}
	// 必须只有一个 V.，不能出现 V.xxxV.yyy
	first := strings.Index(base, "V.")
	if strings.Contains(base[first+1:], "V.") {
		t.Fatalf("copy 出现后缀叠加: %s", base)
	}
	// 旧时间戳 2026_0622_1651 应被更新掉
	if strings.Contains(base, "2026_0622_1651") {
		t.Fatalf("旧时间戳应被更新: %s", base)
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
	// 此时目录里有 2 个带后缀文件（复制产生的 + 移动产生的），均不应有后缀叠加
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !isVersioned(e.Name()) {
			continue
		}
		b := e.Name()
		first := strings.Index(b, "V.")
		if strings.Contains(b[first+1:], "V.") {
			t.Fatalf("出现后缀叠加: %s", b)
		}
	}

	// 再对其中一个带后缀文件 move 一次：应更新为单一最新版本（不叠加、不增减文件数）
	entries2, _ := os.ReadDir(dir)
	var target string
	for _, e := range entries2 {
		if isVersioned(e.Name()) {
			target = e.Name()
			break
		}
	}
	if target == "" {
		t.Fatal("未找到带后缀文件")
	}
	if err := doMove(filepath.Join(dir, target)); err != nil {
		t.Fatalf("二次 doMove: %v", err)
	}
	final, _ := os.ReadDir(dir)
	cnt := 0
	for _, e := range final {
		if isVersioned(e.Name()) {
			cnt++
			b := e.Name()
			first := strings.Index(b, "V.")
			if strings.Contains(b[first+1:], "V.") {
				t.Fatalf("二次 move 后出现后缀叠加: %s", b)
			}
		}
	}
	// 二次 move 不应改变带后缀文件的数量（仅更新时间戳，碰撞时加 _2 也保持数量守恒）
	if cnt != 2 {
		t.Fatalf("二次 move 后带后缀文件应为 2，实际 %d: %v", cnt, final)
	}
}

func TestCreateShortcut(t *testing.T) {
	sendTo := t.TempDir()
	target := `C:\Program Files\FileVersion\fileversion.exe`
	if err := createShortcut(sendTo, "FileCopy.lnk", target, "copy"); err != nil {
		t.Fatalf("createShortcut: %v", err)
	}
	lnk := filepath.Join(sendTo, "FileCopy.lnk")
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

