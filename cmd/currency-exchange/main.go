package main

import (
	"log/slog"
	"os"

	"github.com/Endcru/PlataTestAssignment/internal/config"
	"github.com/Endcru/PlataTestAssignment/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))
	log.Info("Initialize server", slog.String("address", cfg.HTTPServer.Address))
	storage, err := postgres.NewPostgresStorage(cfg)
	if err != nil {
		log.Error("Failed to create storage", "error", err)
		os.Exit(1)
	}
	log.Info("Storage created successfully")
	defer storage.Close()
	err = postgres.AddBaseStatements(storage)
	if err != nil {
		log.Error("Failed to add base statements", "error", err)
		os.Exit(1)
	}
	log.Info("Base statements added successfully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case envDev:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case envProd:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return log
}