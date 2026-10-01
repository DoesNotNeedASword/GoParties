package main

import (
	"context"
	"log/slog"
	"os"

	"Parties/internal/party/repository"
	partyservice "Parties/internal/party/service"
	"Parties/internal/shared/config"
	"Parties/internal/shared/database"
	"Parties/internal/shared/logger"
	"Parties/internal/shared/migration"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel)
	ctx := context.Background()

	if cfg.RunMigrations {
		if err := migration.Run(cfg.DatabaseURL, log); err != nil {
			log.Error("run migrations", "error", err)
			os.Exit(1)
		}
	}

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	partyRepository := repository.New(pool)
	_ = partyservice.New(partyRepository)

	log.Info("party service initialized", "http_addr", cfg.HTTPAddr)
}
