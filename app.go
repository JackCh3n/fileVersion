package main

import (
	"fmt"
	"sort"

	"fileversion/config"
	"fileversion/renamer"
	"fileversion/shell"
)

// App 结构体暴露给前端(Wails)的方法集合。
type App struct {
	files   []*renamer.FileItem
	cfg     *config.Config
	lastLog string
}

// NewApp 创建 App 实例。
func NewApp() *App {
	return &App{
		files: make([]*renamer.FileItem, 0),
		cfg:   config.Load(),
	}
}

// FileView 前端展示用的文件视图（JSON 友好）。
type FileView struct {
	Path    string `json:"path"`
	OldName string `json:"oldName"`
	Preview string `json:"preview"`
	Status  string `json:"status"`
	Err     string `json:"err"`
}

// PlanOptions 前端传来的计划参数。
type PlanOptions struct {
	Tab         string                    `json:"tab"` // overall / replace / addremove
	Template    *renamer.TemplateOptions  `json:"template"`
	Replace     *renamer.ReplaceOptions   `json:"replace"`
	AddRemove   *renamer.AddRemoveOptions `json:"addRemove"`
	Clean       bool                      `json:"clean"`
	Version     bool                      `json:"version"`
	VersionMove bool                      `json:"versionMove"`
	Conflict    string                    `json:"conflict"`
}

// AddFiles 添加文件到列表（支持路径数组）。
func (a *App) AddFiles(paths []string) []FileView {
	for _, p := range paths {
		a.files = append(a.files, renamer.NewFileItem(p))
	}
	return a.view()
}

// RemoveFile 移除指定路径的文件。
func (a *App) RemoveFile(path string) []FileView {
	out := a.files[:0]
	for _, f := range a.files {
		if f.Path != path {
			out = append(out, f)
		}
	}
	a.files = out
	return a.view()
}

// ClearFiles 清空列表。
func (a *App) ClearFiles() []FileView {
	a.files = make([]*renamer.FileItem, 0)
	return a.view()
}

// ComputePreview 计算预览（dry-run，不实际执行）。
func (a *App) ComputePreview(opt PlanOptions) []FileView {
	plan := a.buildPlan(opt)
	plan.Compute()
	return a.view()
}

// ExecuteRename 实际执行重命名，返回结果并记录撤销日志。
func (a *App) ExecuteRename(opt PlanOptions) map[string]interface{} {
	plan := a.buildPlan(opt)
	results, err := plan.Execute()
	if err != nil {
		return map[string]interface{}{"ok": false, "error": err.Error()}
	}
	// 写撤销日志
	logPath := renamer.DefaultRevertPath()
	if e := renamer.WriteRevertLog(results, logPath); e == nil {
		a.lastLog = logPath
	}
	// 更新文件列表（路径可能已变）
	_ = results
	return map[string]interface{}{
		"ok":     true,
		"count":  len(results),
		"revert": a.lastLog,
		"files":  a.view(),
	}
}

// Revert 撤销上一次操作。
func (a *App) Revert() map[string]interface{} {
	n, err := renamer.RevertLast("")
	if err != nil {
		return map[string]interface{}{"ok": false, "error": err.Error()}
	}
	return map[string]interface{}{"ok": true, "count": n}
}

// GetConfig 获取配置。
func (a *App) GetConfig() *config.Config {
	return a.cfg
}

// SaveConfig 保存配置。
func (a *App) SaveConfig(c config.Config) error {
	a.cfg = &c
	return config.Save(&c)
}

// Install 安装右键菜单与快捷方式。
func (a *App) Install() error {
	return shell.Install()
}

// Uninstall 卸载。
func (a *App) Uninstall() error {
	return shell.Uninstall()
}

// buildPlan 由前端参数构造重命名计划。
func (a *App) buildPlan(opt PlanOptions) *renamer.Plan {
	conflict := opt.Conflict
	if conflict == "" {
		conflict = a.cfg.ConflictStrategy
	}
	// 应用配置默认值到模板
	if opt.Template != nil {
		if opt.Template.Start == 0 {
			opt.Template.Start = a.cfg.TemplateStart
		}
		if opt.Template.Increment == 0 {
			opt.Template.Increment = a.cfg.TemplateIncrement
		}
		if opt.Template.Digits == 0 {
			opt.Template.Digits = a.cfg.TemplateDigits
		}
	}
	return &renamer.Plan{
		Items:       a.files,
		Template:    opt.Template,
		Replace:     opt.Replace,
		AddRemove:   opt.AddRemove,
		Clean:       opt.Clean,
		Version:     opt.Version,
		VersionMove: opt.VersionMove,
		Conflict:    conflict,
	}
}

// view 把内部 files 转为前端视图（按目录+名称排序）。
func (a *App) view() []FileView {
	items := make([]*renamer.FileItem, len(a.files))
	copy(items, a.files)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Dir != items[j].Dir {
			return items[i].Dir < items[j].Dir
		}
		return items[i].OldName < items[j].OldName
	})
	out := make([]FileView, 0, len(items))
	for _, f := range items {
		out = append(out, FileView{
			Path:    f.Path,
			OldName: f.OldName,
			Preview: f.Preview,
			Status:  f.Status,
			Err:     f.Err,
		})
	}
	return out
}

// Greet 供前端测试连通性。
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, from FileVersion!", name)
}

// View 返回当前文件列表视图（供前端初始化加载）。
func (a *App) View() []FileView {
	return a.view()
}
