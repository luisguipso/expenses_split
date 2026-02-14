package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestHousehold_CreateAndList(t *testing.T) {
	cleanDB(t)

	user := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, user.AccessToken, "Casa Teste")

	if hhID == "" {
		t.Fatal("household ID should not be empty")
	}

	// List households — should contain the one we created
	resp := doGet(t, "/households", user.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 household, got %d", len(list))
	}
	if list[0]["name"] != "Casa Teste" {
		t.Errorf("expected name 'Casa Teste', got %v", list[0]["name"])
	}
}

func TestHousehold_JoinViaInviteCode(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")

	// Alice creates household
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Get invite code
	resp := doGet(t, "/households/"+hhID, alice.AccessToken, http.StatusOK)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	inviteCode := hh["invite_code"].(string)

	if inviteCode == "" {
		t.Fatal("invite code should not be empty")
	}

	// Bob joins via invite code
	doJSON(t, http.MethodPost, "/households/join",
		domain.JoinHouseholdInput{InviteCode: inviteCode},
		bob.AccessToken, http.StatusOK)

	// Verify Bob is a member
	resp = doGet(t, fmt.Sprintf("/households/%s/members", hhID), alice.AccessToken, http.StatusOK)
	var members []map[string]interface{}
	decodeJSON(t, resp, &members)

	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
}

func TestHousehold_RejectInvalidInviteCode(t *testing.T) {
	cleanDB(t)

	bob := registerUser(t, "Bob", "bob@test.com", "secret456")

	doJSON(t, http.MethodPost, "/households/join",
		domain.JoinHouseholdInput{InviteCode: "INVALID-CODE"},
		bob.AccessToken, http.StatusNotFound)
}

func TestHousehold_UpdateMemberSalary(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Update salary
	doJSON(t, http.MethodPut,
		fmt.Sprintf("/households/%s/members/%s/salary", hhID, alice.ID),
		domain.UpdateSalaryInput{SalaryCents: 500000},
		alice.AccessToken, http.StatusNoContent)

	// Verify salary was updated
	resp := doGet(t, fmt.Sprintf("/households/%s/members", hhID), alice.AccessToken, http.StatusOK)
	var members []map[string]interface{}
	decodeJSON(t, resp, &members)

	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	// JSON numbers decode as float64
	salary := int64(members[0]["salary_cents"].(float64))
	if salary != 500000 {
		t.Errorf("expected salary 500000, got %d", salary)
	}
}

func TestHousehold_NonMemberCannotAccess(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")

	hhID := createHousehold(t, alice.AccessToken, "Casa da Alice")

	// Bob tries to list members of Alice's household
	doGet(t, fmt.Sprintf("/households/%s/members", hhID), bob.AccessToken, http.StatusForbidden)
}
