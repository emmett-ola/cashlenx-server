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

func TestProfileGetCommandPassesAuthenticatedUserToService(t *testing.T) {
	userID := primitive.NewObjectID()
	original := getUserProfile
	getUserProfile = func(serviceUserID string) model.UserEntity {
		if serviceUserID != userID.Hex() {
			t.Fatalf("service user id = %q, want %q", serviceUserID, userID.Hex())
		}
		return model.UserEntity{
			Id:       userID,
			Username: "tester",
			IsActive: true,
			Role:     model.UserRoleUser,
		}
	}
	t.Cleanup(func() {
		getUserProfile = original
		userId = ""
	})

	userId = userID.Hex()
	if err := profileGetCmd.RunE(profileGetCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
}

func TestProfileUpdateCommandPassesFlagsToService(t *testing.T) {
	userID := primitive.NewObjectID()
	var got model.UserProfileUpdateRequest
	original := updateUserProfile
	updateUserProfile = func(serviceUserID string, request model.UserProfileUpdateRequest) (model.UserEntity, error) {
		if serviceUserID != userID.Hex() {
			t.Fatalf("service user id = %q, want %q", serviceUserID, userID.Hex())
		}
		got = request
		return model.UserEntity{
			Id:        userID,
			Username:  "tester",
			Nickname:  request.Nickname,
			AvatarUrl: request.AvatarUrl,
			Gender:    request.Gender,
			IsActive:  true,
			Role:      model.UserRoleUser,
		}, nil
	}
	t.Cleanup(func() {
		updateUserProfile = original
		userId = ""
		profileNickname = ""
		profileAvatar = ""
		profileGender = ""
	})

	userId = userID.Hex()
	profileNickname = "Test User"
	profileAvatar = "https://example.test/avatar.png"
	profileGender = model.GenderOthers

	if err := profileUpdateCmd.RunE(profileUpdateCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
	if got.Nickname != "Test User" || got.AvatarUrl != "https://example.test/avatar.png" || got.Gender != model.GenderOthers {
		t.Fatalf("service request = %+v", got)
	}
}

func TestUserConfigurationCommandsPassUserAndRequestToServices(t *testing.T) {
	userID := primitive.NewObjectID()
	var getUserID string
	var upsertUserID string
	var upsertReq model.UserConfigurationRequest
	originalGet := getUserConfig
	getUserConfig = func(serviceUserID string) (model.UserConfigurationEntity, error) {
		getUserID = serviceUserID
		return testUserConfig(userID), nil
	}
	originalUpsert := upsertUserConfig
	upsertUserConfig = func(serviceUserID string, req model.UserConfigurationRequest) (model.UserConfigurationEntity, error) {
		upsertUserID = serviceUserID
		upsertReq = req
		return testUserConfig(userID), nil
	}
	t.Cleanup(func() {
		getUserConfig = originalGet
		upsertUserConfig = originalUpsert
		resetUserCommandState()
		_ = configurationUpsertCmd.Flags().Set("display-language", "")
		_ = configurationUpsertCmd.Flags().Set("currency-code", "")
		_ = configurationUpsertCmd.Flags().Set("active-theme-color", "")
	})

	userId = userID.Hex()
	if err := configurationGetCmd.RunE(configurationGetCmd, nil); err != nil {
		t.Fatalf("configuration get RunE returned error: %v", err)
	}
	_ = configurationUpsertCmd.Flags().Set("display-language", "en-US")
	_ = configurationUpsertCmd.Flags().Set("currency-code", "USD")
	_ = configurationUpsertCmd.Flags().Set("active-theme-color", "blue")
	if err := configurationUpsertCmd.RunE(configurationUpsertCmd, nil); err != nil {
		t.Fatalf("configuration upsert RunE returned error: %v", err)
	}

	if getUserID != userID.Hex() || upsertUserID != userID.Hex() {
		t.Fatalf("user ids = get %q, upsert %q", getUserID, upsertUserID)
	}
	if upsertReq.DisplayLanguage == nil || *upsertReq.DisplayLanguage != "en-US" || upsertReq.CurrencyCode == nil || *upsertReq.CurrencyCode != "USD" || upsertReq.ActiveThemeColor == nil || *upsertReq.ActiveThemeColor != "blue" {
		t.Fatalf("upsert request = %+v", upsertReq)
	}
}

func TestUserPasswordEmailAccountAndDatabaseCommandsPassUserToServices(t *testing.T) {
	userID := primitive.NewObjectID()
	var passwordArgs, emailChangeArgs, emailConfirmArgs struct {
		userID string
		first  string
		second string
	}
	var deletedUserID string
	var exportArgs, importArgs struct {
		userID string
		path   string
	}

	originalPassword := changeUserPassword
	changeUserPassword = func(serviceUserID, oldPass, newPass string) error {
		passwordArgs.userID = serviceUserID
		passwordArgs.first = oldPass
		passwordArgs.second = newPass
		return nil
	}
	originalEmailChange := requestUserEmailChange
	requestUserEmailChange = func(serviceUserID, newEmail, verificationToken string) error {
		emailChangeArgs.userID = serviceUserID
		emailChangeArgs.first = newEmail
		emailChangeArgs.second = verificationToken
		return nil
	}
	originalEmailConfirm := confirmUserEmailChange
	confirmUserEmailChange = func(serviceUserID, token, password string) error {
		emailConfirmArgs.userID = serviceUserID
		emailConfirmArgs.first = token
		emailConfirmArgs.second = password
		return nil
	}
	originalDelete := deleteUserAccount
	deleteUserAccount = func(serviceUserID string) error {
		deletedUserID = serviceUserID
		return nil
	}
	originalExport := exportUserDataWithProgress
	exportUserDataWithProgress = func(serviceUserID, path string, _ manage_service.ProgressFunc) (manage_service.OperationStats, error) {
		exportArgs.userID = serviceUserID
		exportArgs.path = path
		return manage_service.OperationStats{CashFlows: manage_service.EntityStats{Success: 1}}, nil
	}
	originalImport := importUserDataWithProgress
	importUserDataWithProgress = func(serviceUserID, path string, _ manage_service.ProgressFunc) (manage_service.OperationStats, error) {
		importArgs.userID = serviceUserID
		importArgs.path = path
		return manage_service.OperationStats{Categories: manage_service.EntityStats{Success: 1}}, nil
	}
	t.Cleanup(func() {
		changeUserPassword = originalPassword
		requestUserEmailChange = originalEmailChange
		confirmUserEmailChange = originalEmailConfirm
		deleteUserAccount = originalDelete
		exportUserDataWithProgress = originalExport
		importUserDataWithProgress = originalImport
		resetUserCommandState()
	})

	userId = userID.Hex()
	oldPassword, newPassword = "old-secret", "new-secret"
	if err := passwordCmd.RunE(passwordCmd, nil); err != nil {
		t.Fatalf("password RunE returned error: %v", err)
	}
	emailNewEmail, emailVerificationToken = "new@example.test", "verification-token"
	if err := emailChangeCmd.RunE(emailChangeCmd, nil); err != nil {
		t.Fatalf("email change RunE returned error: %v", err)
	}
	emailConfirmToken, emailConfirmPassword = "confirm-token", "password"
	if err := emailConfirmCmd.RunE(emailConfirmCmd, nil); err != nil {
		t.Fatalf("email confirm RunE returned error: %v", err)
	}
	accountForce = true
	if err := accountCmd.RunE(accountCmd, nil); err != nil {
		t.Fatalf("account RunE returned error: %v", err)
	}
	userBackupPath = "backup.json"
	if err := databaseBackupCmd.RunE(databaseBackupCmd, nil); err != nil {
		t.Fatalf("database backup RunE returned error: %v", err)
	}
	userRestorePath = "restore.json"
	if err := databaseRestoreCmd.RunE(databaseRestoreCmd, nil); err != nil {
		t.Fatalf("database restore RunE returned error: %v", err)
	}

	if passwordArgs.userID != userID.Hex() || passwordArgs.first != "old-secret" || passwordArgs.second != "new-secret" {
		t.Fatalf("password args = %+v", passwordArgs)
	}
	if emailChangeArgs.userID != userID.Hex() || emailChangeArgs.first != "new@example.test" || emailChangeArgs.second != "verification-token" {
		t.Fatalf("email change args = %+v", emailChangeArgs)
	}
	if emailConfirmArgs.userID != userID.Hex() || emailConfirmArgs.first != "confirm-token" || emailConfirmArgs.second != "password" {
		t.Fatalf("email confirm args = %+v", emailConfirmArgs)
	}
	if deletedUserID != userID.Hex() {
		t.Fatalf("deleted user id = %q", deletedUserID)
	}
	if exportArgs.userID != userID.Hex() || exportArgs.path != "backup.json" {
		t.Fatalf("export args = %+v", exportArgs)
	}
	if importArgs.userID != userID.Hex() || importArgs.path != "restore.json" {
		t.Fatalf("import args = %+v", importArgs)
	}
}

func testUserConfig(userID primitive.ObjectID) model.UserConfigurationEntity {
	return model.UserConfigurationEntity{
		Id:               primitive.NewObjectID(),
		BelongsUserId:    userID,
		DisplayLanguage:  "en-US",
		CurrencyCode:     "USD",
		ActiveThemeColor: "blue",
	}
}

func resetUserCommandState() {
	userId = ""
	configLanguage = ""
	configCurrency = ""
	configTheme = ""
	oldPassword = ""
	newPassword = ""
	emailNewEmail = ""
	emailVerificationToken = ""
	emailConfirmToken = ""
	emailConfirmPassword = ""
	accountForce = false
	userBackupPath = ""
	userRestorePath = ""
	profileNickname = ""
	profileAvatar = ""
	profileGender = ""
}
