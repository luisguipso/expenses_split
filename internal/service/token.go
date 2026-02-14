package service

import (
	"fmt"
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
	accessTokenStr, accessExpiry, err := s.signToken(userID, email, s.accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshTokenStr, _, err := s.signToken(userID, email, s.refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}

func (s *JWTTokenService) Validate(tokenStr string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

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
