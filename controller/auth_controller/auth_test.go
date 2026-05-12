package auth_controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/macar-x/cashlenx-server/util"
)

func TestLogoutWithoutTokenReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/logout", nil)
	req = req.WithContext(util.ContextWithRequestID(req.Context(), "request-1"))
	rec := httptest.NewRecorder()

	Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body util.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("code = %q, want OK", body.Code)
	}
	if body.Message != "" {
		t.Fatalf("message = %q, want empty wrapper message", body.Message)
	}
}
