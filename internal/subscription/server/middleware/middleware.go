package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func LogRequest(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echo echo.Context) error {
			start := time.Now()

			err := next(echo)

			stop := time.Now()

			logger.Info("Request: ",
				"Method", echo.Request().Method,
				"URL", echo.Request().URL,
				"Time", stop.Sub(start),
				"Http Code", echo.Response().Status)

			return err
		}
	}
}
