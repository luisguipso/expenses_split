package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestExpense_CreateAndList(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Supermercado",
			AmountCents: 25000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	var exp map[string]interface{}
	decodeJSON(t, resp, &exp)

	if exp["description"] != "Supermercado" {
		t.Errorf("expected 'Supermercado', got %v", exp["description"])
	}
	if exp["paid_by"] != alice.ID {
		t.Errorf("expected paid_by %s, got %v", alice.ID, exp["paid_by"])
	}

	// List
	resp = doGet(t, fmt.Sprintf("/households/%s/expenses", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(list))
	}
}

func TestExpense_FilterByMonthYear(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Create expenses in different months
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Janeiro",
			AmountCents: 10000,
			ExpenseDate: "2024-01-15",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Fevereiro",
			AmountCents: 20000,
			ExpenseDate: "2024-02-15",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Filter for January only
	resp := doGet(t, fmt.Sprintf("/households/%s/expenses?month=1&year=2024", hhID),
		alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 expense in January, got %d", len(list))
	}
	if list[0]["description"] != "Janeiro" {
		t.Errorf("expected 'Janeiro', got %v", list[0]["description"])
	}
}

func TestExpense_FilterByCategory(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	// Create category
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Mercado", Icon: "🛒"},
		alice.AccessToken, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := cat["id"].(string)

	// Create expenses with and without category
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Com categoria",
			AmountCents: 10000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
			CategoryID:  catID,
		}, alice.AccessToken, http.StatusCreated)

	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Sem categoria",
			AmountCents: 5000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Filter by category
	resp = doGet(t, fmt.Sprintf("/households/%s/expenses?category_id=%s", hhID, catID),
		alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 expense with category, got %d", len(list))
	}
	if list[0]["description"] != "Com categoria" {
		t.Errorf("expected 'Com categoria', got %v", list[0]["description"])
	}
}

func TestExpense_SharedVsPersonal(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")
	hhID := createHousehold(t, alice.AccessToken, "Casa")
	joinHousehold(t, bob.AccessToken, hhID, alice.AccessToken)

	// Shared expense
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Mercado",
			AmountCents: 30000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	// Personal expense assigned to Bob
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Remédio Bob",
			AmountCents: 5000,
			ExpenseDate: "2024-02-10",
			IsShared:    false,
			AssignedTo:  bob.ID,
		}, alice.AccessToken, http.StatusCreated)

	resp := doGet(t, fmt.Sprintf("/households/%s/expenses", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(list))
	}

	var shared, personal int
	for _, e := range list {
		if e["is_shared"].(bool) {
			shared++
		} else {
			personal++
			if e["assigned_to"] != bob.ID {
				t.Errorf("personal expense should be assigned to Bob")
			}
		}
	}
	if shared != 1 || personal != 1 {
		t.Errorf("expected 1 shared + 1 personal, got %d + %d", shared, personal)
	}
}

func TestExpense_Delete(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, alice.AccessToken, "Casa")

	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/expenses", hhID),
		domain.CreateExpenseInput{
			Description: "Para deletar",
			AmountCents: 1000,
			ExpenseDate: "2024-02-10",
			IsShared:    true,
		}, alice.AccessToken, http.StatusCreated)

	var exp map[string]interface{}
	decodeJSON(t, resp, &exp)
	expID := exp["id"].(string)

	doJSON(t, http.MethodDelete, fmt.Sprintf("/households/%s/expenses/%s", hhID, expID),
		nil, alice.AccessToken, http.StatusNoContent)

	resp = doGet(t, fmt.Sprintf("/households/%s/expenses", hhID), alice.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 0 {
		t.Errorf("expected 0 expenses after delete, got %d", len(list))
	}
}
