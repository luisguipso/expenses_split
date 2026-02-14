package mock

import (
	"context"

	"github.com/lguilherme/contas/internal/domain"
)

// UserRepository

type UserRepository struct {
	CreateFn      func(ctx context.Context, user *domain.User) error
	FindByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	FindByIDFn    func(ctx context.Context, id string) (*domain.User, error)
}

func (m *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFn(ctx, user)
}

func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.FindByIDFn(ctx, id)
}

// AuthService

type AuthService struct {
	RegisterFn     func(ctx context.Context, input domain.RegisterInput) (*domain.User, *domain.TokenPair, error)
	LoginFn        func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	RefreshTokenFn func(refreshToken string) (*domain.TokenPair, error)
}

func (m *AuthService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, *domain.TokenPair, error) {
	return m.RegisterFn(ctx, input)
}

func (m *AuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
	return m.LoginFn(ctx, input)
}

func (m *AuthService) RefreshToken(refreshToken string) (*domain.TokenPair, error) {
	return m.RefreshTokenFn(refreshToken)
}

// TokenService

type TokenService struct {
	GenerateFn func(userID, email string) (*domain.TokenPair, error)
	ValidateFn func(tokenStr string) (*domain.TokenClaims, error)
}

func (m *TokenService) Generate(userID, email string) (*domain.TokenPair, error) {
	return m.GenerateFn(userID, email)
}

func (m *TokenService) Validate(tokenStr string) (*domain.TokenClaims, error) {
	return m.ValidateFn(tokenStr)
}

// HealthChecker

type HealthChecker struct {
	PingFn func(ctx context.Context) error
}

func (m *HealthChecker) Ping(ctx context.Context) error {
	return m.PingFn(ctx)
}
