// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"solo/internal/tools"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/linux/icon.png
var icon []byte
var programLevel = new(slog.LevelVar) // default info

type AppLogger struct {
	base *slog.Logger
}

func (l *AppLogger) Print(msg string)   { l.base.Info(msg, "source", "frontend") }
func (l *AppLogger) Trace(msg string)   { l.base.Debug(msg, "source", "frontend") }
func (l *AppLogger) Debug(msg string)   { l.base.Debug(msg, "source", "frontend") }
func (l *AppLogger) Info(msg string)    { l.base.Info(msg, "source", "frontend") }
func (l *AppLogger) Warning(msg string) { l.base.Warn(msg, "source", "frontend") }
func (l *AppLogger) Error(msg string)   { l.base.Error(msg, "source", "frontend") }
func (l *AppLogger) Fatal(msg string)   { l.base.Error(msg, "source", "frontend", "fatal", true) }

func getLogPath() string {
	baseDir, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return tools.APP_NAME + ".log"
	}

	logDir := filepath.Join(baseDir, "logs")

	os.MkdirAll(logDir, 0755)

	return filepath.Join(logDir, tools.APP_NAME+".log")
}

func main() {

	// Create an instance of the app structure
	app := NewApp()

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

	mainVersion := "dev"
	if app.GetAppInfo().ProductVersion != "" {
		mainVersion = app.GetAppInfo().ProductVersion
	}

	logger = logger.With("version", mainVersion)

	slog.SetDefault(logger)

	// Create application with options
	err := wails.Run(&options.App{
		Title:       tools.APP_NAME_CAPITALIZED,
		Width:       1024,
		Height:      768,
		StartHidden: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger:           &AppLogger{base: logger},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
		},
		Bind: []interface{}{
			app,
		},
		Linux: &linux.Options{
			Icon: icon,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
