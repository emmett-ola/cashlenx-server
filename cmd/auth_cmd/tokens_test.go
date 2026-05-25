package auth_cmd

import (
	"path/filepath"
	"testing"
)

func TestTokensRequiresLoggedInUser(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))
	tokensUserId = ""

	if err := tokensCmd.RunE(tokensCmd, nil); err == nil {
		t.Fatal("expected missing CLI session error")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		limit int
		want  string
	}{
		{name: "short", value: "abc", limit: 5, want: "abc"},
		{name: "tiny limit", value: "abcdef", limit: 3, want: "abc"},
		{name: "ellipsis", value: "abcdefghijkl", limit: 8, want: "abcde..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.value, tt.limit); got != tt.want {
				t.Fatalf("truncate() = %q, want %q", got, tt.want)
			}
		})
	}
}
