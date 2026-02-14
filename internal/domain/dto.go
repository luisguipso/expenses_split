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

// Household DTOs

type CreateHouseholdInput struct {
	Name string `json:"name"`
}

type UpdateHouseholdInput struct {
	Name string `json:"name"`
}

type JoinHouseholdInput struct {
	InviteCode string `json:"invite_code"`
}

type UpdateSalaryInput struct {
	SalaryCents int64 `json:"salary_cents"`
}

type HouseholdResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type MemberResponse struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	UserEmail   string `json:"user_email"`
	SalaryCents int64  `json:"salary_cents"`
	Role        string `json:"role"`
	JoinedAt    string `json:"joined_at"`
}
