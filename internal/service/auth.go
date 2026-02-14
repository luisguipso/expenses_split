package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lguilherme/contas/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo domain.UserRepository
	tokens   domain.TokenService
}

func NewAuthService(userRepo domain.UserRepository, tokens domain.TokenService) domain.AuthService {
	return &authService{
		userRepo: userRepo,
		tokens:   tokens,
	}
}

func (s *authService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, *domain.TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailExists) {
			return nil, nil, domain.ErrEmailExists
		}
		return nil, nil, err
	}

	tokens, err := s.tokens.Generate(user.ID, user.Email)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
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
