package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/lguilherme/contas/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo       domain.UserRepository
	tokens         domain.TokenService
	verifyRepo     domain.EmailVerificationRepository
	emailSvc       domain.EmailService
	codeExpiration time.Duration
}

func NewAuthService(
	userRepo domain.UserRepository,
	tokens domain.TokenService,
	verifyRepo domain.EmailVerificationRepository,
	emailSvc domain.EmailService,
	codeExpiration time.Duration,
) domain.AuthService {
	return &authService{
		userRepo:       userRepo,
		tokens:         tokens,
		verifyRepo:     verifyRepo,
		emailSvc:       emailSvc,
		codeExpiration: codeExpiration,
	}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Name:          input.Name,
		Email:         input.Email,
		PasswordHash:  string(hash),
		EmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return nil, domain.ErrEmailExists
		}
		return nil, err
	}

	code := generateCode()
	verification := &domain.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.codeExpiration),
	}

	if err := s.verifyRepo.Create(ctx, verification); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(user.Email, code); err != nil {
		return nil, fmt.Errorf("send verification email: %w", err)
	}

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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return nil, nil, domain.ErrEmailNotVerified
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

	if time.Now().After(verification.ExpiresAt) {
		return nil, nil, domain.ErrVerificationExpired
	}

	if verification.Code != input.Code {
		return nil, nil, domain.ErrInvalidVerificationCode
	}

	if err := s.verifyRepo.MarkUsed(ctx, verification.ID); err != nil {
		return nil, nil, err
	}

	if err := s.userRepo.VerifyEmail(ctx, verification.UserID); err != nil {
		return nil, nil, err
	}

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

	if user.EmailVerified {
		return domain.ErrAlreadyVerified
	}

	code := generateCode()
	verification := &domain.EmailVerification{
		UserID:    user.ID,
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(s.codeExpiration),
	}

	if err := s.verifyRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	if err := s.emailSvc.SendVerificationCode(user.Email, code); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}
