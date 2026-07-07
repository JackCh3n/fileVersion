package renamer

import (
	"regexp"
	"strings"
)

// applyReplace 按替换选项处理文件名（不含扩展名）。
func applyReplace(name string, opt *ReplaceOptions) string {
	if opt == nil || opt.Find == "" {
		return name
	}
	find := opt.Find
	repl := opt.Replace
	if opt.UseRegex {
		flags := ""
		if opt.IgnoreCase {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + find)
		if err != nil {
			return name // 正则非法则原样返回
		}
		return re.ReplaceAllString(name, repl)
	}
	if opt.IgnoreCase {
		return replaceIgnoreCase(name, find, repl)
	}
	return strings.ReplaceAll(name, find, repl)
}

func replaceIgnoreCase(s, old, repl string) string {
	var b strings.Builder
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)
	i := 0
	for {
		idx := strings.Index(lowerS[i:], lowerOld)
		if idx < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i : i+idx])
		b.WriteString(repl)
		i += idx + len(old)
	}
	return b.String()
}

// applyAddRemove 按前后缀/位置增删选项处理文件名（不含扩展名）。
func applyAddRemove(name string, opt *AddRemoveOptions) string {
	if opt == nil {
		return name
	}
	// 删除子串（优先于位置删除）
	if opt.RemoveStr != "" {
		name = strings.Replace(name, opt.RemoveStr, "", 1)
	}
	// 位置删除：从 RemoveFrom(1-based) 删除 RemoveCount 个字符
	if opt.RemoveFrom > 0 && opt.RemoveCount > 0 {
		from := opt.RemoveFrom - 1
		if from < len(name) {
			end := from + opt.RemoveCount
			if end > len(name) {
				end = len(name)
			}
			name = name[:from] + name[end:]
		}
	}
	// 前缀 / 后缀
	if opt.Prefix != "" {
		name = opt.Prefix + name
	}
	if opt.Suffix != "" {
		name = name + opt.Suffix
	}
	// 位置插入
	if opt.InsertAt > 0 && opt.InsertStr != "" {
		at := opt.InsertAt - 1
		if at > len(name) {
			at = len(name)
		}
		name = name[:at] + opt.InsertStr + name[at:]
	}
	return name
}
