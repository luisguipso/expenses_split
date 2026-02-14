package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}

type TokenService interface {
	Generate(userID, email string) (*TokenPair, error)
	Validate(tokenStr string) (*TokenClaims, error)
}

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*User, *TokenPair, error)
	Login(ctx context.Context, input LoginInput) (*User, *TokenPair, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
