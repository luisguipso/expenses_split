package handler

import (
	"errors"
	"log/slog"
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

	user, err := h.auth.Register(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return echo.NewHTTPError(http.StatusConflict, "email already taken")
		}
		slog.Error("register failed", "error", err, "email", input.Email)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register")
	}

	return c.JSON(http.StatusCreated, domain.RegisterResponse{
		User:          toUserResponse(user),
		EmailVerified: user.EmailVerified,
		Message:       "verification code sent to your email",
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
		if errors.Is(err, domain.ErrEmailNotVerified) {
			return echo.NewHTTPError(http.StatusForbidden, "email_not_verified")
		}
		slog.Error("login failed", "error", err, "email", input.Email)
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

func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	var input domain.VerifyEmailInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" || input.Code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and code are required")
	}

	user, tokens, err := h.auth.VerifyEmail(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVerificationCode) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid verification code")
		}
		if errors.Is(err, domain.ErrVerificationExpired) {
			return echo.NewHTTPError(http.StatusBadRequest, "verification code expired")
		}
		slog.Error("verify email failed", "error", err, "email", input.Email)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify email")
	}

	return c.JSON(http.StatusOK, domain.AuthResponse{
		User:   toUserResponse(user),
		Tokens: tokens,
	})
}

func (h *AuthHandler) ResendCode(c echo.Context) error {
	var input domain.ResendCodeInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	if err := h.auth.ResendCode(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		if errors.Is(err, domain.ErrAlreadyVerified) {
			return echo.NewHTTPError(http.StatusConflict, "email already verified")
		}
		slog.Error("resend code failed", "error", err, "email", input.Email)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to resend code")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "verification code sent"})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var input domain.ForgotPasswordInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	if !isValidEmail(input.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email format")
	}

	if err := h.auth.ForgotPassword(c.Request().Context(), input); err != nil {
		slog.Error("forgot password failed", "error", err, "email", input.Email)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to process request")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "if the email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var input domain.ResetPasswordInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	if len(input.NewPassword) < 6 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 6 characters")
	}

	if err := h.auth.ResetPassword(c.Request().Context(), input); err != nil {
		if errors.Is(err, domain.ErrResetTokenInvalid) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid or already used reset token")
		}
		if errors.Is(err, domain.ErrResetTokenExpired) {
			return echo.NewHTTPError(http.StatusBadRequest, "reset token has expired")
		}
		if errors.Is(err, domain.ErrPasswordSameAsOld) {
			return echo.NewHTTPError(http.StatusBadRequest, "new password must differ from current password")
		}
		slog.Error("reset password failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reset password")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password reset successfully"})
}

func RegisterAuthRoutes(e *echo.Echo, h *AuthHandler, authMiddleware echo.MiddlewareFunc) {
	auth := e.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/verify-email", h.VerifyEmail)
	auth.POST("/resend-code", h.ResendCode)
	auth.POST("/forgot-password", h.ForgotPassword)
	auth.POST("/reset-password", h.ResetPassword)
	auth.GET("/me", h.Me, authMiddleware)
}

func toUserResponse(u *domain.User) domain.UserResponse {
	return domain.UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}
