package main

import (
	"context"
	"embed"
	"flagged-it/internal/api"
	"log"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

func main() {
	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Flagged It",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appStartup,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func appStartup(ctx context.Context) {
	// Start API server on a different port for the desktop app
	go func() {
		api.SetupRoutes()
		log.Println("API server starting on http://localhost:8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("API server error: %v", err)
		}
	}()
}
