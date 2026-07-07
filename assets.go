package main

import (
	"embed"
	"io/fs"
)

//go:embed all:frontend
var frontendFS embed.FS

// assets 将前端目录暴露给 Wails AssetServer。
var assets fs.FS

func init() {
	sub, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		panic(err)
	}
	assets = sub
}
