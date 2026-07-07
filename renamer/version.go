package renamer

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// reVersion 识别版本号标记：兼容 4 位(分) 与 6 位(秒) 两种历史精度。
var reVersion = regexp.MustCompile(`V\.\d{4}_\d{4}_\d{4,6}`)

// versionSuffix 生成当前时间戳后缀 V.YYYY_MMDD_HHMMSS（精确到秒）。
func versionSuffix(t time.Time) string {
	return "V." + t.Format("2006_0102_150405")
}

// newName 基于已有文件名生成带版本号的新名（版本号插在扩展名之前）。
// 若原文件名已含 V. 标记，则从首个 V. 处截断并替换为最新时间戳，
// 同时丢弃其后可能的“ - 副本”等残留；否则在名称末尾（扩展名前）追加。
func newName(base string) string {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	loc := reVersion.FindStringIndex(name)
	if loc != nil {
		return name[:loc[0]] + versionSuffix(time.Now()) + ext
	}
	return name + versionSuffix(time.Now()) + ext
}
