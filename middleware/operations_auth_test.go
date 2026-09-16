package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBearerTokenProtectsOperationalHandler(t *testing.T) {
	handler := BearerToken(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), "metrics-token")

	unauthorized := httptest.NewRecorder()
	handler.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	request.Header.Set("Authorization", "Bearer metrics-token")
	authorized := httptest.NewRecorder()
	handler.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusNoContent {
		t.Fatalf("authorized status = %d", authorized.Code)
	}
}
