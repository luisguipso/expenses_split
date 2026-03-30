package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

func JWTAuth(tokens domain.TokenService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			slog.Debug("middleware: validating auth token")

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				slog.Warn("middleware: missing authorization header")
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				slog.Warn("middleware: invalid authorization format")
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			claims, err := tokens.Validate(parts[1])
			if err != nil {
				slog.Error("middleware: token validation failed", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			slog.Debug("middleware: auth successful", "user_id", claims.UserID, "email", claims.Email)
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			return next(c)
		}
	}
}
