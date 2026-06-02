package auth_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/model"
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

func TestTokensListsCurrentUserRefreshTokens(t *testing.T) {
	const userID = "507f1f77bcf86cd799439011"
	var listedUserID string
	originalRequire := requireAuthUser
	requireAuthUser = func() (*provider.Claims, error) {
		return &provider.Claims{UserID: userID, Username: "alice", Role: model.UserRoleUser}, nil
	}
	originalTokens := getAuthRefreshTokens
	getAuthRefreshTokens = func(serviceUserID string) []model.RefreshToken {
		listedUserID = serviceUserID
		return []model.RefreshToken{{
			Id:         "token-id",
			UserId:     serviceUserID,
			Token:      "refresh-token",
			DeviceName: "CashLenX CLI",
			IPAddress:  "127.0.0.1",
			ExpiresAt:  time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC),
		}}
	}
	t.Cleanup(func() {
		requireAuthUser = originalRequire
		getAuthRefreshTokens = originalTokens
		tokensUserId = ""
	})

	if err := tokensCmd.RunE(tokensCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
	if listedUserID != userID || tokensUserId != userID {
		t.Fatalf("listed user = %q tokensUserId = %q, want %q", listedUserID, tokensUserId, userID)
	}
}
