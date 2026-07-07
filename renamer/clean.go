package renamer

import (
	"regexp"
	"strconv"
	"strings"
)

// clean 相关正则（从 main.go 迁移并保留行为）。
var (
	reCounter    = regexp.MustCompile(`\(\d+\)\s*$`)                      // 尾部 Windows 复制计数 (N)
	reDateCand   = regexp.MustCompile(`20\d{2}(?:[-_/. ]?\d{2}){1,6}\d?`) // 日期候选（含完整时间戳）
	reHasVersion = regexp.MustCompile(`V\.\d{4}(?:_\d{2,6}){0,3}`) // 已是 V. 形态（V.年 / V.年_月 / V.年_月_日 / V.年_月日_时分秒）
)

// digitsOnly 仅保留数字字符。
func digitsOnly(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
}

// validateDate 校验字符串是否为合法日期/时间戳。
// 支持长度：6(YYYYMM) / 8(YYYYMMDD) / 10,12,14,15+(YYYYMMDD[HHMM[SS]]，多余尾部数字忽略)。
func validateDate(d string) bool {
	d = digitsOnly(d)
	if len(d) < 6 {
		return false
	}
	if len(d) == 6 {
		y, _ := strconv.Atoi(d[0:4])
		mo, _ := strconv.Atoi(d[4:6])
		return y >= 2000 && y <= 2099 && mo >= 1 && mo <= 12
	}
	if len(d) < 8 {
		return false
	}
	y, _ := strconv.Atoi(d[0:4])
	mo, _ := strconv.Atoi(d[4:6])
	da, _ := strconv.Atoi(d[6:8])
	if !validYMD(y, mo, da) {
		return false
	}
	if len(d) >= 12 {
		hh, _ := strconv.Atoi(d[8:10])
		mm, _ := strconv.Atoi(d[10:12])
		if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
			return false
		}
	}
	if len(d) == 14 {
		ss, _ := strconv.Atoi(d[12:14])
		if ss < 0 || ss > 59 {
			return false
		}
	}
	return true
}

func validYMD(y, mo, da int) bool {
	return y >= 2000 && y <= 2099 && mo >= 1 && mo <= 12 && da >= 1 && da <= 31
}

// extractDate 从文件名中提取合法日期，返回 (规范化V串, 是否找到)。
// clean 仅保留日期部分（丢弃时分秒）：YYYYMMDD/YYYYMMDDHHMMSS/... → V.YYYY_MM_DD；YYYYMM → V.YYYY_MM。
func extractDate(name string) (string, bool) {
	locs := reDateCand.FindAllStringIndex(name, -1)
	for _, loc := range locs {
		cand := name[loc[0]:loc[1]]
		dig := digitsOnly(cand)
		if !validateDate(dig) {
			continue
		}
		if len(dig) >= 8 {
			return "V." + dig[0:4] + "_" + dig[4:6] + "_" + dig[6:8], true
		}
		if len(dig) == 6 {
			return "V." + dig[0:4] + "_" + dig[4:6], true
		}
	}
	return "", false
}

// CleanName 规整文件名：去 (N) 计数、提取日期为 V.YYYY_MM_DD 移到末尾。
// 已是 V. 形态的文件只去 (N)，不降级；无日期则仅去 (N)。
// 入参可为完整路径，目录部分会被保留不变。
func CleanName(p string) string {
	// 分离目录与文件名
	dir := ""
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		dir = p[:i+1]
		p = p[i+1:]
	}
	ext := ""
	if i := strings.LastIndex(p, "."); i > 0 {
		ext = p[i:]
		p = p[:i]
	}
	// 已是 V. 形态：仅去残留 (N)
	if reHasVersion.MatchString(p) {
		p = reCounter.ReplaceAllString(p, "")
		return dir + p + ext
	}
	// 去 (N)
	p = reCounter.ReplaceAllString(p, "")
	// 提取日期
	if v, ok := extractDate(p); ok {
		p = reDateCand.ReplaceAllString(p, "")
		p = strings.Trim(p, " _-.") // 清理两端残留分隔符
		if p == "" {
			p = v
		} else {
			p = p + v
		}
	}
	return dir + p + ext
}
