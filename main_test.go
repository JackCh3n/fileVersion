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

// TestCleanName 覆盖 clean 模式的典型场景（含用户给出的真实微信文件名）。
func TestCleanName(t *testing.T) {
	cases := []struct {
		in   string // 输入（Windows 风格路径，仅用于断言基名）
		want string // 期望的完整目标路径
	}{
		// —— 用户此前用例：带日期 + (N) → 日期规整为 V.YYYY_MM_DD，去掉 (N) ——
		{`C:\a\周例会相关工作汇报 - 20260604(1).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\周例会相关工作汇报 - 20260604(2).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		// 无日期，仅 (N) → 只去掉 (N)
		{`C:\a\信息安全自查(1).doc`, `C:\a\信息安全自查.doc`},
		{`C:\a\信息安全自查(2).doc`, `C:\a\信息安全自查.doc`},
		// 带分隔符的日期同样能识别
		{`C:\a\周报 - 2026-06-04(3).docx`, `C:\a\周报V.2026_06_04.docx`},
		{`C:\a\周报 - 2026.06.04(1).docx`, `C:\a\周报V.2026_06_04.docx`},
		// 无 (N) 但含日期 → 同样规整
		{`C:\a\周例会相关工作汇报 - 20260604.docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		// 已经是规整形态（末尾 V.）→ 保持不变（不再叠加）
		{`C:\a\周例会相关工作汇报V.2026_06_04.docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		// 已是规整形态但多了 (N) → 去 (N) 保留 V.
		{`C:\a\周例会相关工作汇报V.2026_06_04(1).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},

		// —— 微信真实文件名：日期在开头，规整后移到末尾 ——
		// 纯日期（无名称）
		{`C:\a\20260618.txt`, `C:\a\V.2026_06_18.txt`},
		// 日期 + 中文名称
		{`C:\a\20250917关于"一网协同"一站式政务工作平台申请开通政务云资源（第二十三批）的函.pdf`,
			`C:\a\关于"一网协同"一站式政务工作平台申请开通政务云资源（第二十三批）的函V.2025_09_17.pdf`},
		// 仅到“月”的日期 YYYYMM
		{`C:\a\202606业务受理信息推送待核实(1).xls`, `C:\a\业务受理信息推送待核实V.2026_06.xls`},
		// 完整时间戳 YYYYMMDDHHMMSS：只取日期部分、丢弃时分、移到末尾
		{`C:\a\20260525000100.xlsx`, `C:\a\V.2026_05_25.xlsx`},
		{`C:\a\20260616172059_license.dat`, `C:\a\licenseV.2026_06_16.dat`},
		{`C:\a\20260611(1).zip`, `C:\a\V.2026_06_11.zip`},
		// 日期 + 多余后缀（脚本包名）连同时间一并移除
		{`C:\a\202504181407035_rocketmq-all-5.3.0-带一键部署脚本安装包附操作视频.zip`,
			`C:\a\rocketmq-all-5.3.0-带一键部署脚本安装包附操作视频V.2025_04_18.zip`},
		// 日期后带空格+时间（空格作为分隔符）
		{`C:\a\厦门一体化政务服务2026-06-02 16.06.pdf`, `C:\a\厦门一体化政务服务V.2026_06_02.pdf`},

		// —— 不应被误改的文件（无合法日期 / 身份证号 / 业务单号 / 年份在括号里）——
		{`C:\a\49xxc.com.zip`, `C:\a\49xxc.com.zip`},
		{`C:\a\6742_厦门一体化政务服务平台.zip`, `C:\a\6742_厦门一体化政务服务平台.zip`},
		{`C:\a\6he-admin-new.rar`, `C:\a\6he-admin-new.rar`},
		{`C:\a\book.js`, `C:\a\book.js`},
		{`C:\a\library_redis_8.8.0_arm64v8.tar`, `C:\a\library_redis_8.8.0_arm64v8.tar`},
		// 身份证号片段里的数字串不得误判为日期
		{`C:\a\xm0080026060101817中华人民共和国居民身份证.ofd`, `C:\a\xm0080026060101817中华人民共和国居民身份证.ofd`},
		// 业务单号里的 200013…/20260684… 属于非法日期（月/日越界），忽略
		{`C:\a\11350200MB1816343GQT2000132026068407W.txt`, `C:\a\11350200MB1816343GQT2000132026068407W.txt`},
		// 年份在〔〕里，不是可提取的日期
		{`C:\a\闽厦路政许〔2026〕127号(1).sspx`, `C:\a\闽厦路政许〔2026〕127号.sspx`},
		// “2026年”中文年份不提取
		{`C:\a\厦门市教育局2026年第二次高中、中职教师资格认定预约人数安排表.xlsx`,
			`C:\a\厦门市教育局2026年第二次高中、中职教师资格认定预约人数安排表.xlsx`},
		// 已是 V. 时间戳形态（move 生成），保持（不降级为日期）
		{`C:\a\一体化平台项目重保报告-0619-端午V.2026_0622_1651.docx`,
			`C:\a\一体化平台项目重保报告-0619-端午V.2026_0622_1651.docx`},
	}
	for _, c := range cases {
		got := cleanName(c.in)
		if got != c.want {
			t.Errorf("cleanName(%q)\n  got = %q\n  want= %q", c.in, got, c.want)
		}
		// 绝不能出现双 V.
		base := filepath.Base(got)
		first := strings.Index(base, "V.")
		if first >= 0 && strings.Contains(base[first+1:], "V.") {
			t.Errorf("出现双 V. 叠加: %q", base)
		}
	}
}

// TestCleanDo 在真实文件系统上验证 doClean 的改名行为。
func TestCleanDo(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "周例会相关工作汇报 - 20260604(1).docx")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := doClean(src); err != nil {
		t.Fatalf("doClean: %v", err)
	}
	want := filepath.Join(dir, "周例会相关工作汇报V.2026_06_04.docx")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("clean 后未得到期望文件 %s: %v", want, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("clean 后原文件应已被改名: %s", src)
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

