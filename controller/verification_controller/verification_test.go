package verification_controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
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

func TestSendCodePassesRequestAndClientIPToService(t *testing.T) {
	var got struct {
		purpose string
		email   string
		ip      string
	}
	original := sendVerificationCode
	sendVerificationCode = func(purpose, recipientEmail, ipAddress string) error {
		got.purpose = purpose
		got.email = recipientEmail
		got.ip = ipAddress
		return nil
	}
	t.Cleanup(func() { sendVerificationCode = original })

	req := httptest.NewRequest(http.MethodPost, "/verification/code", strings.NewReader(`{"purpose":"password_reset","email":"alice@example.test"}`))
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec := httptest.NewRecorder()

	SendCode(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got.purpose != "password_reset" || got.email != "alice@example.test" || got.ip != "203.0.113.10" {
		t.Fatalf("service args = %+v", got)
	}
}

func TestVerifyCodePassesRequestToService(t *testing.T) {
	expiresAt := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	var got struct {
		purpose string
		email   string
		code    string
	}
	original := verifyCodeForPurpose
	verifyCodeForPurpose = func(purpose, recipientEmail, code string) (model.VerifyVerificationCodeResponse, error) {
		got.purpose = purpose
		got.email = recipientEmail
		got.code = code
		return model.VerifyVerificationCodeResponse{Token: "verification-token", ExpiresAt: expiresAt}, nil
	}
	t.Cleanup(func() { verifyCodeForPurpose = original })

	req := httptest.NewRequest(http.MethodPost, "/verification/verify", strings.NewReader(`{"purpose":"email_change","email":"alice@example.test","code":"123456"}`))
	rec := httptest.NewRecorder()

	VerifyCode(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got.purpose != "email_change" || got.email != "alice@example.test" || got.code != "123456" {
		t.Fatalf("service args = %+v", got)
	}
	var response struct {
		Data model.VerifyVerificationCodeResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON decode failed: %v", err)
	}
	if response.Data.Token != "verification-token" || !response.Data.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("response data = %+v", response.Data)
	}
}
