// Package config 管理 FileVersion 的 JSON 配置文件。
// 配置存放于 %APPDATA%/FileVersion/config.json，记录用户偏好。
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 用户配置。
type Config struct {
	ConflictStrategy  string `json:"conflictStrategy"`  // skip / overwrite / auto
	TemplateStart     int    `json:"templateStart"`     // 编号起始
	TemplateIncrement int    `json:"templateIncrement"` // 编号增量
	TemplateDigits    int    `json:"templateDigits"`    // 编号位数
	CopyOutputDir     string `json:"copyOutputDir"`     // copy 模式输出目录（空=同目录）
	CleanKeepTime     bool   `json:"cleanKeepTime"`     // clean 是否保留完整时间戳
	WindowWidth       int    `json:"windowWidth"`
	WindowHeight      int    `json:"windowHeight"`
}

// Default 返回默认配置。
func Default() *Config {
	return &Config{
		ConflictStrategy:  "auto",
		TemplateStart:     1,
		TemplateIncrement: 1,
		TemplateDigits:    3,
		CopyOutputDir:     "",
		CleanKeepTime:     false,
		WindowWidth:       900,
		WindowHeight:      640,
	}
}

// Path 返回配置文件路径。
func Path() string {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "FileVersion", "config.json")
}

// Load 读取配置，不存在则返回默认。
func Load() *Config {
	p := Path()
	data, err := os.ReadFile(p)
	if err != nil {
		return Default()
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Default()
	}
	// 字段缺省兜底
	d := Default()
	if c.ConflictStrategy == "" {
		c.ConflictStrategy = d.ConflictStrategy
	}
	if c.TemplateDigits == 0 {
		c.TemplateDigits = d.TemplateDigits
	}
	if c.TemplateStart == 0 {
		c.TemplateStart = d.TemplateStart
	}
	if c.TemplateIncrement == 0 {
		c.TemplateIncrement = d.TemplateIncrement
	}
	if c.WindowWidth == 0 {
		c.WindowWidth = d.WindowWidth
	}
	if c.WindowHeight == 0 {
		c.WindowHeight = d.WindowHeight
	}
	return &c
}

// Save 保存配置到文件。
func Save(c *Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
