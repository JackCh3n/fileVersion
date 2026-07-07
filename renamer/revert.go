package renamer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RevertLog 撤销日志：记录一次批量操作的全部改名映射。
type RevertLog struct {
	Time  string        `json:"time"`
	Items []RevertEntry `json:"items"`
}

// RevertEntry 单条改名记录。
type RevertEntry struct {
	Old string `json:"old"` // 原始完整路径
	New string `json:"new"` // 改名后完整路径
}

// WriteRevertLog 把执行结果写入撤销日志文件（默认 %APPDATA%/FileVersion/revert.json）。
func WriteRevertLog(results []Result, logPath string) error {
	if logPath == "" {
		logPath = DefaultRevertPath()
	}
	entries := make([]RevertEntry, 0, len(results))
	for _, r := range results {
		if r.Skipped || r.Error != "" || r.NewPath == "" {
			continue
		}
		entries = append(entries, RevertEntry{Old: r.OldPath, New: r.NewPath})
	}
	if len(entries) == 0 {
		return nil
	}
	log := RevertLog{Time: time.Now().Format(time.RFC3339), Items: entries}
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(logPath, data, 0o644)
}

// DefaultRevertPath 默认撤销日志路径（%APPDATA%/FileVersion/revert.json）。
func DefaultRevertPath() string {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "FileVersion", "revert.json")
}

// RevertLast 读取撤销日志，把 New→Old 还原。返回成功还原的条数。
func RevertLast(logPath string) (int, error) {
	if logPath == "" {
		logPath = DefaultRevertPath()
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		return 0, err
	}
	var log RevertLog
	if err := json.Unmarshal(data, &log); err != nil {
		return 0, err
	}
	n := 0
	for _, e := range log.Items {
		// 仅当当前文件是 New（改名后）才还原为 Old
		if _, err := os.Stat(e.New); err != nil {
			continue
		}
		if err := os.Rename(e.New, e.Old); err == nil {
			n++
		}
	}
	return n, nil
}
