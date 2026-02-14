package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lguilherme/contas/internal/domain"
)

func TestCategory_CreateAndList(t *testing.T) {
	cleanDB(t)

	user := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, user.AccessToken, "Casa")

	// Create category
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Mercado", Icon: "🛒"},
		user.AccessToken, http.StatusCreated)

	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)

	if cat["name"] != "Mercado" || cat["icon"] != "🛒" {
		t.Errorf("unexpected category: %+v", cat)
	}

	// List categories
	resp = doGet(t, fmt.Sprintf("/households/%s/categories", hhID), user.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 1 {
		t.Fatalf("expected 1 category, got %d", len(list))
	}
	if list[0]["name"] != "Mercado" {
		t.Errorf("expected 'Mercado', got %v", list[0]["name"])
	}
}

func TestCategory_Update(t *testing.T) {
	cleanDB(t)

	user := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, user.AccessToken, "Casa")

	// Create
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Energia", Icon: "⚡"},
		user.AccessToken, http.StatusCreated)

	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := cat["id"].(string)

	// Update
	resp = doJSON(t, http.MethodPut, fmt.Sprintf("/households/%s/categories/%s", hhID, catID),
		domain.UpdateCategoryInput{Name: "Luz", Icon: "💡"},
		user.AccessToken, http.StatusOK)

	var updated map[string]interface{}
	decodeJSON(t, resp, &updated)

	if updated["name"] != "Luz" || updated["icon"] != "💡" {
		t.Errorf("expected updated category, got %+v", updated)
	}
}

func TestCategory_Delete(t *testing.T) {
	cleanDB(t)

	user := registerUser(t, "Alice", "alice@test.com", "secret123")
	hhID := createHousehold(t, user.AccessToken, "Casa")

	// Create
	resp := doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Internet", Icon: "🌐"},
		user.AccessToken, http.StatusCreated)

	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := cat["id"].(string)

	// Delete
	doJSON(t, http.MethodDelete, fmt.Sprintf("/households/%s/categories/%s", hhID, catID),
		nil, user.AccessToken, http.StatusNoContent)

	// List should be empty
	resp = doGet(t, fmt.Sprintf("/households/%s/categories", hhID), user.AccessToken, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)

	if len(list) != 0 {
		t.Errorf("expected 0 categories after delete, got %d", len(list))
	}
}

func TestCategory_NonMemberCannotAccess(t *testing.T) {
	cleanDB(t)

	alice := registerUser(t, "Alice", "alice@test.com", "secret123")
	bob := registerUser(t, "Bob", "bob@test.com", "secret456")

	hhID := createHousehold(t, alice.AccessToken, "Casa da Alice")

	// Bob tries to list Alice's categories
	doGet(t, fmt.Sprintf("/households/%s/categories", hhID), bob.AccessToken, http.StatusForbidden)

	// Bob tries to create a category in Alice's household
	doJSON(t, http.MethodPost, fmt.Sprintf("/households/%s/categories", hhID),
		domain.CreateCategoryInput{Name: "Hack", Icon: "💀"},
		bob.AccessToken, http.StatusForbidden)
}
