package domain

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

type TokenClaims struct {
	UserID string
	Email  string
}

type UserResponse struct {
	ID    interface{} `json:"id"`
	Name  string      `json:"name"`
	Email string      `json:"email"`
}

type AuthResponse struct {
	User   UserResponse `json:"user"`
	Tokens *TokenPair   `json:"tokens"`
}

type TokenResponse struct {
	Tokens *TokenPair `json:"tokens"`
}

type MeResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
