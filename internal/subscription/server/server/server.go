package server

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/Ostmind/subscriptionservice/docs"
	"github.com/Ostmind/subscriptionservice/internal/storage/postgres"
	"github.com/Ostmind/subscriptionservice/internal/subscription/server/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Server struct {
	server  *echo.Echo
	logger  *slog.Logger
	storage *postgres.Storage
}

func New(db *postgres.Storage, logger *slog.Logger) *Server {
	server := echo.New()

	server.Use(middleware.LogRequest(logger))
	subController := NewSubscriptionHandler(db, logger)

	server.POST("subscription", subController.PostSubscription)
	server.GET("subscription/users", subController.GetSubscriptionListByUserID)
	server.PUT("subscription", subController.UpdateSubscription)
	server.DELETE("subscription", subController.DeleteSubscription)

	server.GET("subscription/total-price", subController.GetTotalPeriodCostByDatesAndServiceName)

	server.GET("/swagger/*", echoSwagger.WrapHandler)

	return &Server{
		logger:  logger,
		server:  server,
		storage: db,
	}
}
func (s Server) Run(serverPort int) {
	s.logger.Info("Server is running on: localhost", "Port", serverPort)

	if err := s.server.Start(fmt.Sprintf("localhost:%d", serverPort)); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server starting error: %v", slog.Any("error_details", err))
		}
	}
}

func (s Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping DB Connection")

	s.storage.Close()

	s.logger.Info("Stopping server...")
	err := s.server.Shutdown(ctx)

	if err != nil {
		s.logger.Error("Error: ", slog.Any("error_details", err))

		return fmt.Errorf("error while stopping Server Request %w", err)
	}

	return nil
}
