package cli_auth

import (
	"path/filepath"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSaveCurrentSessionAndClear(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "cli_auth.json")
	t.Setenv("CASHLENX_CLI_AUTH_FILE", authFile)

	userID := primitive.NewObjectID()
	user := model.UserEntity{
		Id:       userID,
		Username: "alice",
		Role:     model.UserRoleUser,
	}

	if err := Save("access-token", "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	session, err := CurrentSession()
	if err != nil {
		t.Fatalf("CurrentSession returned error: %v", err)
	}
	if session.AccessToken != "access-token" || session.RefreshToken != "refresh-token" {
		t.Fatalf("session tokens = %q/%q", session.AccessToken, session.RefreshToken)
	}
	if session.UserID != userID.Hex() || session.Username != "alice" || session.Role != model.UserRoleUser {
		t.Fatalf("session user = %q/%q/%q", session.UserID, session.Username, session.Role)
	}

	if err := Clear(); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	if _, err := CurrentSession(); err == nil {
		t.Fatal("CurrentSession returned nil error after Clear")
	}
}
