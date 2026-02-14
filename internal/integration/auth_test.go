package integration

import (
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestAuth_RegisterAndLogin(t *testing.T) {
	cleanDB(t)

	// Register
	user := registerUser(t, "Alice", "alice@test.com", "secret123")
	if user.Name != "Alice" || user.Email != "alice@test.com" {
		t.Fatalf("unexpected user: %+v", user)
	}

	// Access protected route with token
	doGet(t, "/auth/me", user.AccessToken, http.StatusOK)

	// Login with same credentials
	resp := doJSON(t, http.MethodPost, "/auth/login",
		domain.LoginInput{Email: "alice@test.com", Password: "secret123"},
		"", http.StatusOK)

	var loginResult domain.AuthResponse
	decodeJSON(t, resp, &loginResult)

	if loginResult.Tokens.AccessToken == "" {
		t.Fatal("login should return access token")
	}

	// Access protected route with login token
	doGet(t, "/auth/me", loginResult.Tokens.AccessToken, http.StatusOK)
}

func TestAuth_RegisterDuplicateEmail(t *testing.T) {
	cleanDB(t)

	registerUser(t, "Alice", "dup@test.com", "secret123")

	// Attempt duplicate
	doJSON(t, http.MethodPost, "/auth/register",
		domain.RegisterInput{Name: "Bob", Email: "dup@test.com", Password: "secret456"},
		"", http.StatusConflict)
}

func TestAuth_LoginWrongPassword(t *testing.T) {
	cleanDB(t)

	registerUser(t, "Alice", "alice@test.com", "secret123")

	doJSON(t, http.MethodPost, "/auth/login",
		domain.LoginInput{Email: "alice@test.com", Password: "wrongpassword"},
		"", http.StatusUnauthorized)
}

func TestAuth_ProtectedRouteWithoutToken(t *testing.T) {
	doGet(t, "/auth/me", "", http.StatusUnauthorized)
}

func TestAuth_RefreshTokenFlow(t *testing.T) {
	cleanDB(t)

	// Register to get initial tokens
	resp := doJSON(t, http.MethodPost, "/auth/register",
		domain.RegisterInput{Name: "Alice", Email: "alice@test.com", Password: "secret123"},
		"", http.StatusCreated)

	var regResult domain.AuthResponse
	decodeJSON(t, resp, &regResult)

	// Use refresh token to get new access token
	resp = doJSON(t, http.MethodPost, "/auth/refresh",
		map[string]string{"refresh_token": regResult.Tokens.RefreshToken},
		"", http.StatusOK)

	var refreshResult domain.AuthResponse
	decodeJSON(t, resp, &refreshResult)

	if refreshResult.Tokens.AccessToken == "" {
		t.Fatal("refresh should return new access token")
	}

	// New access token should work
	doGet(t, "/auth/me", refreshResult.Tokens.AccessToken, http.StatusOK)
}
