// Package renamer 提供批量文件重命名的核心引擎。
// 支持：模板命名(整体)、查找替换、前后缀/位置增删、clean 整理、
// 版本号(copy/move)、冲突处理、预览(dry-run)与撤销(revert)。
// 所有逻辑仅依赖标准库，可在任意平台编译与单测。
package renamer

import (
	"path/filepath"
	"strings"
)

// FileItem 表示一个待处理的文件。
type FileItem struct {
	Path    string // 完整路径
	Dir     string // 所在目录
	Base    string // 含扩展名的文件名
	Name    string // 不含扩展名的名称
	Ext     string // 扩展名（含点，如 .txt）
	OldName string // 原始 Base（撤销/结果显示用）
	NewName string // 计算后的新 Base
	Preview string // 预览结果（同 NewName，供前端展示）
	Status  string // "" 正常 / "conflict" 冲突 / "unchanged" 无变化 / "error" 出错
	Err     string // 错误信息
}

// NewFileItem 由路径构造一个 FileItem。
func NewFileItem(path string) *FileItem {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return &FileItem{
		Path:    path,
		Dir:     filepath.Dir(path),
		Base:    base,
		Name:    name,
		Ext:     ext,
		OldName: base,
	}
}

// ConflictStrategy 冲突处理策略。
const (
	ConflictSkip      = "skip"      // 跳过（保留原名）
	ConflictOverwrite = "overwrite" // 覆盖已存在文件
	ConflictAuto      = "auto"      // 自动添加序号 (_1, _2 ...) 避免冲突
)

// TemplateOptions 整体(模板)重命名选项，对应截图“整体”Tab。
// Pattern 支持占位符：
//
//   - 原文件名（插入位置）
//     #  序号（插入位置，按 Start/Increment/Digits 生成）
//
// 例： "A_#" → A_001 ; "*V.#" → 原文件V.001
type TemplateOptions struct {
	Pattern      string // 命名规则，如 A_# , *V.#
	Start        int    // 起始序号
	Increment    int    // 增量
	Digits       int    // 位数（不足补 PadChar）
	PadChar      string // 补齐字符："0" 或 " "
	Random       bool   // 随机编号（替代顺序序号）
	RandomLower  bool   // 随机编号使用小写字母
	ExtOverride  string // 扩展名改写（不含点；空=不改）
	AutoConflict bool   // 自动解决命名冲突
}

// ReplaceOptions 查找替换选项，对应截图“替换”Tab。
type ReplaceOptions struct {
	Find       string // 查找内容
	Replace    string // 替换为
	UseRegex   bool   // 使用正则
	IgnoreCase bool   // 忽略大小写
}

// AddRemoveOptions 前后缀/位置增删选项，对应截图“添加/删除”Tab。
type AddRemoveOptions struct {
	Prefix      string // 文件名前添加
	Suffix      string // 文件名后添加
	InsertAt    int    // 从左侧第 N 个字符插入（1-based；0=不插入）
	InsertStr   string // 插入字符串
	RemoveStr   string // 删除该子串（首次出现）
	RemoveFrom  int    // 从左侧第 N 个字符开始删除（1-based；0=不删）
	RemoveCount int    // 删除 N 个字符
}

// Plan 一次批量重命名计划。
// 可按需组合 Template / Replace / AddRemove；执行顺序为：
// 1) Template（若设置）  2) Replace（若设置）  3) AddRemove（若设置）
// 之后统一做 clean（若 Clean=true）与版本号（若 Version=true）。
type Plan struct {
	Items       []*FileItem
	Template    *TemplateOptions
	Replace     *ReplaceOptions
	AddRemove   *AddRemoveOptions
	Clean       bool   // 整理：去 (N)、日期规整为 V.YYYY_MM_DD
	Version     bool   // 追加/更新版本号后缀（copy/move 行为）
	VersionMove bool   // Version 为真时：true=原地改名(move)，false=复制(copy)
	Conflict    string // 冲突策略：skip/overwrite/auto
}
