package app

import (
	"context"
	"fmt"
	"github.com/Ostmind/subscriptionservice/internal/storage/postgres"
	"github.com/Ostmind/subscriptionservice/internal/subscription/config"
	srv "github.com/Ostmind/subscriptionservice/internal/subscription/server/server"
	"log/slog"
)

type App struct {
	server *srv.Server
	logger *slog.Logger
	db     *postgres.Storage
	cfg    *config.AppConfig
}

func New(logger *slog.Logger, cfg *config.AppConfig) (*App, error) {
	db, err := postgres.New(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("couldn't establish db connection %w", err)
	}

	server := srv.New(db, logger)

	return &App{
		server: server,
		logger: logger,
		db:     db,
	}, nil
}

func (a App) Run(serverHost string, serverPort int) {
	a.logger.Info("Starting app...")

	a.server.Run(serverHost, serverPort)
}

func (a App) Stop(ctx context.Context) {
	a.logger.Info("Stopping app...")

	doneCh := make(chan error)
	go func() {
		doneCh <- a.server.Stop(ctx)
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			a.logger.Error("Error while stopping server: %v", slog.Any("error_details", err))
		}

		a.logger.Info("App has been stopped gracefully")

	case <-ctx.Done():
		a.logger.Warn("App stopped forced")
	}
}
