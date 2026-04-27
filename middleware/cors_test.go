package middleware

import "testing"

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

