package main

import (
	"log/slog"
	"nora/app"
	"os"
)

func main() {
	cfg, err := app.InitConfig("config.json")
	if err != nil {
		slog.Error("Failed to read config", "error", err)
		os.Exit(1)
	}

	f, err := app.InitLogger(cfg.LogLevel, cfg.LogDir)
	if err != nil {
		slog.Error("Failed to create log file", "error", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := app.StartServer(cfg); err != nil {
		slog.Error("Server stopped", "error", err)
		os.Exit(1)
	}
}
