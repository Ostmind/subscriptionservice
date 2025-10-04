package main

import (
	"context"
	application "github.com/Ostmind/subscriptionservice/internal/subscription/app"
	"github.com/Ostmind/subscriptionservice/internal/subscription/config"
	"github.com/Ostmind/subscriptionservice/internal/subscription/logger"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustNew()

	sloger := logger.SetupLogger(cfg.LogLevel)

	sloger.Info("starting Subscription Service")

	app, err := application.New(sloger, cfg)
	if err != nil {
		log.Fatal("No App cannot start server", slog.String("err", err.Error()))
	}

	go app.Run(cfg.Srv.Port)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM)

	<-stopChan
	sloger.Info("Received interrupt signal")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Srv.ServerShutdownTimeout)
	defer cancel()

	app.Stop(ctx)
}
