package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func RegisterHealthRoutes(e *echo.Echo, db *pgxpool.Pool) {
	e.GET("/health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := db.Ping(ctx); err != nil {
			dbStatus = "error"
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status":   "ok",
			"database": dbStatus,
		})
	})
}
