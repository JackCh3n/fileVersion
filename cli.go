package main

import (
	"fileversion/renamer"
	"fileversion/shell"
)

// runCLI 处理子命令：install / uninstall / copy / move / clean。
func runCLI(args []string) {
	cmd := args[0]
	switch cmd {
	case "install":
		if err := shell.Install(); err != nil {
			msgBoxErr("安装失败", err.Error())
			return
		}
		msgBox("FileVersion", "安装完成。右键文件即可看到“FileVersion 批量改名”等菜单。")
	case "uninstall":
		if err := shell.Uninstall(); err != nil {
			msgBoxErr("卸载失败", err.Error())
			return
		}
		msgBox("FileVersion", "已卸载，右键菜单与快捷方式已移除。")
	case "copy", "move", "clean":
		files := args[1:]
		if len(files) == 0 {
			msgBoxErr("用法", "fileversion "+cmd+" \"文件1\" \"文件2\" ...")
			return
		}
		if err := runCLIMode(cmd, files); err != nil {
			notify("FileVersion", "操作失败："+err.Error(), true)
			return
		}
		notify("FileVersion", "操作完成（"+cmd+"）", false)
	default:
		msgBox("FileVersion 用法",
			"install   安装右键菜单与快捷方式\n"+
				"uninstall 卸载\n"+
				"copy      复制并加版本号\n"+
				"move      重命名加版本号（已存在则更新）\n"+
				"clean     整理文件名（去(N)、日期规整V.YYYY_MM_DD）\n"+
				"（不带参数运行则打开 GUI）")
	}
}

// runCLIMode 对给定文件执行 copy/move/clean。
func runCLIMode(mode string, files []string) error {
	plan := &renamer.Plan{Items: make([]*renamer.FileItem, 0, len(files))}
	for _, f := range files {
		plan.Items = append(plan.Items, renamer.NewFileItem(f))
	}
	switch mode {
	case "copy":
		plan.Version = true
		plan.VersionMove = false
	case "move":
		plan.Version = true
		plan.VersionMove = true
	case "clean":
		plan.Clean = true
	}
	_, err := plan.Execute()
	return err
}
