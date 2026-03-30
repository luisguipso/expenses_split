package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

func RegisterHealthRoutes(e *echo.Echo, health domain.HealthChecker) {
	e.GET("/health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := health.Ping(ctx); err != nil {
			slog.Error("handler: health check database ping failed",
				"error", err,
			)
			dbStatus = "error"
		}

		return c.JSON(http.StatusOK, domain.HealthResponse{
			Status:   "ok",
			Database: dbStatus,
		})
	})
}
