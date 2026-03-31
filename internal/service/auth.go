package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/lguilherme/contas/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo        domain.UserRepository
	tokens          domain.TokenService
	verifyRepo      domain.EmailVerificationRepository
	resetRepo       domain.PasswordResetRepository
	emailSvc        domain.EmailService
	codeExpiration  time.Duration
	resetExpiration time.Duration
	frontendURL     string
}

func NewAuthService(
	userRepo domain.UserRepository,
	tokens domain.TokenService,
	verifyRepo domain.EmailVerificationRepository,
	emailSvc domain.EmailService,
	codeExpiration time.Duration,
	resetRepo domain.PasswordResetRepository,
	resetExpiration time.Duration,
	frontendURL string,
) domain.AuthService {
	return &authService{
		userRepo:        userRepo,
		tokens:          tokens,
		verifyRepo:      verifyRepo,
		resetRepo:       resetRepo,
		emailSvc:        emailSvc,
		codeExpiration:  codeExpiration,
		resetExpiration: resetExpiration,
		frontendURL:     frontendURL,
	}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:          input.Name,
		Email:         input.Email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return nil, domain.ErrEmailExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)

	if err := s.createAndSendVerificationCode(ctx, user.ID, user.Email); err != nil {
		return nil, err
	}

	slog.Info("verification code sent", "email", user.Email)

	return user, nil
}

func (s *authService) Login(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, nil, domain.ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.EmailVerified {
		slog.Warn("login attempt with unverified email", "email", input.Email)
		return nil, nil, domain.ErrEmailNotVerified
	}

	tokens, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	return user, tokens, nil
}

func (s *authService) RefreshToken(refreshTokenStr string) (*domain.TokenPair, error) {
	claims, err := s.tokens.Validate(refreshTokenStr)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	return s.tokens.Generate(claims.UserID, claims.Email)
}

func (s *authService) VerifyEmail(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error) {
	verification, err := s.verifyRepo.FindLatestByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVerificationCode) {
			return nil, nil, domain.ErrInvalidVerificationCode
		}
		return nil, nil, fmt.Errorf("find latest verification: %w", err)
	}

	if time.Now().After(verification.ExpiresAt) {
		return nil, nil, domain.ErrVerificationExpired
	}

	if verification.Code != input.Code {
		return nil, nil, domain.ErrInvalidVerificationCode
	}

	if err := s.verifyRepo.MarkUsed(ctx, verification.ID); err != nil {
		return nil, nil, fmt.Errorf("mark verification used: %w", err)
	}

	if err := s.userRepo.VerifyEmail(ctx, verification.UserID); err != nil {
		return nil, nil, fmt.Errorf("verify user email: %w", err)
	}

	slog.Info("email verified", "email", input.Email, "user_id", verification.UserID)

	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("find user by id: %w", err)
	}

	tokens, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("generate tokens: %w", err)
	}

	return user, tokens, nil
}

func (s *authService) ResendCode(ctx context.Context, input domain.ResendCodeInput) error {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("find user by email: %w", err)
	}

	if user.EmailVerified {
		return domain.ErrAlreadyVerified
	}

	if err := s.createAndSendVerificationCode(ctx, user.ID, user.Email); err != nil {
		return err
	}

	slog.Info("verification code resent", "email", user.Email)

	return nil
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *authService) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Silent success to avoid leaking whether the email exists
			slog.Info("forgot password for unknown email", "email", input.Email)
			return nil
		}
		return fmt.Errorf("find user by email: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	reset := &domain.PasswordReset{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(s.resetExpiration),
	}

	if err := s.resetRepo.Create(ctx, reset); err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}

	resetLink := fmt.Sprintf("%s/password-recover?token=%s", s.frontendURL, token)
	if err := s.emailSvc.SendPasswordResetLink(user.Email, resetLink); err != nil {
		slog.Error("failed to send password reset email", "error", err, "email", user.Email)
		return fmt.Errorf("send password reset email: %w", err)
	}

	slog.Info("password reset email sent", "email", user.Email)
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error {
	reset, err := s.resetRepo.FindByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, domain.ErrResetTokenInvalid) {
			return domain.ErrResetTokenInvalid
		}
		return fmt.Errorf("find reset by token: %w", err)
	}

	if time.Now().After(reset.ExpiresAt) {
		return domain.ErrResetTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, reset.UserID)
	if err != nil {
		return fmt.Errorf("find user by id: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.NewPassword)); err == nil {
		return domain.ErrPasswordSameAsOld
	}

	passwordHash, err := hashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	if err := s.resetRepo.MarkUsed(ctx, reset.ID); err != nil {
		return fmt.Errorf("mark reset token used: %w", err)
	}

	slog.Info("password reset successful", "user_id", user.ID, "email", user.Email)
	return nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *authService) createAndSendVerificationCode(ctx context.Context, userID, email string) error {
	code := generateCode()
	verification := &domain.EmailVerification{
		UserID:    userID,
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.codeExpiration),
	}

	if err := s.verifyRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(email, code); err != nil {
		slog.Error("failed to send verification email", "error", err, "email", email)
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}
