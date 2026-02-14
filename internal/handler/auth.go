package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var input service.RegisterInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Name == "" || input.Email == "" || input.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, email, and password are required")
	}

	if len(input.Password) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 6 characters")
	}

	user, tokens, err := h.authService.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyTaken) {
			return echo.NewHTTPError(http.StatusConflict, "email already taken")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var input service.LoginInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" || input.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	user, tokens, err := h.authService.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to login")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if body.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	tokens, err := h.authService.RefreshToken(body.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tokens": tokens,
	})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID := c.Get("user_id").(string)
	email := c.Get("user_email").(string)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"email":   email,
	})
}

func RegisterAuthRoutes(e *echo.Echo, authHandler *AuthHandler, authMiddleware echo.MiddlewareFunc) {
	auth := e.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)
	auth.GET("/me", authHandler.Me, authMiddleware)
}
