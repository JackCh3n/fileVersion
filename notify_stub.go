//go:build !windows

package main

// 非 Windows 平台占位实现（GUI 与右键菜单仅支持 Windows）。
func msgBox(title, text string)              {}
func msgBoxErr(title, text string)           {}
func notify(title, msg string, isError bool) {}
