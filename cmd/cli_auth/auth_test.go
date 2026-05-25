package cli_auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/mapper/refresh_token_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSaveCurrentSessionAndClear(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)

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

func TestClearIgnoresMissingSessionFile(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)

	if err := Clear(); err != nil {
		t.Fatalf("Clear returned error for missing session: %v", err)
	}
}

func TestClearReturnsUnexpectedRemoveError(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "session-dir")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	if err := os.MkdirAll(authFile, 0o700); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authFile, "child"), []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if err := Clear(); err == nil {
		t.Fatal("Clear returned nil error for non-empty directory session path")
	}
}

func TestRequireUserReturnsLoginErrorWhenSessionMissing(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)

	if _, err := RequireUser(); err == nil {
		t.Fatal("RequireUser returned nil error for missing session")
	}
}

func TestSessionPathDefaultsToConfigDir(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", "")
	t.Setenv("CASHLENX_CLI_AUTH_FILE", "")

	path, err := sessionPath()
	if err != nil {
		t.Fatalf("sessionPath returned error: %v", err)
	}
	if filepath.Base(path) != "session" || filepath.Base(filepath.Dir(path)) != ".cli" {
		t.Fatalf("sessionPath = %q, want cashlenx/.cli/session suffix", path)
	}
}

func TestRequireUserUsesStoredAccessToken(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{
		Id:       userID,
		Username: "alice",
		Role:     model.UserRoleUser,
		IsActive: true,
	}
	installUserMapperStub(t, user)

	accessToken, err := auth.Service.GenerateToken(userID.Hex(), "alice", model.UserRoleUser)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := Save(accessToken, "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	claims, err := RequireUser()
	if err != nil {
		t.Fatalf("RequireUser returned error: %v", err)
	}
	if claims.UserID != userID.Hex() || claims.Username != "alice" || claims.Role != model.UserRoleUser {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestRequireUserIDRejectsMismatchedUserFlag(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{Id: userID, Username: "alice", Role: model.UserRoleUser, IsActive: true}
	installUserMapperStub(t, user)

	accessToken, err := auth.Service.GenerateToken(userID.Hex(), "alice", model.UserRoleUser)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := Save(accessToken, "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	target := primitive.NewObjectID().Hex()
	if err := RequireUserID(&target); err == nil {
		t.Fatal("RequireUserID returned nil error for mismatched user")
	}
}

func TestRequireAdminRejectsNonAdminUser(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{Id: userID, Username: "alice", Role: model.UserRoleUser, IsActive: true}
	installUserMapperStub(t, user)

	accessToken, err := auth.Service.GenerateToken(userID.Hex(), "alice", model.UserRoleUser)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := Save(accessToken, "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if _, err := RequireAdmin(); err == nil {
		t.Fatal("RequireAdmin returned nil error for non-admin user")
	}
}

func TestRequireAdminAcceptsAdminUser(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{Id: userID, Username: "admin", Role: model.UserRoleAdmin, IsActive: true}
	installUserMapperStub(t, user)

	accessToken, err := auth.Service.GenerateToken(userID.Hex(), "admin", model.UserRoleAdmin)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := Save(accessToken, "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	claims, err := RequireAdmin()
	if err != nil {
		t.Fatalf("RequireAdmin returned error: %v", err)
	}
	if claims.Role != model.UserRoleAdmin {
		t.Fatalf("role = %q, want %q", claims.Role, model.UserRoleAdmin)
	}
}

func TestRequireUserIDSetsEmptyTargetToLoggedInUser(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{Id: userID, Username: "alice", Role: model.UserRoleUser, IsActive: true}
	installUserMapperStub(t, user)

	accessToken, err := auth.Service.GenerateToken(userID.Hex(), "alice", model.UserRoleUser)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := Save(accessToken, "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	target := ""
	if err := RequireUserID(&target); err != nil {
		t.Fatalf("RequireUserID returned error: %v", err)
	}
	if target != userID.Hex() {
		t.Fatalf("target = %q, want %q", target, userID.Hex())
	}
}

func TestRequireUserClearsSessionWhenRefreshFails(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)
	installRefreshTokenMapperStub(t)

	user := model.UserEntity{Id: primitive.NewObjectID(), Username: "alice", Role: model.UserRoleUser}
	if err := Save("invalid-access-token", "missing-refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if _, err := RequireUser(); err == nil {
		t.Fatal("RequireUser returned nil error for invalid refresh token")
	}
	if _, err := CurrentSession(); err == nil {
		t.Fatal("CurrentSession returned nil error after failed refresh cleared session")
	}
}

func TestRequireUserRefreshesExpiredAccessTokenAndReplacesSession(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	restoreAuthConfig(t)

	userID := primitive.NewObjectID()
	user := model.UserEntity{Id: userID, Username: "alice", Role: model.UserRoleUser, IsActive: true}
	installUserMapperStub(t, user)
	refreshMapper := installRefreshTokenMapperStub(t)
	refreshMapper.token = model.RefreshToken{
		Id:        "refresh-id",
		UserId:    userID.Hex(),
		Token:     "old-refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
		UserAgent: UserAgent,
	}
	refreshMapper.createdToken = "new-refresh-token"

	if err := Save("expired-access-token", "old-refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	claims, err := RequireUser()
	if err != nil {
		t.Fatalf("RequireUser returned error: %v", err)
	}
	if claims.UserID != userID.Hex() {
		t.Fatalf("claims.UserID = %q, want %q", claims.UserID, userID.Hex())
	}
	if refreshMapper.revokedToken != "old-refresh-token" {
		t.Fatalf("revokedToken = %q, want old-refresh-token", refreshMapper.revokedToken)
	}
	session, err := CurrentSession()
	if err != nil {
		t.Fatalf("CurrentSession returned error: %v", err)
	}
	if session.AccessToken == "expired-access-token" {
		t.Fatal("access token was not replaced")
	}
	if session.RefreshToken != "new-refresh-token" {
		t.Fatalf("refresh token = %q, want new-refresh-token", session.RefreshToken)
	}
}

func TestCurrentSessionRejectsMalformedJSON(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), ".cli", "session")
	t.Setenv("CASHLENX_CLI_SESSION_FILE", authFile)
	if err := os.MkdirAll(filepath.Dir(authFile), 0o700); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(authFile, []byte("{"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	if _, err := CurrentSession(); err == nil {
		t.Fatal("CurrentSession returned nil error for malformed JSON")
	}
}

func TestCurrentSessionSupportsLegacyAuthFileOverride(t *testing.T) {
	authFile := filepath.Join(t.TempDir(), "legacy_auth.json")
	t.Setenv("CASHLENX_CLI_AUTH_FILE", authFile)

	user := model.UserEntity{Id: primitive.NewObjectID(), Username: "alice", Role: model.UserRoleUser}
	if err := Save("access-token", "refresh-token", user); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if _, err := os.Stat(authFile); err != nil {
		t.Fatalf("expected session at legacy path: %v", err)
	}
}

func restoreAuthConfig(t *testing.T) {
	t.Helper()
	originalSecret := util.GetConfigByKey("auth.jwt.secret")
	originalExpiration := util.GetConfigByKey("auth.jwt.expiration_minutes")
	t.Cleanup(func() {
		util.SetConfigByKey("auth.jwt.secret", originalSecret)
		util.SetConfigByKey("auth.jwt.expiration_minutes", originalExpiration)
	})
	util.SetConfigByKey("auth.jwt.secret", "test-secret")
	util.SetConfigByKey("auth.jwt.expiration_minutes", "30")
}

func installUserMapperStub(t *testing.T, users ...model.UserEntity) *userMapperStub {
	t.Helper()
	original := user_mapper.INSTANCE
	stub := &userMapperStub{users: map[string]model.UserEntity{}}
	for _, user := range users {
		stub.users[user.Id.Hex()] = user
	}
	user_mapper.INSTANCE = stub
	t.Cleanup(func() {
		user_mapper.INSTANCE = original
	})
	return stub
}

type userMapperStub struct {
	users map[string]model.UserEntity
}

func (stub *userMapperStub) GetUserByObjectId(plainId string) model.UserEntity {
	return stub.users[plainId]
}

func (stub *userMapperStub) GetUserByUsername(username string) model.UserEntity {
	for _, user := range stub.users {
		if user.Username == username {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userMapperStub) GetUserByUsernameIncludeDeleted(username string) model.UserEntity {
	return stub.GetUserByUsername(username)
}

func (stub *userMapperStub) GetUserByEmail(email string) model.UserEntity {
	for _, user := range stub.users {
		if user.EmailAddress == email {
			return user
		}
	}
	return model.UserEntity{}
}

func (stub *userMapperStub) InsertUserByEntity(newEntity model.UserEntity) string {
	stub.users[newEntity.Id.Hex()] = newEntity
	return newEntity.Id.Hex()
}

func (stub *userMapperStub) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	stub.users[plainId] = updatedEntity
	return updatedEntity
}

func (stub *userMapperStub) GetAllUsers(limit, offset int) []model.UserEntity {
	users := make([]model.UserEntity, 0, len(stub.users))
	for _, user := range stub.users {
		users = append(users, user)
	}
	return users
}

func (stub *userMapperStub) GetAllUsersIncludeDeleted(limit, offset int) []model.UserEntity {
	return stub.GetAllUsers(limit, offset)
}

func (stub *userMapperStub) GetUsersByRole(role string) []model.UserEntity {
	users := []model.UserEntity{}
	for _, user := range stub.users {
		if user.Role == role {
			users = append(users, user)
		}
	}
	return users
}

func (stub *userMapperStub) CountAllUsers() int64 {
	return int64(len(stub.users))
}

func (stub *userMapperStub) DeleteUserByObjectId(plainId string) model.UserEntity {
	user := stub.users[plainId]
	delete(stub.users, plainId)
	return user
}

func (stub *userMapperStub) TruncateUsers() error {
	stub.users = map[string]model.UserEntity{}
	return nil
}

func installRefreshTokenMapperStub(t *testing.T) *refreshTokenMapperStub {
	t.Helper()
	original := refresh_token_mapper.INSTANCE
	stub := &refreshTokenMapperStub{}
	refresh_token_mapper.INSTANCE = stub
	t.Cleanup(func() {
		refresh_token_mapper.INSTANCE = original
	})
	return stub
}

type refreshTokenMapperStub struct {
	token        model.RefreshToken
	createdToken string
	revokedToken string
}

func (stub *refreshTokenMapperStub) CreateToken(token model.RefreshToken) string {
	if stub.createdToken != "" {
		return stub.createdToken
	}
	return token.Token
}

func (stub *refreshTokenMapperStub) GetTokenByToken(token string) model.RefreshToken {
	if stub.token.Token == token {
		return stub.token
	}
	return model.RefreshToken{}
}

func (stub *refreshTokenMapperStub) GetTokensByUserId(userId string) []model.RefreshToken {
	return nil
}

func (stub *refreshTokenMapperStub) RevokeToken(token string, revokedBy string) error {
	stub.revokedToken = token
	return nil
}

func (stub *refreshTokenMapperStub) RevokeAllTokensByUserId(userId string) error {
	return nil
}
