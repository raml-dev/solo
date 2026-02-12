package main

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"yapla/internal/tools"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed all:frontend/dist
var assets embed.FS
var programLevel = new(slog.LevelVar) // default info

func getLogPath() string {
	baseDir, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return "app.log"
	}

	logDir := filepath.Join(baseDir, "logs")

	os.MkdirAll(logDir, 0755)

	return filepath.Join(logDir, "yapla.log")
}

func main() {

	logWriter := &lumberjack.Logger{
		Filename:   getLogPath(),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     1, //days
		Compress:   true,
	}

	logger := slog.New(slog.NewJSONHandler(logWriter,
		&slog.HandlerOptions{
			Level: programLevel,
		}))

	slog.SetDefault(logger)

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "yapla",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
