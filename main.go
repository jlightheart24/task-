package main

import (
	"embed"
	"fmt"

	"taskpp/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed cmd/desktop/assets/*
var assets embed.FS

func main() {
	appInstance, err := app.New("dev", "file:taskminus.db?_pragma=foreign_keys(1)")
	if err != nil {
		panic(fmt.Errorf("init app: %w", err))
	}

	err = wails.Run(&options.App{
		Title:  "task-",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			appInstance,
		},
	})
	if err != nil {
		panic(err)
	}
}
