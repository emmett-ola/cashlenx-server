package auth_controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/mapper/refresh_token_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLogoutWithoutTokenReturnsOK(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/logout", nil)
	req = req.WithContext(util.ContextWithRequestID(req.Context(), "request-1"))
	rec := httptest.NewRecorder()

	Logout(rec, req)

	assertOKResponse(t, rec)
	if mapper.revokedToken != "" || mapper.revokedAllUserID != "" {
		t.Fatalf("expected no revocation, got revokedToken=%q revokedAllUserID=%q", mapper.revokedToken, mapper.revokedAllUserID)
	}
}

func TestLoginRejectsInvalidRequestBodies(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{"username":`},
		{name: "missing credentials", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			Login(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestRegisterRejectsInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/register", bytes.NewBufferString(`{"username":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterAcceptsEmailFieldName(t *testing.T) {
	originalRegistrationEnabled := util.GetConfigByKey("auth.registration.enabled")
	defer util.SetConfigByKey("auth.registration.enabled", originalRegistrationEnabled)
	util.SetConfigByKey("auth.registration.enabled", "false")

	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/register", bytes.NewBufferString(`{"username":"alice","password":"secret123","email":"alice@example.com","verification_token":"token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Register(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestRegisterDelegatesVerifiedRegistrationAndReturnsCreatedUser(t *testing.T) {
	userID := primitive.NewObjectID()
	var username, password, emailAddress, verificationToken string
	originalRegister := registerPublicUser
	originalGet := getRegisteredUser
	registerPublicUser = func(gotUsername, gotPassword, gotEmail, gotToken string) (string, error) {
		username, password, emailAddress, verificationToken = gotUsername, gotPassword, gotEmail, gotToken
		return userID.Hex(), nil
	}
	getRegisteredUser = func(id string) model.UserEntity {
		if id != userID.Hex() {
			t.Fatalf("created user id = %q, want %q", id, userID.Hex())
		}
		return model.UserEntity{Id: userID, Username: "alice", EmailAddress: "alice@example.test", IsEmailVerified: true}
	}
	t.Cleanup(func() {
		registerPublicUser = originalRegister
		getRegisteredUser = originalGet
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/register", bytes.NewBufferString(`{"username":"alice","password":"StrongPass123!","email":"alice@example.test","verification_token":"signup-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if username != "alice" || password != "StrongPass123!" || emailAddress != "alice@example.test" || verificationToken != "signup-token" {
		t.Fatalf("registration args = %q/%q/%q/%q", username, password, emailAddress, verificationToken)
	}
}

func TestGetTokensRequiresUserContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/tokens", nil)
	rec := httptest.NewRecorder()

	GetTokens(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGetTokensReturnsUserTokens(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)
	mapper.tokens["refresh-token"] = model.RefreshToken{
		Id:        "token-id",
		UserId:    "user-id",
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v0/auth/tokens", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-id"))
	rec := httptest.NewRecorder()

	GetTokens(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUserIDFromAuthorizationHeaderRejectsMalformedHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/logout", nil)
	req.Header.Set("Authorization", "Basic token")

	if got := userIDFromAuthorizationHeader(req); got != "" {
		t.Fatalf("userIDFromAuthorizationHeader() = %q, want empty", got)
	}
}

func TestLogoutWithInvalidRefreshTokenReturnsOK(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)
	req := logoutRequest(`{"refresh_token":"missing-token"}`)
	rec := httptest.NewRecorder()

	Logout(rec, req)

	assertOKResponse(t, rec)
	if mapper.revokedToken != "" || mapper.revokedAllUserID != "" {
		t.Fatalf("expected no revocation, got revokedToken=%q revokedAllUserID=%q", mapper.revokedToken, mapper.revokedAllUserID)
	}
}

func TestLogoutWithRefreshTokenRevokesSingleSession(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)
	mapper.tokens["refresh-token"] = model.RefreshToken{
		Id:        "token-id",
		UserId:    "user-id",
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	req := logoutRequest(`{"refresh_token":"refresh-token"}`)
	rec := httptest.NewRecorder()

	Logout(rec, req)

	assertOKResponse(t, rec)
	if mapper.revokedToken != "refresh-token" {
		t.Fatalf("revokedToken = %q, want refresh-token", mapper.revokedToken)
	}
	if mapper.revokedBy != "user-id" {
		t.Fatalf("revokedBy = %q, want user-id", mapper.revokedBy)
	}
	if mapper.revokedAllUserID != "" {
		t.Fatalf("expected no revoke-all, got %q", mapper.revokedAllUserID)
	}
}

func TestLogoutWithInvalidBearerTokenReturnsOK(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)
	req := logoutRequest("")
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	Logout(rec, req)

	assertOKResponse(t, rec)
	if mapper.revokedToken != "" || mapper.revokedAllUserID != "" {
		t.Fatalf("expected no revocation, got revokedToken=%q revokedAllUserID=%q", mapper.revokedToken, mapper.revokedAllUserID)
	}
}

func TestLogoutWithBearerTokenRevokesAllSessions(t *testing.T) {
	mapper := installRefreshTokenMapperStub(t)

	originalSecret := util.GetConfigByKey("auth.jwt.secret")
	originalExpirationMinutes := util.GetConfigByKey("auth.jwt.expiration_minutes")
	defer util.SetConfigByKey("auth.jwt.secret", originalSecret)
	defer util.SetConfigByKey("auth.jwt.expiration_minutes", originalExpirationMinutes)
	util.SetConfigByKey("auth.jwt.secret", "test-secret")
	util.SetConfigByKey("auth.jwt.expiration_minutes", "30")

	token, err := auth.Service.GenerateToken("user-id", "tester", "user")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	req := logoutRequest("")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	Logout(rec, req)

	assertOKResponse(t, rec)
	if mapper.revokedAllUserID != "user-id" {
		t.Fatalf("revokedAllUserID = %q, want user-id", mapper.revokedAllUserID)
	}
	if mapper.revokedToken != "" {
		t.Fatalf("expected no single-token revoke, got %q", mapper.revokedToken)
	}
}

func logoutRequest(body string) *http.Request {
	var payload *bytes.Reader
	if body == "" {
		payload = bytes.NewReader(nil)
	} else {
		payload = bytes.NewReader([]byte(body))
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/logout", payload)
	req = req.WithContext(util.ContextWithRequestID(req.Context(), "request-1"))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func assertOKResponse(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body util.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("code = %q, want OK", body.Code)
	}
	if body.Message != "" {
		t.Fatalf("message = %q, want empty wrapper message", body.Message)
	}
}

func installRefreshTokenMapperStub(t *testing.T) *refreshTokenMapperStub {
	t.Helper()

	original := refresh_token_mapper.INSTANCE
	stub := &refreshTokenMapperStub{tokens: map[string]model.RefreshToken{}}
	refresh_token_mapper.INSTANCE = stub
	t.Cleanup(func() {
		refresh_token_mapper.INSTANCE = original
	})
	return stub
}

type refreshTokenMapperStub struct {
	tokens           map[string]model.RefreshToken
	revokedToken     string
	revokedBy        string
	revokedAllUserID string
	revokeErr        error
	revokeAllErr     error
}

func (stub *refreshTokenMapperStub) CreateToken(token model.RefreshToken) string {
	stub.tokens[token.Token] = token
	return token.Token
}

func (stub *refreshTokenMapperStub) GetTokenByToken(token string) model.RefreshToken {
	return stub.tokens[token]
}

func (stub *refreshTokenMapperStub) GetTokensByUserId(userId string) []model.RefreshToken {
	tokens := []model.RefreshToken{}
	for _, token := range stub.tokens {
		if token.UserId == userId {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (stub *refreshTokenMapperStub) RevokeToken(token string, revokedBy string) error {
	stub.revokedToken = token
	stub.revokedBy = revokedBy
	if stub.revokeErr != nil {
		return stub.revokeErr
	}
	if _, ok := stub.tokens[token]; !ok {
		return errors.New("token not found")
	}
	return nil
}

func (stub *refreshTokenMapperStub) RevokeAllTokensByUserId(userId string) error {
	stub.revokedAllUserID = userId
	return stub.revokeAllErr
}
