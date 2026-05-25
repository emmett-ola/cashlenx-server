package verification_controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerificationHandlersRejectInvalidJSON(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{"send code", SendCode},
		{"verify code", VerifyCode},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/verification", strings.NewReader("{"))
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}
