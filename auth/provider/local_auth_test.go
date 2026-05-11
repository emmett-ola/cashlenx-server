package provider

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/macar-x/cashlenx-server/util"
)

func TestGenerateTokenUsesConfiguredExpirationHours(t *testing.T) {
	originalSecret := util.GetConfigByKey("auth.jwt.secret")
	originalExpirationHours := util.GetConfigByKey("auth.jwt.expiration_hours")
	defer util.SetConfigByKey("auth.jwt.secret", originalSecret)
	defer util.SetConfigByKey("auth.jwt.expiration_hours", originalExpirationHours)

	util.SetConfigByKey("auth.jwt.secret", "test-secret")
	util.SetConfigByKey("auth.jwt.expiration_hours", "2")

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

	minExpiration := before.Add(2*time.Hour - time.Second)
	maxExpiration := time.Now().Add(2*time.Hour + time.Second)
	if claims.ExpiresAt.Time.Before(minExpiration) || claims.ExpiresAt.Time.After(maxExpiration) {
		t.Fatalf("ExpiresAt = %v, want around 2 hours from now between %v and %v", claims.ExpiresAt.Time, minExpiration, maxExpiration)
	}
}
