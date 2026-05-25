package admin_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAdminRequiresAdminSession(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))
	adminSessionUserId = ""

	if err := AdminCmd.PersistentPreRunE(AdminCmd, nil); err == nil {
		t.Fatal("expected missing CLI session error")
	}
	if adminSessionUserId != "" {
		t.Fatalf("adminSessionUserId = %q, want empty", adminSessionUserId)
	}
}

func TestAdminPrintHelpers(t *testing.T) {
	printAdminUser(model.UserEntity{
		Id:              primitive.NewObjectID(),
		Username:        "admin",
		Role:            model.UserRoleAdmin,
		IsActive:        true,
		Nickname:        "Administrator",
		EmailAddress:    "admin@example.test",
		IsEmailVerified: true,
		Gender:          "others",
		BaseEntity: model.BaseEntity{
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	})

	if got := truncateUserField("abcdef", 3); got != "abc" {
		t.Fatalf("truncateUserField tiny limit = %q, want abc", got)
	}
	if got := truncateUserField("abcdefghijkl", 8); got != "abcde..." {
		t.Fatalf("truncateUserField ellipsis = %q, want abcde...", got)
	}
	if got := truncateUserField("abc", 8); got != "abc" {
		t.Fatalf("truncateUserField short = %q, want abc", got)
	}
}
