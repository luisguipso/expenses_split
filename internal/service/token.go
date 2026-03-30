package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lguilherme/contas/internal/domain"
)

type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type JWTTokenService struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTTokenService(secret string) *JWTTokenService {
	return &JWTTokenService{
		secret:        []byte(secret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (s *JWTTokenService) Generate(userID, email string) (*domain.TokenPair, error) {
	slog.Info("service: generating token pair",
		"user_id", userID,
		"email", email,
	)

	accessTokenStr, accessExpiry, err := s.signToken(userID, email, s.accessExpiry)
	if err != nil {
		slog.Error("service: failed to sign access token",
			"error", err,
			"user_id", userID,
		)
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshTokenStr, _, err := s.signToken(userID, email, s.refreshExpiry)
	if err != nil {
		slog.Error("service: failed to sign refresh token",
			"error", err,
			"user_id", userID,
		)
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	slog.Info("service: token pair generated",
		"user_id", userID,
	)
	return &domain.TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

func (s *JWTTokenService) Validate(tokenStr string) (*domain.TokenClaims, error) {
	slog.Debug("service: validating token")

	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		slog.Warn("service: token validation failed",
			"error", err,
		)
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		slog.Warn("service: token claims invalid")
		return nil, domain.ErrInvalidToken
	}

	slog.Debug("service: token validated",
		"user_id", claims.UserID,
	)
	return &domain.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (s *JWTTokenService) signToken(userID, email string, expiry time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(expiry)
	claims := &jwtClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.secret)
	return tokenStr, expiresAt, err
}
