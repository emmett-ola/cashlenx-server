package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSPrefightAllowsDevLoopbackOrigin(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight request should not reach the next handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v0/open/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:55500")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:55500" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:55500")
	}
}

func TestOriginMatchesRule(t *testing.T) {
	tests := []struct {
		name          string
		origin        string
		allowedOrigin string
		want          bool
	}{
		{
			name:          "exact origin match",
			origin:        "http://localhost:3000",
			allowedOrigin: "http://localhost:3000",
			want:          true,
		},
		{
			name:          "localhost wildcard port match",
			origin:        "http://localhost:53921",
			allowedOrigin: "http://localhost:*",
			want:          true,
		},
		{
			name:          "loopback wildcard port match",
			origin:        "http://127.0.0.1:42000",
			allowedOrigin: "http://127.0.0.1:*",
			want:          true,
		},
		{
			name:          "scheme mismatch is rejected",
			origin:        "https://localhost:53921",
			allowedOrigin: "http://localhost:*",
			want:          false,
		},
		{
			name:          "non loopback host is rejected",
			origin:        "http://example.com:3000",
			allowedOrigin: "http://localhost:*",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := originMatchesRule(tt.origin, tt.allowedOrigin)
			if got != tt.want {
				t.Fatalf("originMatchesRule(%q, %q) = %v, want %v", tt.origin, tt.allowedOrigin, got, tt.want)
			}
		})
	}
}

func TestShouldAllowOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins string
		env            string
		want           bool
	}{
		{
			name:           "configured origin is allowed",
			origin:         "http://localhost:3000",
			allowedOrigins: "http://localhost:3000",
			env:            "prod",
			want:           true,
		},
		{
			name:           "dynamic localhost port is allowed in dev even with old config",
			origin:         "http://localhost:55500",
			allowedOrigins: "http://localhost:3000,http://localhost:8080",
			env:            "dev",
			want:           true,
		},
		{
			name:           "dynamic loopback port is allowed in test",
			origin:         "http://127.0.0.1:55500",
			allowedOrigins: "http://localhost:3000",
			env:            "test",
			want:           true,
		},
		{
			name:           "dynamic localhost port is rejected in prod unless configured",
			origin:         "http://localhost:55500",
			allowedOrigins: "http://localhost:3000",
			env:            "prod",
			want:           false,
		},
		{
			name:           "external origin is rejected in dev when not configured",
			origin:         "http://example.com:55500",
			allowedOrigins: "http://localhost:3000",
			env:            "dev",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldAllowOrigin(tt.origin, tt.allowedOrigins, tt.env)
			if got != tt.want {
				t.Fatalf("shouldAllowOrigin(%q, %q, %q) = %v, want %v", tt.origin, tt.allowedOrigins, tt.env, got, tt.want)
			}
		})
	}
}
