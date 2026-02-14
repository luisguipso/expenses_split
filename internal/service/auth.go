package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lguilherme/contas/internal/model"
	"github.com/lguilherme/contas/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyTaken  = errors.New("email already taken")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, *TokenPair, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return nil, nil, ErrEmailAlreadyTaken
		}
		return nil, nil, err
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*model.User, *TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

func (s *AuthService) RefreshToken(refreshTokenStr string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Create a temporary user to generate new tokens
	user := &model.User{
		Email: claims.Email,
	}
	// Parse UUID back
	if err := user.ID.Scan(claims.UserID); err != nil {
		return nil, ErrInvalidToken
	}

	return s.generateTokenPair(user)
}

func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
	userID := fmt.Sprintf("%x-%x-%x-%x-%x",
		user.ID.Bytes[0:4], user.ID.Bytes[4:6], user.ID.Bytes[6:8],
		user.ID.Bytes[8:10], user.ID.Bytes[10:16])

	accessExpiry := time.Now().Add(15 * time.Minute)
	accessClaims := &Claims{
		UserID: userID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID: userID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		ExpiresAt:    accessExpiry.Unix(),
	}, nil
}
