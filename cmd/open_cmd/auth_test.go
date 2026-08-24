package open_cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOpenAuthValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		reset func()
		run   func() error
	}{
		{
			name: "login requires credentials",
			reset: func() {
				loginUsername = ""
				loginPassword = ""
				loginRefreshToken = ""
			},
			run: func() error { return loginCmd.RunE(loginCmd, nil) },
		},
		{
			name: "register requires all fields",
			reset: func() {
				registerUsername = ""
				registerPassword = ""
				registerEmail = ""
				registerVerificationToken = ""
			},
			run: func() error { return registerCmd.RunE(registerCmd, nil) },
		},
		{
			name: "reset password requires identity",
			reset: func() {
				resetEmailOrUsername = ""
			},
			run: func() error { return resetPasswordCmd.RunE(resetPasswordCmd, nil) },
		},
		{
			name: "reset password confirm requires token and password",
			reset: func() {
				resetToken = ""
				resetPassword = ""
			},
			run: func() error { return resetPasswordConfirmCmd.RunE(resetPasswordConfirmCmd, nil) },
		},
		{
			name: "send verification code requires purpose and email",
			reset: func() {
				verificationPurpose = ""
				verificationEmail = ""
			},
			run: func() error { return sendVerificationCodeCmd.RunE(sendVerificationCodeCmd, nil) },
		},
		{
			name: "verify code requires purpose email and code",
			reset: func() {
				verificationPurpose = ""
				verificationEmail = ""
				verificationCode = ""
			},
			run: func() error { return verifyVerificationCodeCmd.RunE(verifyVerificationCodeCmd, nil) },
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

func TestOpenAuthAndVerificationCommandsPassInputsToServices(t *testing.T) {
	userID := primitive.NewObjectID()
	var authenticateArgs, refreshArgs, saveArgs, registerArgs, resetArgs, resetConfirmArgs, sendCodeArgs, verifyArgs struct {
		first  string
		second string
		third  string
		userID string
	}
	var revokedToken, revokedBy, revokedAllUserID string

	originalAuthenticate := authenticateOpenUser
	authenticateOpenUser = func(username, password, deviceID, deviceName, ipAddress, userAgent string) (string, string, model.UserEntity, error) {
		authenticateArgs.first = username
		authenticateArgs.second = password
		authenticateArgs.third = deviceID
		return "access-token", "refresh-token", testOpenUser(userID), nil
	}
	originalRefresh := refreshOpenToken
	refreshOpenToken = func(token, deviceID, deviceName, ipAddress, userAgent string) (string, string, model.UserEntity, error) {
		refreshArgs.first = token
		refreshArgs.second = deviceID
		return "new-access-token", "new-refresh-token", testOpenUser(userID), nil
	}
	originalSave := saveOpenSession
	saveOpenSession = func(accessToken, refreshToken string, user model.UserEntity) error {
		saveArgs.first = accessToken
		saveArgs.second = refreshToken
		saveArgs.userID = user.Id.Hex()
		return nil
	}
	originalRegister := registerOpenUser
	registerOpenUser = func(username, password, email, token string) (string, error) {
		registerArgs.first = username
		registerArgs.second = email
		registerArgs.third = token
		return userID.Hex(), nil
	}
	originalGetRefreshToken := getOpenRefreshToken
	getOpenRefreshToken = func(token, deviceID, deviceName, ipAddress, userAgent string) (model.RefreshToken, error) {
		return model.RefreshToken{Token: token, UserId: userID.Hex()}, nil
	}
	originalRevoke := revokeOpenRefreshToken
	revokeOpenRefreshToken = func(token, userID string) error {
		revokedToken = token
		revokedBy = userID
		return nil
	}
	originalValidate := validateOpenAccessToken
	validateOpenAccessToken = func(token string) (*provider.Claims, error) {
		return &provider.Claims{UserID: userID.Hex(), Username: "alice", Role: model.UserRoleUser}, nil
	}
	originalRevokeAll := revokeAllOpenRefreshTokens
	revokeAllOpenRefreshTokens = func(userID string) error {
		revokedAllUserID = userID
		return nil
	}
	originalClear := clearOpenSession
	clearOpenSession = func() error { return nil }
	originalReset := requestOpenPasswordReset
	requestOpenPasswordReset = func(identity, ipAddress string) error {
		resetArgs.first = identity
		resetArgs.second = ipAddress
		return nil
	}
	originalResetConfirm := confirmOpenPasswordReset
	confirmOpenPasswordReset = func(token, password string) error {
		resetConfirmArgs.first = token
		resetConfirmArgs.second = password
		return nil
	}
	originalSendCode := sendOpenVerificationCode
	sendOpenVerificationCode = func(purpose, email, ipAddress string) error {
		sendCodeArgs.first = purpose
		sendCodeArgs.second = email
		sendCodeArgs.third = ipAddress
		return nil
	}
	originalVerifyCode := verifyOpenCode
	verifyOpenCode = func(purpose, email, code string) (model.VerifyVerificationCodeResponse, error) {
		verifyArgs.first = purpose
		verifyArgs.second = email
		verifyArgs.third = code
		return model.VerifyVerificationCodeResponse{Token: "verification-token", ExpiresAt: time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)}, nil
	}
	t.Cleanup(func() {
		authenticateOpenUser = originalAuthenticate
		refreshOpenToken = originalRefresh
		saveOpenSession = originalSave
		registerOpenUser = originalRegister
		getOpenRefreshToken = originalGetRefreshToken
		revokeOpenRefreshToken = originalRevoke
		validateOpenAccessToken = originalValidate
		revokeAllOpenRefreshTokens = originalRevokeAll
		clearOpenSession = originalClear
		requestOpenPasswordReset = originalReset
		confirmOpenPasswordReset = originalResetConfirm
		sendOpenVerificationCode = originalSendCode
		verifyOpenCode = originalVerifyCode
		resetOpenCommandState()
	})

	loginUsername, loginPassword = "alice", "secret"
	if err := loginCmd.RunE(loginCmd, nil); err != nil {
		t.Fatalf("login RunE returned error: %v", err)
	}
	resetOpenCommandState()
	loginRefreshToken = "refresh-token"
	if err := loginCmd.RunE(loginCmd, nil); err != nil {
		t.Fatalf("refresh login RunE returned error: %v", err)
	}
	registerUsername, registerPassword, registerEmail, registerVerificationToken = "alice", "secret", "alice@example.test", "verification-token"
	if err := registerCmd.RunE(registerCmd, nil); err != nil {
		t.Fatalf("register RunE returned error: %v", err)
	}
	logoutRefreshToken = "refresh-token"
	if err := logoutCmd.RunE(logoutCmd, nil); err != nil {
		t.Fatalf("logout refresh RunE returned error: %v", err)
	}
	logoutRefreshToken, logoutAccessToken = "", "access-token"
	if err := logoutCmd.RunE(logoutCmd, nil); err != nil {
		t.Fatalf("logout access RunE returned error: %v", err)
	}
	resetEmailOrUsername = "alice@example.test"
	if err := resetPasswordCmd.RunE(resetPasswordCmd, nil); err != nil {
		t.Fatalf("reset password RunE returned error: %v", err)
	}
	resetToken, resetPassword = "reset-token", "new-secret"
	if err := resetPasswordConfirmCmd.RunE(resetPasswordConfirmCmd, nil); err != nil {
		t.Fatalf("reset password confirm RunE returned error: %v", err)
	}
	verificationPurpose, verificationEmail = "password_reset", "alice@example.test"
	if err := sendVerificationCodeCmd.RunE(sendVerificationCodeCmd, nil); err != nil {
		t.Fatalf("send verification RunE returned error: %v", err)
	}
	verificationCode = "123456"
	if err := verifyVerificationCodeCmd.RunE(verifyVerificationCodeCmd, nil); err != nil {
		t.Fatalf("verify code RunE returned error: %v", err)
	}

	if authenticateArgs.first != "alice" || authenticateArgs.second != "secret" || authenticateArgs.third != cli_auth.DeviceID {
		t.Fatalf("authenticate args = %+v", authenticateArgs)
	}
	if refreshArgs.first != "refresh-token" || refreshArgs.second != cli_auth.DeviceID {
		t.Fatalf("refresh args = %+v", refreshArgs)
	}
	if saveArgs.userID != userID.Hex() || saveArgs.first == "" || saveArgs.second == "" {
		t.Fatalf("save args = %+v", saveArgs)
	}
	if registerArgs.first != "alice" || registerArgs.second != "alice@example.test" || registerArgs.third != "verification-token" {
		t.Fatalf("register args = %+v", registerArgs)
	}
	if revokedToken != "refresh-token" || revokedBy != userID.Hex() {
		t.Fatalf("revoke args = token %q by %q", revokedToken, revokedBy)
	}
	if revokedAllUserID != userID.Hex() {
		t.Fatalf("revoke all user id = %q", revokedAllUserID)
	}
	if resetArgs.first != "alice@example.test" || resetArgs.second != cli_auth.IPAddress {
		t.Fatalf("reset args = %+v", resetArgs)
	}
	if resetConfirmArgs.first != "reset-token" || resetConfirmArgs.second != "new-secret" {
		t.Fatalf("reset confirm args = %+v", resetConfirmArgs)
	}
	if sendCodeArgs.first != "password_reset" || sendCodeArgs.second != "alice@example.test" || sendCodeArgs.third != cli_auth.IPAddress {
		t.Fatalf("send code args = %+v", sendCodeArgs)
	}
	if verifyArgs.first != "password_reset" || verifyArgs.second != "alice@example.test" || verifyArgs.third != "123456" {
		t.Fatalf("verify args = %+v", verifyArgs)
	}
}

func TestStartCommandUsesConfiguredOrExplicitPort(t *testing.T) {
	var started []int32
	originalStart := startOpenServer
	startOpenServer = func(port int32) {
		started = append(started, port)
	}
	t.Cleanup(func() {
		startOpenServer = originalStart
		port = 8080
		_ = startCmd.Flags().Set("port", "8080")
	})

	_ = startCmd.Flags().Set("port", "9090")
	startCmd.Run(startCmd, nil)

	if len(started) != 1 || started[0] != 9090 {
		t.Fatalf("started ports = %+v", started)
	}
}

func TestStartCommandRejectsUnsupportedTimezone(t *testing.T) {
	originalTimezone := util.GetConfigByKey("timezone")
	util.SetConfigByKey("timezone", "UTC+8")
	t.Cleanup(func() {
		util.SetConfigByKey("timezone", originalTimezone)
	})

	err := startCmd.PreRunE(startCmd, nil)
	if err == nil {
		t.Fatal("PreRunE() error = nil")
	}
	if !strings.Contains(err.Error(), "TIMEZONE") {
		t.Fatalf("PreRunE() error does not identify TIMEZONE: %v", err)
	}
	if strings.Contains(err.Error(), "UTC+8") {
		t.Fatalf("PreRunE() error exposed configured value: %v", err)
	}
}

func testOpenUser(id primitive.ObjectID) model.UserEntity {
	return model.UserEntity{Id: id, Username: "alice", Role: model.UserRoleUser, IsActive: true}
}

func resetOpenCommandState() {
	loginUsername = ""
	loginPassword = ""
	loginRefreshToken = ""
	registerUsername = ""
	registerPassword = ""
	registerEmail = ""
	registerVerificationToken = ""
	logoutRefreshToken = ""
	logoutAccessToken = ""
	resetEmailOrUsername = ""
	resetToken = ""
	resetPassword = ""
	verificationPurpose = ""
	verificationEmail = ""
	verificationCode = ""
}
