package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/macar-x/cashlenx-server/util"
)

func TestLoggingAddsRequestID(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := util.RequestIDFromContext(r.Context()); got == "" {
			t.Fatal("expected request ID in context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/open/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(util.RequestIDHeader) == "" {
		t.Fatalf("expected %s response header", util.RequestIDHeader)
	}
}

func TestLoggingPreservesIncomingRequestID(t *testing.T) {
	const requestID = "request-1"
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := util.RequestIDFromContext(r.Context()); got != requestID {
			t.Fatalf("request ID in context = %q, want %q", got, requestID)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v0/open/health", nil)
	req.Header.Set(util.RequestIDHeader, requestID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(util.RequestIDHeader); got != requestID {
		t.Fatalf("%s response header = %q, want %q", util.RequestIDHeader, got, requestID)
	}
}
