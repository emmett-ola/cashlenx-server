package util

import (
	"testing"
	"time"
)

func TestLoadTimezone_UTC(t *testing.T) {
	// Save original timezone
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	// Test with UTC
	SetConfigByKey("timezone", "UTC")
	loadTimezone()
	if timezone != time.UTC {
		t.Errorf("Expected UTC timezone, got %v", timezone)
	}
}

func TestLoadTimezone_IANAName(t *testing.T) {
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	SetConfigByKey("timezone", "Asia/Shanghai")
	loadTimezone()

	testUTC := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := testUTC.In(timezone).Hour(); got != 20 {
		t.Fatalf("Asia/Shanghai hour = %d, want 20", got)
	}
}

func TestValidateConfiguredTimezone(t *testing.T) {
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	for _, name := range []string{"UTC", "Asia/Shanghai", "America/New_York"} {
		t.Run("accept_"+name, func(t *testing.T) {
			SetConfigByKey("timezone", name)
			if err := ValidateConfiguredTimezone(); err != nil {
				t.Fatalf("ValidateConfiguredTimezone() error = %v", err)
			}
		})
	}

	for _, name := range []string{"UTC+8", "UTC-5:30", "GMT+8", "CST", "Etc/GMT+8", "Mars/Olympus"} {
		t.Run("reject_"+name, func(t *testing.T) {
			SetConfigByKey("timezone", name)
			if err := ValidateConfiguredTimezone(); err == nil {
				t.Fatal("ValidateConfiguredTimezone() error = nil")
			}
		})
	}
}

func TestLoadTimezone_InvalidFallsBackToUTC(t *testing.T) {
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	SetConfigByKey("timezone", "UTC+8")
	loadTimezone()
	if timezone != time.UTC {
		t.Errorf("Expected invalid timezone to fall back to UTC, got %v", timezone)
	}
}

// Use the existing SetConfigByKey function from config_util.go
