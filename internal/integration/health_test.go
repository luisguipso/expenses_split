package integration

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	resp := doGet(t, "/health", "", http.StatusOK)
	var result map[string]string
	decodeJSON(t, resp, &result)

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}
