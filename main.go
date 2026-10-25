// calc-app is a multi-mode desktop calculator built with Wails v3:
// normal, scientific, developer and finance modes over a Go calculation
// backend and a SvelteKit frontend.
package main

import (
	"embed"
	"log"
	"os"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"

	"calc-app/backend/services"
)

//go:embed all:frontend/dist
var assets embed.FS

func serverPort() int {
	if value := os.Getenv("CALC_PORT"); value != "" {
		if port, err := strconv.Atoi(value); err == nil && port > 0 && port < 65536 {
			return port
		}
	}
	return 34115
}

func main() {
	history := &services.HistoryService{}
	if err := history.Load(); err != nil {
		log.Printf("history: %v", err)
	}
	settings := &services.SettingsService{}
	if err := settings.Load(); err != nil {
		log.Printf("settings: %v", err)
	}

	app := application.New(application.Options{
		Name:        "Calc",
		Description: "A multi-mode calculator: normal, scientific, developer and finance.",
		Services: []application.Service{
			application.NewService(&services.CalcService{}),
			application.NewService(history),
			application.NewService(settings),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		// Used by the browser preview build (go build -tags server).
		Server: application.ServerOptions{
			Host: "localhost",
			Port: serverPort(),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	history.OnChanged = func() {
		app.Event.Emit("history:changed")
	}
	settings.OnChanged = func() {
		app.Event.Emit("settings:changed", settings.Get())
	}

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "Calc",
		Width:          420,
		Height:         700,
		MinWidth:       360,
		MinHeight:      560,
		URL:            "/",
		StartState:     application.WindowStateNormal,
		BackgroundType: application.BackgroundTypeSolid,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
