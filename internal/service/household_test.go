package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lguilherme/contas/internal/domain"
	"github.com/lguilherme/contas/internal/mock"
)

func newMockHouseholdRepo() *mock.HouseholdRepository {
	return &mock.HouseholdRepository{}
}

func TestHouseholdService_Create_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.CreateFn = func(ctx context.Context, h *domain.Household, adminUserID string) error {
		h.ID = "hh-1"
		h.CreatedAt = time.Now()
		if adminUserID != "user-1" {
			t.Errorf("expected adminUserID user-1, got %s", adminUserID)
		}
		if h.InviteCode == "" {
			t.Error("invite code should be generated")
		}
		return nil
	}
	svc := NewHouseholdService(repo)

	h, err := svc.Create(context.Background(), domain.CreateHouseholdInput{Name: "Casa"}, "user-1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if h.ID != "hh-1" {
		t.Errorf("expected ID hh-1, got %s", h.ID)
	}
	if h.Name != "Casa" {
		t.Errorf("expected name Casa, got %s", h.Name)
	}
}

func TestHouseholdService_Create_RepoError(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.CreateFn = func(ctx context.Context, h *domain.Household, adminUserID string) error {
		return errors.New("db error")
	}
	svc := NewHouseholdService(repo)

	_, err := svc.Create(context.Background(), domain.CreateHouseholdInput{Name: "Casa"}, "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHouseholdService_GetByID_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member"}, nil
	}
	repo.FindByIDFn = func(ctx context.Context, id string) (*domain.Household, error) {
		return &domain.Household{ID: id, Name: "Casa"}, nil
	}
	svc := NewHouseholdService(repo)

	h, err := svc.GetByID(context.Background(), "hh-1", "user-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if h.Name != "Casa" {
		t.Errorf("expected name Casa, got %s", h.Name)
	}
}

func TestHouseholdService_GetByID_NotMember(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return nil, domain.ErrNotMember
	}
	svc := NewHouseholdService(repo)

	_, err := svc.GetByID(context.Background(), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_Update_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin"}, nil
	}
	repo.UpdateFn = func(ctx context.Context, h *domain.Household) error {
		return nil
	}
	repo.FindByIDFn = func(ctx context.Context, id string) (*domain.Household, error) {
		return &domain.Household{ID: id, Name: "Casa Nova"}, nil
	}
	svc := NewHouseholdService(repo)

	h, err := svc.Update(context.Background(), "hh-1", domain.UpdateHouseholdInput{Name: "Casa Nova"}, "user-1")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if h.Name != "Casa Nova" {
		t.Errorf("expected name Casa Nova, got %s", h.Name)
	}
}

func TestHouseholdService_Update_NotAdmin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member"}, nil
	}
	svc := NewHouseholdService(repo)

	_, err := svc.Update(context.Background(), "hh-1", domain.UpdateHouseholdInput{Name: "X"}, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_Delete_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin"}, nil
	}
	repo.DeleteFn = func(ctx context.Context, id string) error {
		return nil
	}
	svc := NewHouseholdService(repo)

	if err := svc.Delete(context.Background(), "hh-1", "user-1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestHouseholdService_Delete_NotAdmin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member"}, nil
	}
	svc := NewHouseholdService(repo)

	err := svc.Delete(context.Background(), "hh-1", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_Join_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.FindByInviteCodeFn = func(ctx context.Context, code string) (*domain.Household, error) {
		return &domain.Household{ID: "hh-1", Name: "Casa"}, nil
	}
	repo.AddMemberFn = func(ctx context.Context, householdID, userID, role string) error {
		if role != "member" {
			t.Errorf("expected role member, got %s", role)
		}
		return nil
	}
	svc := NewHouseholdService(repo)

	h, err := svc.Join(context.Background(), "abc123", "user-2")
	if err != nil {
		t.Fatalf("Join failed: %v", err)
	}
	if h.ID != "hh-1" {
		t.Errorf("expected ID hh-1, got %s", h.ID)
	}
}

func TestHouseholdService_Join_InvalidCode(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.FindByInviteCodeFn = func(ctx context.Context, code string) (*domain.Household, error) {
		return nil, domain.ErrInvalidInviteCode
	}
	svc := NewHouseholdService(repo)

	_, err := svc.Join(context.Background(), "bad-code", "user-2")
	if !errors.Is(err, domain.ErrInvalidInviteCode) {
		t.Errorf("expected ErrInvalidInviteCode, got %v", err)
	}
}

func TestHouseholdService_Join_AlreadyMember(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.FindByInviteCodeFn = func(ctx context.Context, code string) (*domain.Household, error) {
		return &domain.Household{ID: "hh-1"}, nil
	}
	repo.AddMemberFn = func(ctx context.Context, householdID, userID, role string) error {
		return domain.ErrAlreadyMember
	}
	svc := NewHouseholdService(repo)

	_, err := svc.Join(context.Background(), "abc123", "user-1")
	if !errors.Is(err, domain.ErrAlreadyMember) {
		t.Errorf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestHouseholdService_UpdateMemberSalary_Admin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin", UserID: "admin-1"}, nil
	}
	repo.UpdateMemberSalaryFn = func(ctx context.Context, householdID, userID string, salaryCents int64) error {
		if salaryCents != 500000 {
			t.Errorf("expected salary 500000, got %d", salaryCents)
		}
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSalary(context.Background(), "hh-1", "user-2", 500000, "admin-1")
	if err != nil {
		t.Fatalf("UpdateMemberSalary failed: %v", err)
	}
}

func TestHouseholdService_UpdateMemberSalary_Self(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: userID}, nil
	}
	repo.UpdateMemberSalaryFn = func(ctx context.Context, householdID, userID string, salaryCents int64) error {
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSalary(context.Background(), "hh-1", "user-1", 300000, "user-1")
	if err != nil {
		t.Fatalf("UpdateMemberSalary (self) failed: %v", err)
	}
}

func TestHouseholdService_UpdateMemberSalary_Forbidden(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: "user-1"}, nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSalary(context.Background(), "hh-1", "user-2", 300000, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_RemoveMember_Admin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin", UserID: "admin-1"}, nil
	}
	repo.RemoveMemberFn = func(ctx context.Context, householdID, userID string) error {
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.RemoveMember(context.Background(), "hh-1", "user-2", "admin-1")
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}
}

func TestHouseholdService_RemoveMember_Self(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: userID}, nil
	}
	repo.RemoveMemberFn = func(ctx context.Context, householdID, userID string) error {
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.RemoveMember(context.Background(), "hh-1", "user-1", "user-1")
	if err != nil {
		t.Fatalf("RemoveMember (self) failed: %v", err)
	}
}

func TestHouseholdService_RemoveMember_Forbidden(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: "user-1"}, nil
	}
	svc := NewHouseholdService(repo)

	err := svc.RemoveMember(context.Background(), "hh-1", "user-2", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_ListMembers_Success(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member"}, nil
	}
	repo.ListMembersFn = func(ctx context.Context, householdID string) ([]domain.HouseholdMember, error) {
		return []domain.HouseholdMember{
			{UserID: "u1", UserName: "Alice"},
			{UserID: "u2", UserName: "Bob"},
		}, nil
	}
	svc := NewHouseholdService(repo)

	members, err := svc.ListMembers(context.Background(), "hh-1", "u1")
	if err != nil {
		t.Fatalf("ListMembers failed: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}

func TestHouseholdService_ListMembers_NotMember(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return nil, domain.ErrNotMember
	}
	svc := NewHouseholdService(repo)

	_, err := svc.ListMembers(context.Background(), "hh-1", "outsider")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- UpdateSplitMode tests ---

func TestHouseholdService_UpdateSplitMode_Admin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin", UserID: "admin-1"}, nil
	}
	repo.UpdateSplitModeFn = func(ctx context.Context, householdID, splitMode string) error {
		if splitMode != "percentage" {
			t.Errorf("expected percentage, got %s", splitMode)
		}
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateSplitMode(context.Background(), "hh-1", "percentage", "admin-1")
	if err != nil {
		t.Fatalf("UpdateSplitMode failed: %v", err)
	}
}

func TestHouseholdService_UpdateSplitMode_NotAdmin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: "user-1"}, nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateSplitMode(context.Background(), "hh-1", "percentage", "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_UpdateSplitMode_InvalidMode(t *testing.T) {
	repo := newMockHouseholdRepo()
	svc := NewHouseholdService(repo)

	err := svc.UpdateSplitMode(context.Background(), "hh-1", "invalid", "admin-1")
	if !errors.Is(err, domain.ErrInvalidSplitMode) {
		t.Errorf("expected ErrInvalidSplitMode, got %v", err)
	}
}

// --- UpdateMemberSplitPercentage tests ---

func TestHouseholdService_UpdateMemberSplitPercentage_Admin(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "admin", UserID: "admin-1"}, nil
	}
	repo.UpdateMemberSplitPercentageFn = func(ctx context.Context, householdID, userID string, percentage int) error {
		if percentage != 6000 {
			t.Errorf("expected 6000, got %d", percentage)
		}
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSplitPercentage(context.Background(), "hh-1", "user-2", 6000, "admin-1")
	if err != nil {
		t.Fatalf("UpdateMemberSplitPercentage failed: %v", err)
	}
}

func TestHouseholdService_UpdateMemberSplitPercentage_Self(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: userID}, nil
	}
	repo.UpdateMemberSplitPercentageFn = func(ctx context.Context, householdID, userID string, percentage int) error {
		return nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSplitPercentage(context.Background(), "hh-1", "user-1", 5000, "user-1")
	if err != nil {
		t.Fatalf("UpdateMemberSplitPercentage (self) failed: %v", err)
	}
}

func TestHouseholdService_UpdateMemberSplitPercentage_Forbidden(t *testing.T) {
	repo := newMockHouseholdRepo()
	repo.GetMemberFn = func(ctx context.Context, householdID, userID string) (*domain.HouseholdMember, error) {
		return &domain.HouseholdMember{Role: "member", UserID: "user-1"}, nil
	}
	svc := NewHouseholdService(repo)

	err := svc.UpdateMemberSplitPercentage(context.Background(), "hh-1", "user-2", 5000, "user-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestHouseholdService_UpdateMemberSplitPercentage_InvalidRange(t *testing.T) {
	repo := newMockHouseholdRepo()
	svc := NewHouseholdService(repo)

	tests := []struct {
		name       string
		percentage int
	}{
		{"negative", -1},
		{"over 10000", 10001},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateMemberSplitPercentage(context.Background(), "hh-1", "user-1", tt.percentage, "user-1")
			if !errors.Is(err, domain.ErrInvalidSplitPercentage) {
				t.Errorf("expected ErrInvalidSplitPercentage, got %v", err)
			}
		})
	}
}
