package refresh_token_service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/refresh_token_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

func TestCreateRefreshTokenStoresDigestAndReturnsOpaqueToken(t *testing.T) {
	mapper := installRefreshTokenServiceMapper(t)

	rawToken, err := CreateRefreshToken("507f1f77bcf86cd799439011", "device", "name", "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}
	if rawToken == "" || strings.HasPrefix(rawToken, "sha256:") {
		t.Fatalf("returned token has unexpected format: %q", rawToken)
	}
	if mapper.created.Token != refreshTokenDigest(rawToken) {
		t.Fatal("stored refresh credential is not the digest of the returned token")
	}
}

func TestGetRefreshTokenRejectsExpiredToken(t *testing.T) {
	mapper := installRefreshTokenServiceMapper(t)
	rawToken := "expired-token"
	mapper.tokens[refreshTokenDigest(rawToken)] = model.RefreshToken{
		Id:        "token-id",
		Token:     refreshTokenDigest(rawToken),
		ExpiresAt: time.Now().Add(-time.Minute),
	}

	if _, err := GetRefreshTokenByToken(rawToken, "", "", "", ""); err == nil {
		t.Fatal("expired refresh token was accepted")
	}
}

func TestRevokedRefreshTokenCannotBeConsumedAgain(t *testing.T) {
	installRefreshTokenServiceMapper(t)
	rawToken, err := CreateRefreshToken("507f1f77bcf86cd799439011", "", "", "", "")
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}
	if _, err := GetRefreshTokenByToken(rawToken, "", "", "", ""); err != nil {
		t.Fatalf("GetRefreshTokenByToken() before revoke error = %v", err)
	}
	if err := RevokeRefreshToken(rawToken, "507f1f77bcf86cd799439011"); err != nil {
		t.Fatalf("RevokeRefreshToken() error = %v", err)
	}
	if _, err := GetRefreshTokenByToken(rawToken, "", "", "", ""); err == nil {
		t.Fatal("revoked refresh token was accepted")
	}
	if err := RevokeRefreshToken(rawToken, "507f1f77bcf86cd799439011"); err == nil {
		t.Fatal("refresh token replay was accepted")
	}
}

func TestGetUserRefreshTokensRedactsStoredCredentials(t *testing.T) {
	mapper := installRefreshTokenServiceMapper(t)
	mapper.tokens["sha256:digest"] = model.RefreshToken{
		Id:        "token-id",
		UserId:    "user-id",
		Token:     "sha256:digest",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tokens := GetUserRefreshTokens("user-id")
	if len(tokens) != 1 || tokens[0].Token != "" {
		t.Fatalf("session inventory exposed refresh credential: %#v", tokens)
	}
}

func TestRefreshTokenExpirationDaysUsesConfiguredValue(t *testing.T) {
	originalExpirationDays := util.GetConfigByKey("auth.refresh_token.expiration_days")
	defer util.SetConfigByKey("auth.refresh_token.expiration_days", originalExpirationDays)

	util.SetConfigByKey("auth.refresh_token.expiration_days", "14")

	if got := refreshTokenExpirationDays(); got != 14 {
		t.Fatalf("refreshTokenExpirationDays() = %d, want 14", got)
	}
}

func TestRefreshTokenExpirationDaysDefaultsInvalidValue(t *testing.T) {
	originalExpirationDays := util.GetConfigByKey("auth.refresh_token.expiration_days")
	defer util.SetConfigByKey("auth.refresh_token.expiration_days", originalExpirationDays)

	util.SetConfigByKey("auth.refresh_token.expiration_days", "0")

	if got := refreshTokenExpirationDays(); got != 14 {
		t.Fatalf("refreshTokenExpirationDays() = %d, want 14", got)
	}
}

func installRefreshTokenServiceMapper(t *testing.T) *refreshTokenServiceMapperStub {
	t.Helper()
	original := refresh_token_mapper.INSTANCE
	stub := &refreshTokenServiceMapperStub{tokens: map[string]model.RefreshToken{}}
	refresh_token_mapper.INSTANCE = stub
	t.Cleanup(func() { refresh_token_mapper.INSTANCE = original })
	return stub
}

type refreshTokenServiceMapperStub struct {
	created model.RefreshToken
	tokens  map[string]model.RefreshToken
}

func (stub *refreshTokenServiceMapperStub) CreateToken(token model.RefreshToken) string {
	stub.created = token
	stub.tokens[token.Token] = token
	return token.Token
}

func (stub *refreshTokenServiceMapperStub) GetTokenByToken(token string) model.RefreshToken {
	return stub.tokens[token]
}

func (stub *refreshTokenServiceMapperStub) GetTokensByUserId(userID string) []model.RefreshToken {
	tokens := make([]model.RefreshToken, 0)
	for _, token := range stub.tokens {
		if token.UserId == userID {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (stub *refreshTokenServiceMapperStub) RevokeToken(token, revokedBy string) error {
	stored, ok := stub.tokens[token]
	if !ok || stored.RevokedAt != nil || !stored.ExpiresAt.After(time.Now()) {
		return errors.New("active refresh token not found")
	}
	now := time.Now()
	stored.RevokedAt = &now
	stored.RevokedBy = revokedBy
	stub.tokens[token] = stored
	return nil
}

func (stub *refreshTokenServiceMapperStub) RevokeAllTokensByUserId(userID string) error {
	for key, token := range stub.tokens {
		if token.UserId == userID && token.RevokedAt == nil {
			now := time.Now()
			token.RevokedAt = &now
			stub.tokens[key] = token
		}
	}
	return nil
}
