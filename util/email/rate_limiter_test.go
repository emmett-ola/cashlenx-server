package email

import (
	"strconv"
	"testing"

	appErrors "github.com/macar-x/cashlenx-server/errors"
)

func TestCheckAndRecordPurposeEmailAllowanceLimitsByIPAndPurpose(t *testing.T) {
	resetEmailRateLimiterForTest(t)

	for i := 0; i < 5; i++ {
		if err := CheckAndRecordPurposeEmailAllowance("password_reset", "203.0.113.10", []string{"user@example.test"}); err != nil {
			t.Fatalf("attempt %d returned error: %v", i+1, err)
		}
	}

	err := CheckAndRecordPurposeEmailAllowance("password_reset", "203.0.113.10", []string{"other@example.test"})
	if !appErrors.IsRateLimitedError(err) {
		t.Fatalf("error = %v, want rate limited", err)
	}

	if err := CheckAndRecordPurposeEmailAllowance("email_change", "203.0.113.10", []string{"other@example.test"}); err != nil {
		t.Fatalf("different purpose should not be rate limited: %v", err)
	}
}

func TestCheckAndRecordPurposeEmailAllowanceLimitsByRecipientAndPurpose(t *testing.T) {
	resetEmailRateLimiterForTest(t)

	for i := 0; i < 5; i++ {
		ip := "203.0.113." + strconv.Itoa(i+1)
		if err := CheckAndRecordPurposeEmailAllowance("email_change", ip, []string{"USER@example.test"}); err != nil {
			t.Fatalf("attempt %d returned error: %v", i+1, err)
		}
	}

	err := CheckAndRecordPurposeEmailAllowance("email_change", "203.0.113.99", []string{"user@example.test"})
	if !appErrors.IsRateLimitedError(err) {
		t.Fatalf("error = %v, want rate limited", err)
	}
}

func resetEmailRateLimiterForTest(t *testing.T) {
	t.Helper()

	limiter.mu.Lock()
	originalByIP := limiter.byIP
	originalByEmail := limiter.byEmail
	limiter.byIP = map[string]dailyRateLimitEntry{}
	limiter.byEmail = map[string]dailyRateLimitEntry{}
	limiter.mu.Unlock()
	originalGetDailyPerIPLimit := getDailyPerIPLimit
	originalGetDailyPerEmailLimit := getDailyPerEmailLimit
	getDailyPerIPLimit = func() int { return 5 }
	getDailyPerEmailLimit = func() int { return 5 }

	t.Cleanup(func() {
		limiter.mu.Lock()
		limiter.byIP = originalByIP
		limiter.byEmail = originalByEmail
		limiter.mu.Unlock()
		getDailyPerIPLimit = originalGetDailyPerIPLimit
		getDailyPerEmailLimit = originalGetDailyPerEmailLimit
	})
}
