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

// @title Subscription Service API
// @version 1.0
// @description API для сервиса подписок
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @schemes http https
func main() {
	cfg := config.MustNew()

	sloger := logger.SetupLogger(cfg.LogLevel)

	sloger.Info("starting Subscription Service")

	app, err := application.New(sloger, cfg)
	if err != nil {
		log.Fatal("No App cannot start server", slog.String("err", err.Error()))
	}

	go app.Run(cfg.Srv.Host, cfg.Srv.Port)

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
