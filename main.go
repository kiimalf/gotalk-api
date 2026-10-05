package main

import (
	"context"
	"log/slog"
	"os"

	"gotalk-api/config"
	"gotalk-api/database"
	"gotalk-api/route"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("Gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	app := config.NewApp(logger, route.Dependencies{
		Pool: pool,
	})

	port := config.GetEnv("APP_PORT", "3000")
	logger.Info("Server berjalan", slog.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		logger.Error("Server berhenti", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
