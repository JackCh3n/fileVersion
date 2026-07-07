package main

import (
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// 若提供了子命令（install/uninstall/copy/move/clean），走 CLI 模式；
// 否则启动 GUI。
func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		runCLI(args)
		return
	}
	app := NewApp()
	err := wails.Run(&options.App{
		Title:  "FileVersion 批量改名",
		Width:  app.cfg.WindowWidth,
		Height: app.cfg.WindowHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup: func(ctx *wails.Context) {
			// 可在此注入 ctx（如需从 Go 主动调用 JS）
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Println("GUI 启动失败:", err)
		os.Exit(1)
	}
}
