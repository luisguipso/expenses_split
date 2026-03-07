package mock

import (
	"context"

	"github.com/lguilherme/contas/internal/domain"
)

// UserRepository

type UserRepository struct {
	CreateFn         func(ctx context.Context, user *domain.User) error
	FindByEmailFn    func(ctx context.Context, email string) (*domain.User, error)
	FindByIDFn       func(ctx context.Context, id string) (*domain.User, error)
	VerifyEmailFn    func(ctx context.Context, userID string) error
	UpdatePasswordFn func(ctx context.Context, userID, passwordHash string) error
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

func (m *UserRepository) VerifyEmail(ctx context.Context, userID string) error {
	return m.VerifyEmailFn(ctx, userID)
}

func (m *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return m.UpdatePasswordFn(ctx, userID, passwordHash)
}

// AuthService

type AuthService struct {
	RegisterFn       func(ctx context.Context, input domain.RegisterInput) (*domain.User, error)
	LoginFn          func(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error)
	RefreshTokenFn   func(refreshToken string) (*domain.TokenPair, error)
	VerifyEmailFn    func(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error)
	ResendCodeFn     func(ctx context.Context, input domain.ResendCodeInput) error
	ForgotPasswordFn func(ctx context.Context, input domain.ForgotPasswordInput) error
	ResetPasswordFn  func(ctx context.Context, input domain.ResetPasswordInput) error
}

func (m *AuthService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	return m.RegisterFn(ctx, input)
}

func (m *AuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.User, *domain.TokenPair, error) {
	return m.LoginFn(ctx, input)
}

func (m *AuthService) RefreshToken(refreshToken string) (*domain.TokenPair, error) {
	return m.RefreshTokenFn(refreshToken)
}

func (m *AuthService) VerifyEmail(ctx context.Context, input domain.VerifyEmailInput) (*domain.User, *domain.TokenPair, error) {
	return m.VerifyEmailFn(ctx, input)
}

func (m *AuthService) ResendCode(ctx context.Context, input domain.ResendCodeInput) error {
	return m.ResendCodeFn(ctx, input)
}

func (m *AuthService) ForgotPassword(ctx context.Context, input domain.ForgotPasswordInput) error {
	return m.ForgotPasswordFn(ctx, input)
}

func (m *AuthService) ResetPassword(ctx context.Context, input domain.ResetPasswordInput) error {
	return m.ResetPasswordFn(ctx, input)
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

// CategoryRepository

type CategoryRepository struct {
	CreateFn          func(ctx context.Context, category *domain.Category) error
	FindByIDFn        func(ctx context.Context, id string) (*domain.Category, error)
	ListByHouseholdFn func(ctx context.Context, householdID string) ([]domain.Category, error)
	UpdateFn          func(ctx context.Context, category *domain.Category) error
	DeleteFn          func(ctx context.Context, id string) error
}

func (m *CategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	return m.CreateFn(ctx, c)
}
func (m *CategoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *CategoryRepository) ListByHousehold(ctx context.Context, householdID string) ([]domain.Category, error) {
	return m.ListByHouseholdFn(ctx, householdID)
}
func (m *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	return m.UpdateFn(ctx, c)
}
func (m *CategoryRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

// CategoryService

type CategoryService struct {
	CreateFn func(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error)
	ListFn   func(ctx context.Context, householdID, userID string) ([]domain.Category, error)
	UpdateFn func(ctx context.Context, id string, input domain.UpdateCategoryInput, userID string) (*domain.Category, error)
	DeleteFn func(ctx context.Context, id, userID string) error
}

func (m *CategoryService) Create(ctx context.Context, input domain.CreateCategoryInput, householdID, userID string) (*domain.Category, error) {
	return m.CreateFn(ctx, input, householdID, userID)
}
func (m *CategoryService) List(ctx context.Context, householdID, userID string) ([]domain.Category, error) {
	return m.ListFn(ctx, householdID, userID)
}
func (m *CategoryService) Update(ctx context.Context, id string, input domain.UpdateCategoryInput, userID string) (*domain.Category, error) {
	return m.UpdateFn(ctx, id, input, userID)
}
func (m *CategoryService) Delete(ctx context.Context, id, userID string) error {
	return m.DeleteFn(ctx, id, userID)
}

// FixedBillRepository

type FixedBillRepository struct {
	CreateFn          func(ctx context.Context, bill *domain.FixedBill) error
	FindByIDFn        func(ctx context.Context, id string) (*domain.FixedBill, error)
	ListByHouseholdFn func(ctx context.Context, householdID string) ([]domain.FixedBill, error)
	UpdateFn          func(ctx context.Context, bill *domain.FixedBill) error
	DeleteFn          func(ctx context.Context, id string) error
}

func (m *FixedBillRepository) Create(ctx context.Context, b *domain.FixedBill) error {
	return m.CreateFn(ctx, b)
}
func (m *FixedBillRepository) FindByID(ctx context.Context, id string) (*domain.FixedBill, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *FixedBillRepository) ListByHousehold(ctx context.Context, householdID string) ([]domain.FixedBill, error) {
	return m.ListByHouseholdFn(ctx, householdID)
}
func (m *FixedBillRepository) Update(ctx context.Context, b *domain.FixedBill) error {
	return m.UpdateFn(ctx, b)
}
func (m *FixedBillRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

// FixedBillSnapshotRepository

type FixedBillSnapshotRepository struct {
	CreateFn      func(ctx context.Context, snapshot *domain.FixedBillSnapshot) error
	FindByMonthFn func(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error)
	FindByIDFn    func(ctx context.Context, id string) (*domain.FixedBillSnapshot, error)
	UpdateFn      func(ctx context.Context, snapshot *domain.FixedBillSnapshot) error
	DeleteFn      func(ctx context.Context, id string) error
}

func (m *FixedBillSnapshotRepository) Create(ctx context.Context, s *domain.FixedBillSnapshot) error {
	return m.CreateFn(ctx, s)
}
func (m *FixedBillSnapshotRepository) FindByMonth(ctx context.Context, householdID string, year, month int) ([]domain.FixedBillSnapshot, error) {
	return m.FindByMonthFn(ctx, householdID, year, month)
}
func (m *FixedBillSnapshotRepository) FindByID(ctx context.Context, id string) (*domain.FixedBillSnapshot, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *FixedBillSnapshotRepository) Update(ctx context.Context, s *domain.FixedBillSnapshot) error {
	return m.UpdateFn(ctx, s)
}
func (m *FixedBillSnapshotRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

// FixedBillService

type FixedBillService struct {
	CreateFn func(ctx context.Context, input domain.CreateFixedBillInput, householdID, userID string) (*domain.FixedBill, error)
	ListFn   func(ctx context.Context, householdID, userID string) ([]domain.FixedBill, error)
	UpdateFn func(ctx context.Context, id string, input domain.UpdateFixedBillInput, userID string) (*domain.FixedBill, error)
	DeleteFn func(ctx context.Context, id, userID string) error
}

func (m *FixedBillService) Create(ctx context.Context, input domain.CreateFixedBillInput, householdID, userID string) (*domain.FixedBill, error) {
	return m.CreateFn(ctx, input, householdID, userID)
}
func (m *FixedBillService) List(ctx context.Context, householdID, userID string) ([]domain.FixedBill, error) {
	return m.ListFn(ctx, householdID, userID)
}
func (m *FixedBillService) Update(ctx context.Context, id string, input domain.UpdateFixedBillInput, userID string) (*domain.FixedBill, error) {
	return m.UpdateFn(ctx, id, input, userID)
}
func (m *FixedBillService) Delete(ctx context.Context, id, userID string) error {
	return m.DeleteFn(ctx, id, userID)
}

// FixedBillSnapshotService

type FixedBillSnapshotService struct {
	UpdateFn func(ctx context.Context, id string, input domain.UpdateFixedBillSnapshotInput, userID string) (*domain.FixedBillSnapshot, error)
}

func (m *FixedBillSnapshotService) Update(ctx context.Context, id string, input domain.UpdateFixedBillSnapshotInput, userID string) (*domain.FixedBillSnapshot, error) {
	return m.UpdateFn(ctx, id, input, userID)
}

// ExpenseRepository

type ExpenseRepository struct {
	CreateFn          func(ctx context.Context, expense *domain.Expense) error
	FindByIDFn        func(ctx context.Context, id string) (*domain.Expense, error)
	ListByHouseholdFn func(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error)
	UpdateFn          func(ctx context.Context, expense *domain.Expense) error
	DeleteFn          func(ctx context.Context, id string) error
}

func (m *ExpenseRepository) Create(ctx context.Context, e *domain.Expense) error {
	return m.CreateFn(ctx, e)
}
func (m *ExpenseRepository) FindByID(ctx context.Context, id string) (*domain.Expense, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *ExpenseRepository) ListByHousehold(ctx context.Context, householdID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
	return m.ListByHouseholdFn(ctx, householdID, filter)
}
func (m *ExpenseRepository) Update(ctx context.Context, e *domain.Expense) error {
	return m.UpdateFn(ctx, e)
}
func (m *ExpenseRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

// ExpenseService

type ExpenseService struct {
	CreateFn func(ctx context.Context, input domain.CreateExpenseInput, householdID, userID string) (*domain.Expense, error)
	ListFn   func(ctx context.Context, householdID, userID string, filter domain.ExpenseFilter) ([]domain.Expense, error)
	UpdateFn func(ctx context.Context, id string, input domain.UpdateExpenseInput, userID string) (*domain.Expense, error)
	DeleteFn func(ctx context.Context, id, userID string) error
}

func (m *ExpenseService) Create(ctx context.Context, input domain.CreateExpenseInput, householdID, userID string) (*domain.Expense, error) {
	return m.CreateFn(ctx, input, householdID, userID)
}
func (m *ExpenseService) List(ctx context.Context, householdID, userID string, filter domain.ExpenseFilter) ([]domain.Expense, error) {
	return m.ListFn(ctx, householdID, userID, filter)
}
func (m *ExpenseService) Update(ctx context.Context, id string, input domain.UpdateExpenseInput, userID string) (*domain.Expense, error) {
	return m.UpdateFn(ctx, id, input, userID)
}
func (m *ExpenseService) Delete(ctx context.Context, id, userID string) error {
	return m.DeleteFn(ctx, id, userID)
}

// SummaryRepository

type SummaryRepository struct {
	UpsertFn      func(ctx context.Context, summary *domain.MonthlySummary) error
	FindByMonthFn func(ctx context.Context, householdID string, year, month int) (*domain.MonthlySummary, error)
}

func (m *SummaryRepository) Upsert(ctx context.Context, summary *domain.MonthlySummary) error {
	return m.UpsertFn(ctx, summary)
}
func (m *SummaryRepository) FindByMonth(ctx context.Context, householdID string, year, month int) (*domain.MonthlySummary, error) {
	return m.FindByMonthFn(ctx, householdID, year, month)
}

// SummaryService

type SummaryService struct {
	GenerateFn      func(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error)
	GetDashboardFn  func(ctx context.Context, householdID, userID string) (*domain.DashboardResponse, error)
	GetUserDetailFn func(ctx context.Context, householdID string, year, month int, targetUserID, requestingUserID string) (*domain.SummaryDetailResponse, error)
}

func (m *SummaryService) Generate(ctx context.Context, householdID string, year, month int, userID string) (*domain.SummaryResponse, error) {
	return m.GenerateFn(ctx, householdID, year, month, userID)
}
func (m *SummaryService) GetDashboard(ctx context.Context, householdID, userID string) (*domain.DashboardResponse, error) {
	return m.GetDashboardFn(ctx, householdID, userID)
}
func (m *SummaryService) GetUserDetail(ctx context.Context, householdID string, year, month int, targetUserID, requestingUserID string) (*domain.SummaryDetailResponse, error) {
	return m.GetUserDetailFn(ctx, householdID, year, month, targetUserID, requestingUserID)
}

// EmailVerificationRepository

type EmailVerificationRepository struct {
	CreateFn          func(ctx context.Context, verification *domain.EmailVerification) error
	FindLatestByEmailFn func(ctx context.Context, email string) (*domain.EmailVerification, error)
	MarkUsedFn        func(ctx context.Context, id string) error
}

func (m *EmailVerificationRepository) Create(ctx context.Context, v *domain.EmailVerification) error {
	return m.CreateFn(ctx, v)
}

func (m *EmailVerificationRepository) FindLatestByEmail(ctx context.Context, email string) (*domain.EmailVerification, error) {
	return m.FindLatestByEmailFn(ctx, email)
}

func (m *EmailVerificationRepository) MarkUsed(ctx context.Context, id string) error {
	return m.MarkUsedFn(ctx, id)
}

// EmailService

type EmailServiceMock struct {
	SendVerificationCodeFn   func(to, code string) error
	SendPasswordResetLinkFn  func(to, resetLink string) error
}

func (m *EmailServiceMock) SendVerificationCode(to, code string) error {
	return m.SendVerificationCodeFn(to, code)
}

func (m *EmailServiceMock) SendPasswordResetLink(to, resetLink string) error {
	return m.SendPasswordResetLinkFn(to, resetLink)
}

// PasswordResetRepository

type PasswordResetRepository struct {
	CreateFn      func(ctx context.Context, reset *domain.PasswordReset) error
	FindByTokenFn func(ctx context.Context, token string) (*domain.PasswordReset, error)
	MarkUsedFn    func(ctx context.Context, id string) error
}

func (m *PasswordResetRepository) Create(ctx context.Context, pr *domain.PasswordReset) error {
	return m.CreateFn(ctx, pr)
}

func (m *PasswordResetRepository) FindByToken(ctx context.Context, token string) (*domain.PasswordReset, error) {
	return m.FindByTokenFn(ctx, token)
}

func (m *PasswordResetRepository) MarkUsed(ctx context.Context, id string) error {
	return m.MarkUsedFn(ctx, id)
}
