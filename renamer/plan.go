package renamer

import (
	"os"
	"path/filepath"
	"strings"
)

// Compute 计算所有文件的新名称，填充 Preview / NewName / Status，但不实际执行。
// 这是“预览(dry-run)”的核心：前端在点击“开始重命名”前调用。
func (p *Plan) Compute() {
	for i, it := range p.Items {
		name := it.Name
		ext := it.Ext

		// 1) 模板
		name = applyTemplate(name, p.Template, i)
		// 2) 替换
		name = applyReplace(name, p.Replace)
		// 3) 前后缀/增删
		name = applyAddRemove(name, p.AddRemove)
		// 4) clean 整理
		if p.Clean {
			full := name + ext
			full = CleanName(full)
			ext = filepath.Ext(full)
			name = strings.TrimSuffix(full, ext)
		}
		// 5) 版本号
		if p.Version {
			full := name + ext
			full = newName(full)
			ext = filepath.Ext(full)
			name = strings.TrimSuffix(full, ext)
		}
		// 扩展名改写
		if p.Template != nil && p.Template.ExtOverride != "" {
			ext = "." + strings.TrimPrefix(p.Template.ExtOverride, ".")
		}

		it.NewName = name + ext
		it.Preview = it.NewName
		if it.NewName == it.OldName {
			it.Status = "unchanged"
		} else {
			it.Status = ""
		}
	}
	p.resolveConflicts()
}

// resolveConflicts 在计算出 NewName 后，按策略处理批次内/已存在的冲突。
func (p *Plan) resolveConflicts() {
	strategy := p.Conflict
	if strategy == "" {
		strategy = ConflictSkip
	}
	seen := make(map[string]int) // 目标名 -> 已出现次数
	for _, it := range p.Items {
		if it.Status == "unchanged" {
			continue
		}
		target := it.NewName
		// 批次内重复
		if cnt, ok := seen[target]; ok {
			switch strategy {
			case ConflictSkip:
				it.Status = "conflict"
				it.Err = "目标名在批次内重复，已跳过"
				it.NewName = it.OldName
				it.Preview = it.OldName
				continue
			case ConflictOverwrite:
				// 允许覆盖，不处理
			case ConflictAuto:
				it.NewName = uniqueName(target, cnt+1, seen)
				it.Preview = it.NewName
			}
		}
		// 与磁盘已有文件冲突
		dst := filepath.Join(it.Dir, it.NewName)
		if _, err := os.Stat(dst); err == nil && it.NewName != it.OldName {
			switch strategy {
			case ConflictSkip:
				it.Status = "conflict"
				it.Err = "目标文件已存在，已跳过"
				it.NewName = it.OldName
				it.Preview = it.OldName
				continue
			case ConflictOverwrite:
				// 允许覆盖
			case ConflictAuto:
				cnt := seen[target] + 1
				it.NewName = uniqueName(target, cnt, seen)
				it.Preview = it.NewName
			}
		}
		seen[it.NewName]++
	}
}

// uniqueName 在 base 基础上追加 _N 直到不重复。
func uniqueName(base string, n int, seen map[string]int) string {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	for {
		cand := name + "_" + itoa(n) + ext
		if _, ok := seen[cand]; !ok {
			return cand
		}
		n++
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// Result 单条执行结果。
type Result struct {
	OldPath string
	NewPath string
	OldName string
	NewName string
	Skipped bool
	Error   string
}

// Execute 实际执行重命名（或复制）并返回结果列表。
// 返回的结果可用于写入撤销日志。
func (p *Plan) Execute() ([]Result, error) {
	p.Compute() // 确保已计算
	var results []Result
	for _, it := range p.Items {
		r := Result{OldPath: it.Path, OldName: it.OldName, NewName: it.NewName}
		if it.Status == "unchanged" {
			r.Skipped = true
			results = append(results, r)
			continue
		}
		if it.Status == "conflict" {
			r.Skipped = true
			r.Error = it.Err
			results = append(results, r)
			continue
		}
		dst := filepath.Join(it.Dir, it.NewName)
		if p.Version && p.VersionMove {
			// move：原地改名
			if err := os.Rename(it.Path, dst); err != nil {
				r.Error = err.Error()
				results = append(results, r)
				continue
			}
		} else if p.Version && !p.VersionMove {
			// copy：复制并加版本号
			if err := copyFile(it.Path, dst); err != nil {
				r.Error = err.Error()
				results = append(results, r)
				continue
			}
		} else {
			// 普通重命名
			if err := os.Rename(it.Path, dst); err != nil {
				r.Error = err.Error()
				results = append(results, r)
				continue
			}
		}
		r.NewPath = dst
		it.Path = dst
		results = append(results, r)
	}
	return results, nil
}

// copyFile 复制文件（用于 copy/version 模式）。
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 32*1024)
	for {
		n, e := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if e != nil {
			break
		}
	}
	return nil
}
