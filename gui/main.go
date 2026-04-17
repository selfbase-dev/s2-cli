package main

import (
	"context"
	"embed"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// Register the systray BEFORE wails.Run. systray.Register is
	// non-blocking on macOS 10.15+ — it sets up NSStatusItem callbacks
	// and returns immediately, then relies on Wails' NSApplication run
	// loop to dispatch target-actions. This sidesteps the
	// main-thread-ownership conflict that systray.Run would cause.
	systray.Register(func() { onTrayReady(app) }, func() {})

	err := wails.Run(&options.App{
		Title:             "s2sync",
		Width:             520,
		Height:            640,
		MinWidth:          420,
		MinHeight:         520,
		HideWindowOnClose: true,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "dev.selfbase.s2sync",
			OnSecondInstanceLaunch: func(_ options.SecondInstanceData) {
				if app.ctx != nil {
					wailsruntime.WindowShow(app.ctx)
				}
			},
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(_ context.Context) {
			systray.Quit()
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("wails error:", err.Error())
	}
}
