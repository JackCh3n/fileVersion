package renamer

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// applyTemplate 按模板生成第 idx 个文件的新名称（不含扩展名）。
// 占位符： * 原文件名； # 序号。
func applyTemplate(name string, opt *TemplateOptions, idx int) string {
	if opt == nil {
		return name
	}
	pat := opt.Pattern
	if pat == "" {
		pat = "*"
	}
	digits := opt.Digits
	if digits < 1 {
		digits = 1
	}
	pad := opt.PadChar
	if pad != " " {
		pad = "0"
	}

	var token string
	if opt.Random {
		token = randomToken(digits, opt.RandomLower)
	} else {
		n := opt.Start + idx*opt.Increment
		s := strconv.Itoa(n)
		if len(s) < digits {
			s = strings.Repeat(pad, digits-len(s)) + s
		}
		token = s
	}

	// 替换占位符：先处理 #，再把 * 替换为原名
	res := strings.ReplaceAll(pat, "#", token)
	res = strings.ReplaceAll(res, "*", name)
	return res
}

// randomToken 生成长度为 n 的随机编号（数字或字母）。
func randomToken(n int, lower bool) string {
	const digits = "0123456789"
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lowerS = "abcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	for i := 0; i < n; i++ {
		if r.Intn(2) == 0 || !lower {
			b.WriteByte(digits[r.Intn(len(digits))])
		} else {
			b.WriteByte(lowerS[r.Intn(len(lowerS))])
		}
	}
	_ = upper
	return b.String()
}
