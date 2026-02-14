package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/lguilherme/contas/internal/domain"
)

type AuthHandler struct {
	auth domain.AuthService
}

func NewAuthHandler(auth domain.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var input domain.RegisterInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Name == "" || input.Email == "" || input.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, email, and password are required")
	}

	if err := validateMaxLen("name", input.Name, 255); err != nil {
		return err
	}
	if err := validateMaxLen("email", input.Email, 255); err != nil {
		return err
	}
	if !isValidEmail(input.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email format")
	}

	if len(input.Password) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 6 characters")
	}

	user, tokens, err := h.auth.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return echo.NewHTTPError(http.StatusConflict, "email already taken")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register")
	}

	return c.JSON(http.StatusCreated, domain.AuthResponse{
		User:   toUserResponse(user),
		Tokens: tokens,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var input domain.LoginInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" || input.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	user, tokens, err := h.auth.Login(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to login")
	}

	return c.JSON(http.StatusOK, domain.AuthResponse{
		User:   toUserResponse(user),
		Tokens: tokens,
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

	tokens, err := h.auth.RefreshToken(body.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	return c.JSON(http.StatusOK, domain.TokenResponse{Tokens: tokens})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	email, _ := c.Get("user_email").(string)

	return c.JSON(http.StatusOK, domain.MeResponse{
		UserID: userID,
		Email:  email,
	})
}

func RegisterAuthRoutes(e *echo.Echo, h *AuthHandler, authMiddleware echo.MiddlewareFunc) {
	auth := e.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.GET("/me", h.Me, authMiddleware)
}

func toUserResponse(u *domain.User) domain.UserResponse {
	return domain.UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}
