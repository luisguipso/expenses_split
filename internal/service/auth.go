package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lguilherme/contas/internal/domain"
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
	user := &domain.User{
		Name:          input.Name,
		Email:         input.Email,
		EmailVerified: false,
	}

	if err := user.SetPassword(input.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return nil, domain.ErrEmailExists
		}
		return nil, err
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)

	verification := domain.NewEmailVerification(user.ID, user.Email, s.codeExpiration)

	if err := s.verifyRepo.Create(ctx, verification); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(user.Email, verification.Code); err != nil {
		slog.Error("failed to send verification email", "error", err, "email", user.Email)
		return nil, fmt.Errorf("send verification email: %w", err)
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
		return nil, nil, err
	}

	if err := user.VerifyPassword(input.Password); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if err := user.RequireVerifiedEmail(); err != nil {
		slog.Warn("login attempt with unverified email", "email", input.Email)
		return nil, nil, err
	}

	tokens, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	if err := verification.MatchesCode(input.Code); err != nil {
		return nil, nil, err
	}

	if err := s.verifyRepo.MarkUsed(ctx, verification.ID); err != nil {
		return nil, nil, err
	}

	if err := s.userRepo.VerifyEmail(ctx, verification.UserID); err != nil {
		return nil, nil, err
	}

	slog.Info("email verified", "email", input.Email, "user_id", verification.UserID)

	user, err := s.userRepo.FindByID(ctx, verification.UserID)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *authService) ResendCode(ctx context.Context, input domain.ResendCodeInput) error {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		return err
	}

	if err := user.RequireVerifiedEmail(); err == nil {
		return domain.ErrAlreadyVerified
	}

	verification := domain.NewEmailVerification(user.ID, user.Email, s.codeExpiration)

	if err := s.verifyRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(user.Email, verification.Code); err != nil {
		slog.Error("failed to resend verification email", "error", err, "email", user.Email)
		return fmt.Errorf("send verification email: %w", err)
	}

	slog.Info("verification code resent", "email", user.Email)

	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			slog.Info("forgot password for unknown email", "email", input.Email)
			return nil
		}
		return err
	}

	reset, err := domain.NewPasswordReset(user.ID, user.Email, s.resetExpiration)
	if err != nil {
		return err
	}

	if err := s.resetRepo.Create(ctx, reset); err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}

	resetLink := fmt.Sprintf("%s/password-recover?token=%s", s.frontendURL, reset.Token)
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
		return err
	}

	if reset.IsExpired() {
		return domain.ErrResetTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, reset.UserID)
	if err != nil {
		return err
	}

	if user.IsSamePassword(input.NewPassword) {
		return domain.ErrPasswordSameAsOld
	}

	if err := user.SetPassword(input.NewPassword); err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, user.PasswordHash); err != nil {
		return err
	}

	if err := s.resetRepo.MarkUsed(ctx, reset.ID); err != nil {
		return err
	}

	slog.Info("password reset successful", "user_id", user.ID, "email", user.Email)
	return nil
}
