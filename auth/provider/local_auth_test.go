package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/macar-x/cashlenx-server/util"
)

func TestGenerateTokenUsesConfiguredExpirationMinutes(t *testing.T) {
	originalSecret := util.GetConfigByKey("auth.jwt.secret")
	originalExpirationMinutes := util.GetConfigByKey("auth.jwt.expiration_minutes")
	defer util.SetConfigByKey("auth.jwt.secret", originalSecret)
	defer util.SetConfigByKey("auth.jwt.expiration_minutes", originalExpirationMinutes)

	util.SetConfigByKey("auth.jwt.secret", "test-secret")
	util.SetConfigByKey("auth.jwt.expiration_minutes", "30")

	before := time.Now()
	tokenString, err := NewLocalAuthService().GenerateToken("user-id", "tester", "user")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims returned error: %v", err)
	}
	if !token.Valid {
		t.Fatal("token is not valid")
	}

	minExpiration := before.Add(30*time.Minute - time.Second)
	maxExpiration := time.Now().Add(30*time.Minute + time.Second)
	if claims.ExpiresAt.Time.Before(minExpiration) || claims.ExpiresAt.Time.After(maxExpiration) {
		t.Fatalf("ExpiresAt = %v, want around 30 minutes from now between %v and %v", claims.ExpiresAt.Time, minExpiration, maxExpiration)
	}
}

func TestMiddlewareTreatsOpenLogoutAsPublic(t *testing.T) {
	originalAPIVersion := util.GetConfigByKey("api.version")
	defer util.SetConfigByKey("api.version", originalAPIVersion)
	util.SetConfigByKey("api.version", "v0")

	called := false
	handler := NewLocalAuthService().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/logout", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected open logout request to reach next handler without auth")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
