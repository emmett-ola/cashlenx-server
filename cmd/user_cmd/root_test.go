package user_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserRootRequiresLoggedInUser(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))
	userId = ""

	if err := UserCmd.PersistentPreRunE(UserCmd, nil); err == nil {
		t.Fatal("expected missing CLI session error")
	}
}

func TestUserCommandValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		reset func()
		run   func() error
	}{
		{
			name: "password requires old and new password",
			reset: func() {
				oldPassword = ""
				newPassword = ""
			},
			run: func() error { return passwordCmd.RunE(passwordCmd, nil) },
		},
		{
			name: "email change requires new email",
			reset: func() {
				emailNewEmail = ""
				emailVerificationToken = ""
			},
			run: func() error { return emailChangeCmd.RunE(emailChangeCmd, nil) },
		},
		{
			name: "email confirm requires token",
			reset: func() {
				emailConfirmToken = ""
				emailConfirmPassword = ""
			},
			run: func() error { return emailConfirmCmd.RunE(emailConfirmCmd, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reset()
			if err := tt.run(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestUserPrintHelpers(t *testing.T) {
	userID := primitive.NewObjectID()
	now := time.Now()

	printUser(model.UserEntity{
		Id:              userID,
		Username:        "tester",
		Role:            model.UserRoleUser,
		IsActive:        true,
		Nickname:        "Test User",
		EmailAddress:    "tester@example.test",
		IsEmailVerified: true,
		Gender:          "others",
		BaseEntity: model.BaseEntity{
			CreateTime: now,
			UpdateTime: now,
		},
	})

	printConfiguration(model.UserConfigurationEntity{
		Id:               primitive.NewObjectID(),
		BelongsUserId:    userID,
		DisplayLanguage:  "en-US",
		CurrencyCode:     "USD",
		ActiveThemeColor: "blue",
	})

	printStats(manage_service.OperationStats{
		Users:       manage_service.EntityStats{Success: 1, Failed: 2},
		UserConfigs: manage_service.EntityStats{Success: 3, Failed: 4},
		Categories:  manage_service.EntityStats{Success: 5, Failed: 6},
		CashFlows:   manage_service.EntityStats{Success: 7, Failed: 8},
	})
}
