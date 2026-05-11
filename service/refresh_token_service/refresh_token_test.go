package refresh_token_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/util"
)

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
