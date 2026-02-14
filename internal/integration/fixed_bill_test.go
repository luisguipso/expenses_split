package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestFixedBill_CreateWithPaidBy(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")
	hhID := createHousehold(t, alice.AccessToken, "Casa")
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// Create bill paid by Bob
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Internet",
			AmountCents: 12000,
			DueDay:      10,
			IsShared:    true,
			PaidBy:      bob.ID,
		}, alice.AccessToken, http.StatusCreated)

	var bill map[string]interface{}
	decodeJSON(t, resp, &bill)

	if bill["paid_by"] != bob.ID {
		t.Errorf("expected paid_by %s, got %v", bob.ID, bill["paid_by"])
	}

	// List and verify paid_by_name
	resp = doGet(t, fmt.Sprintf("/households/%s/bills", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 bill, got %d", len(list))
	}
	if list[0]["paid_by_name"] != "Bob" {
		t.Errorf("expected paid_by_name 'Bob', got %v", list[0]["paid_by_name"])
	}
}

func TestFixedBill_DefaultPaidByToCurrentUser(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Create bill without paid_by
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Aluguel",
			AmountCents: 200000,
			DueDay:      5,
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	var bill map[string]interface{}
	decodeJSON(t, resp, &bill)

	if bill["paid_by"] != alice.ID {
		t.Errorf("expected paid_by defaulting to %s, got %v", alice.ID, bill["paid_by"])
	}
}

func TestFixedBill_Update(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Create
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Energia",
			AmountCents: 15000,
			DueDay:      15,
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	var bill map[string]interface{}
	decodeJSON(t, resp, &bill)
	billID := bill["id"].(string)

	// Update
	resp = doJSON(t, http.MethodPut, fmt.Sprintf("/households/%s/bills/%s", hhID, billID),
		domain.UpdateFixedBillInput{
			Description: "Luz",
			AmountCents: 18000,
			DueDay:      20,
			IsShared:    true,
			PaidBy:      alice.ID,
			IsActive:    true,
		}, alice.AccessToken, http.StatusOK)

	var updated map[string]interface{}
	decodeJSON(t, resp, &updated)

	if updated["description"] != "Luz" {
		t.Errorf("expected 'Luz', got %v", updated["description"])
	}
	if int64(updated["amount_cents"].(float64)) != 18000 {
		t.Errorf("expected 18000, got %v", updated["amount_cents"])
	}
}

func TestFixedBill_Delete(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Create
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Água",
			AmountCents: 8000,
			DueDay:      25,
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	var bill map[string]interface{}
	decodeJSON(t, resp, &bill)
	billID := bill["id"].(string)

	// Delete
	doJSON(t, http.MethodDelete, fmt.Sprintf("/households/%s/bills/%s", hhID, billID),
		nil, alice.AccessToken, http.StatusNoContent)

	// List should be empty
	resp = doGet(t, fmt.Sprintf("/households/%s/bills", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 0 {
		t.Errorf("expected 0 bills after delete, got %d", len(list))
	}
}

func TestFixedBill_SharedVsPersonal(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")
	hhID := createHousehold(t, alice.AccessToken, "Casa")
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// Create shared bill
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Aluguel",
			AmountCents: 200000,
			DueDay:      5,
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Create personal bill assigned to Bob
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/bills", hhID),
		domain.CreateFixedBillInput{
			Description: "Plano celular Bob",
			AmountCents: 5000,
			DueDay:      10,
			IsShared:    false,
			AssignedTo:  bob.ID,
			PaidBy:      bob.ID,
		}, alice.AccessToken, http.StatusCreated)

	// List all
	resp := doGet(t, fmt.Sprintf("/households/%s/bills", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 2 {
		t.Fatalf("expected 2 bills, got %d", len(list))
	}

	var shared, personal int
	for _, b := range list {
		if b["is_shared"].(bool) {
			shared++
		} else {
			personal++
			if b["assigned_to"] != bob.ID {
				t.Errorf("personal bill should be assigned to Bob, got %v", b["assigned_to"])
			}
		}
	}
	if shared != 1 || personal != 1 {
		t.Errorf("expected 1 shared + 1 personal, got %d shared + %d personal", shared, personal)
	}
}
