package renamer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`C:\a\周例会相关工作汇报 - 20260604(1).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\周例会相关工作汇报 - 20260604(2).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\信息安全自查(1).doc`, `C:\a\信息安全自查.doc`},
		{`C:\a\周报 - 2026-06-04(3).docx`, `C:\a\周报V.2026_06_04.docx`},
		{`C:\a\周报 - 2026.06.04(1).docx`, `C:\a\周报V.2026_06_04.docx`},
		{`C:\a\周例会相关工作汇报 - 20260604.docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\周例会相关工作汇报V.2026_06_04.docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\周例会相关工作汇报V.2026_06_04(1).docx`, `C:\a\周例会相关工作汇报V.2026_06_04.docx`},
		{`C:\a\20260618.txt`, `C:\a\V.2026_06_18.txt`},
		{`C:\a\20250917关于"一网协同"一站式政务工作平台申请开通政务云资源（第二十三批）的函.pdf`,
			`C:\a\关于"一网协同"一站式政务工作平台申请开通政务云资源（第二十三批）的函V.2025_09_17.pdf`},
		{`C:\a\202606业务受理信息推送待核实(1).xls`, `C:\a\业务受理信息推送待核实V.2026_06.xls`},
		{`C:\a\20260525000100.xlsx`, `C:\a\V.2026_05_25.xlsx`},
		{`C:\a\20260616172059_license.dat`, `C:\a\licenseV.2026_06_16.dat`},
		{`C:\a\20260611(1).zip`, `C:\a\V.2026_06_11.zip`},
		{`C:\a\202504181407035_rocketmq-all-5.3.0-带一键部署脚本安装包附操作视频.zip`,
			`C:\a\rocketmq-all-5.3.0-带一键部署脚本安装包附操作视频V.2025_04_18.zip`},
		{`C:\a\厦门一体化政务服务2026-06-02 16.06.pdf`, `C:\a\厦门一体化政务服务V.2026_06_02.pdf`},
		// 不应被误改
		{`C:\a\49xxc.com.zip`, `C:\a\49xxc.com.zip`},
		{`C:\a\6742_厦门一体化政务服务平台.zip`, `C:\a\6742_厦门一体化政务服务平台.zip`},
		{`C:\a\xm0080026060101817中华人民共和国居民身份证.ofd`, `C:\a\xm0080026060101817中华人民共和国居民身份证.ofd`},
		{`C:\a\11350200MB1816343GQT2000132026068407W.txt`, `C:\a\11350200MB1816343GQT2000132026068407W.txt`},
		{`C:\a\闽厦路政许〔2026〕127号(1).sspx`, `C:\a\闽厦路政许〔2026〕127号.sspx`},
		{`C:\a\厦门市教育局2026年第二次高中、中职教师资格认定预约人数安排表.xlsx`,
			`C:\a\厦门市教育局2026年第二次高中、中职教师资格认定预约人数安排表.xlsx`},
		{`C:\a\一体化平台项目重保报告-0619-端午V.2026_0622_1651.docx`,
			`C:\a\一体化平台项目重保报告-0619-端午V.2026_0622_1651.docx`},
	}
	for _, c := range cases {
		got := CleanName(c.in)
		if got != c.want {
			t.Errorf("CleanName(%q)\n  got = %q\n  want= %q", c.in, got, c.want)
		}
		base := filepath.Base(got)
		first := strings.Index(base, "V.")
		if first >= 0 && strings.Contains(base[first+1:], "V.") {
			t.Errorf("出现双 V. 叠加: %q", base)
		}
	}
}

func TestValidateDate(t *testing.T) {
	ok := []string{"20260604", "202606", "20260604163500", "202606041635"}
	for _, d := range ok {
		if !validateDate(d) {
			t.Errorf("validateDate(%q) should be true", d)
		}
	}
	bad := []string{"20260684", "20001301", "19991231", "20261301"}
	for _, d := range bad {
		if validateDate(d) {
			t.Errorf("validateDate(%q) should be false", d)
		}
	}
}

func TestApplyTemplate(t *testing.T) {
	got := applyTemplate("原文件", &TemplateOptions{Pattern: "A_#", Start: 1, Increment: 1, Digits: 3}, 0)
	if got != "A_001" {
		t.Errorf("got %q want A_001", got)
	}
	got = applyTemplate("原文件", &TemplateOptions{Pattern: "*V.#", Start: 1, Increment: 1, Digits: 2}, 0)
	if got != "原文件V.01" {
		t.Errorf("got %q want 原文件V.01", got)
	}
	got = applyTemplate("f", &TemplateOptions{Pattern: "#", Start: 10, Increment: 5, Digits: 4}, 1)
	if got != "0015" {
		t.Errorf("got %q want 0015", got)
	}
	// 空格补齐
	got = applyTemplate("f", &TemplateOptions{Pattern: "P_#", Start: 7, Increment: 1, Digits: 4, PadChar: " "}, 0)
	if got != "P_   7" {
		t.Errorf("got %q want P_   7", got)
	}
}

func TestApplyReplace(t *testing.T) {
	got := applyReplace("报告2026版", &ReplaceOptions{Find: "2026", Replace: "2027"})
	if got != "报告2027版" {
		t.Errorf("got %q", got)
	}
	got = applyReplace("AbCdef", &ReplaceOptions{Find: "abc", Replace: "x", IgnoreCase: true})
	if got != "xdef" {
		t.Errorf("got %q want xdef", got)
	}
	got = applyReplace("a1b1c1", &ReplaceOptions{Find: `\d+`, Replace: "#", UseRegex: true})
	if got != "a#b#c#" {
		t.Errorf("got %q want a#b#c#", got)
	}
}

func TestApplyAddRemove(t *testing.T) {
	got := applyAddRemove("doc", &AddRemoveOptions{Prefix: "pre_", Suffix: "_post"})
	if got != "pre_doc_post" {
		t.Errorf("got %q", got)
	}
	got = applyAddRemove("abcdef", &AddRemoveOptions{RemoveFrom: 3, RemoveCount: 2})
	if got != "abef" {
		t.Errorf("got %q want abef", got)
	}
	got = applyAddRemove("abcCOPYdef", &AddRemoveOptions{RemoveStr: "COPY"})
	if got != "abcdef" {
		t.Errorf("got %q", got)
	}
	got = applyAddRemove("hello", &AddRemoveOptions{InsertAt: 3, InsertStr: "XX"})
	if got != "heXXllo" {
		t.Errorf("got %q want heXXllo", got)
	}
}

func TestPlanComputeAndConflict(t *testing.T) {
	items := []*FileItem{
		NewFileItem("C:\\a\\report.docx"),
		NewFileItem("C:\\a\\data.xlsx"),
	}
	p := &Plan{
		Items:    items,
		Template: &TemplateOptions{Pattern: "F_#", Start: 1, Increment: 1, Digits: 2},
		Conflict: ConflictAuto,
	}
	p.Compute()
	if items[0].NewName != "F_01.docx" {
		t.Errorf("item0 = %q", items[0].NewName)
	}
	if items[1].NewName != "F_02.xlsx" {
		t.Errorf("item1 = %q", items[1].NewName)
	}
}

func TestPlanExecuteRealFS(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "月度报告-6月.docx")
	if err := os.WriteFile(src, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	// copy 模式：原文件保留，生成带版本号副本
	p := &Plan{
		Items:       []*FileItem{NewFileItem(src)},
		Version:     true,
		VersionMove: false,
		Conflict:    ConflictAuto,
	}
	results, err := p.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Error != "" {
		t.Fatalf("execute error: %s", results[0].Error)
	}
	entries, _ := os.ReadDir(dir)
	var versioned, origin int
	for _, e := range entries {
		if e.Name() == "月度报告-6月.docx" {
			origin++
		}
		if isVersioned(e.Name()) {
			versioned++
		}
	}
	if origin != 1 || versioned != 1 {
		t.Fatalf("copy 后应有 1 原 + 1 副本，实际 origin=%d versioned=%d", origin, versioned)
	}

	// move 模式：把原文件改名加版本号（实际是第二个带版本号文件）
	p2 := &Plan{
		Items:       []*FileItem{NewFileItem(src)},
		Version:     true,
		VersionMove: true,
		Conflict:    ConflictAuto,
	}
	if _, err := p2.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("move 后原文件应不存在: %s", src)
	}
}

func TestNewNameUpdate(t *testing.T) {
	in := "新建文本文档V.2026_0706_163543 - 副本.txt"
	got := newName(in)
	if !strings.Contains(got, "新建文本文档V.") {
		t.Errorf("newName(%q) = %q, should contain 新建文本文档V.", in, got)
	}
	if strings.Contains(got, "副本") {
		t.Errorf("newName should drop 副本, got %q", got)
	}
	first := strings.Index(got, "V.")
	if strings.Contains(got[first+1:], "V.") {
		t.Errorf("newName 不应叠加 V., got %q", got)
	}
}

func isVersioned(name string) bool {
	return reVersion.MatchString(strings.TrimSuffix(name, filepath.Ext(name)))
}
