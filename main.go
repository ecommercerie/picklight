package main

import (
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Handle --install flag (running elevated to perform installation)
	if len(os.Args) > 1 && os.Args[1] == "--install" {
		if err := selfInstall(); err != nil {
			showInstallError(err.Error())
			os.Exit(1)
		}
		relaunchInstalled()
		os.Exit(0)
	}

	// First launch: not installed yet — ask user
	if needsInstall() {
		if showInstallDialog() {
			if err := runElevated(); err != nil {
				showInstallError(err.Error())
			}
			// Exit this instance — the elevated one will install and relaunch
			os.Exit(0)
		}
		// User declined install — run from current location anyway
	}

	if !acquireSingleInstance() {
		return
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "PickLight",
		Width:             800,
		Height:            650,
		MinWidth:          600,
		MinHeight:         450,
		HideWindowOnClose: true,
		StartHidden:       true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatalf("PickLight failed to start: %v", err)
	}
}
