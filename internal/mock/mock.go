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

// HouseholdRepository

type HouseholdRepository struct {
	CreateFn             func(ctx context.Context, household *domain.Household, adminUserID string) error
	FindByIDFn           func(ctx context.Context, id string) (*domain.Household, error)
	FindByInviteCodeFn   func(ctx context.Context, code string) (*domain.Household, error)
	ListByUserFn         func(ctx context.Context, userID string) ([]domain.Household, error)
	UpdateFn             func(ctx context.Context, household *domain.Household) error
	DeleteFn             func(ctx context.Context, id string) error
	AddMemberFn          func(ctx context.Context, householdID, userID, role string) error
	RemoveMemberFn       func(ctx context.Context, householdID, userID string) error
	UpdateMemberSalaryFn func(ctx context.Context, householdID, userID string, salaryCents int64) error
	ListMembersFn        func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error)
	GetMemberFn          func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error)
}

func (m *HouseholdRepository) Create(ctx context.Context, h *domain.Household, adminUserID string) error {
	return m.CreateFn(ctx, h, adminUserID)
}
func (m *HouseholdRepository) FindByID(ctx context.Context, id string) (*domain.Household, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *HouseholdRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Household, error) {
	return m.FindByInviteCodeFn(ctx, code)
}
func (m *HouseholdRepository) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	return m.ListByUserFn(ctx, userID)
}
func (m *HouseholdRepository) Update(ctx context.Context, h *domain.Household) error {
	return m.UpdateFn(ctx, h)
}
func (m *HouseholdRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}
func (m *HouseholdRepository) AddMember(ctx context.Context, householdID, userID, role string) error {
	return m.AddMemberFn(ctx, householdID, userID, role)
}
func (m *HouseholdRepository) RemoveMember(ctx context.Context, householdID, userID string) error {
	return m.RemoveMemberFn(ctx, householdID, userID)
}
func (m *HouseholdRepository) UpdateMemberSalary(ctx context.Context, householdID, userID string, salaryCents int64) error {
	return m.UpdateMemberSalaryFn(ctx, householdID, userID, salaryCents)
}
func (m *HouseholdRepository) ListMembers(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
	return m.ListMembersFn(ctx, householdID)
}
func (m *HouseholdRepository) GetMember(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
	return m.GetMemberFn(ctx, householdID, userID)
}

// HouseholdService

type HouseholdService struct {
	CreateFn             func(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error)
	GetByIDFn            func(ctx context.Context, id, userID string) (*domain.Household, error)
	ListByUserFn         func(ctx context.Context, userID string) ([]domain.Household, error)
	UpdateFn             func(ctx context.Context, id string, input domain.UpdateHouseholdInput, userID string) (*domain.Household, error)
	DeleteFn             func(ctx context.Context, id, userID string) error
	JoinFn               func(ctx context.Context, inviteCode, userID string) (*domain.Household, error)
	ListMembersFn        func(ctx context.Context, householdID, userID string) ([]domain.HouseholdMember, error)
	UpdateMemberSalaryFn func(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error
	RemoveMemberFn       func(ctx context.Context, householdID, memberID, userID string) error
}

func (m *HouseholdService) Create(ctx context.Context, input domain.CreateHouseholdInput, userID string) (*domain.Household, error) {
	return m.CreateFn(ctx, input, userID)
}
func (m *HouseholdService) GetByID(ctx context.Context, id, userID string) (*domain.Household, error) {
	return m.GetByIDFn(ctx, id, userID)
}
func (m *HouseholdService) ListByUser(ctx context.Context, userID string) ([]domain.Household, error) {
	return m.ListByUserFn(ctx, userID)
}
func (m *HouseholdService) Update(ctx context.Context, id string, input domain.UpdateHouseholdInput, userID string) (*domain.Household, error) {
	return m.UpdateFn(ctx, id, input, userID)
}
func (m *HouseholdService) Delete(ctx context.Context, id, userID string) error {
	return m.DeleteFn(ctx, id, userID)
}
func (m *HouseholdService) Join(ctx context.Context, inviteCode, userID string) (*domain.Household, error) {
	return m.JoinFn(ctx, inviteCode, userID)
}
func (m *HouseholdService) ListMembers(ctx context.Context, householdID, userID string) ([]domain.HouseholdMember, error) {
	return m.ListMembersFn(ctx, householdID, userID)
}
func (m *HouseholdService) UpdateMemberSalary(ctx context.Context, householdID, memberID string, salaryCents int64, userID string) error {
	return m.UpdateMemberSalaryFn(ctx, householdID, memberID, salaryCents, userID)
}
func (m *HouseholdService) RemoveMember(ctx context.Context, householdID, memberID, userID string) error {
	return m.RemoveMemberFn(ctx, householdID, memberID, userID)
}
