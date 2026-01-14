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

func TestLoadTimezone_UTCOffsetHours(t *testing.T) {
	// Save original timezone
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	// Create a test UTC time
	testUTC := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test with UTC+1
	SetConfigByKey("timezone", "UTC+1")
	loadTimezone()
	localTime := testUTC.In(timezone)
	expectedHour := 13 // UTC+1 should be 13:00
	if localTime.Hour() != expectedHour {
		t.Errorf("Expected UTC+1 to be 13:00, got %d:00", localTime.Hour())
	}

	// Test with UTC-5
	SetConfigByKey("timezone", "UTC-5")
	loadTimezone()
	localTime = testUTC.In(timezone)
	expectedHour = 7 // UTC-5 should be 07:00
	if localTime.Hour() != expectedHour {
		t.Errorf("Expected UTC-5 to be 07:00, got %d:00", localTime.Hour())
	}

	// Test with UTC+0
	SetConfigByKey("timezone", "UTC+0")
	loadTimezone()
	localTime = testUTC.In(timezone)
	expectedHour = 12 // UTC+0 should be 12:00
	if localTime.Hour() != expectedHour {
		t.Errorf("Expected UTC+0 to be 12:00, got %d:00", localTime.Hour())
	}
}

func TestLoadTimezone_UTCOffsetMinutes(t *testing.T) {
	// Save original timezone
	originalTz := GetConfigByKey("timezone")
	defer SetConfigByKey("timezone", originalTz)

	// Create a test UTC time
	testUTC := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test with UTC+1:30
	SetConfigByKey("timezone", "UTC+1:30")
	loadTimezone()
	localTime := testUTC.In(timezone)
	expectedHour := 13
	expectedMinute := 30 // UTC+1:30 should be 13:30
	if localTime.Hour() != expectedHour || localTime.Minute() != expectedMinute {
		t.Errorf("Expected UTC+1:30 to be 13:30, got %d:%d", localTime.Hour(), localTime.Minute())
	}

	// Test with UTC-5:30
	SetConfigByKey("timezone", "UTC-5:30")
	loadTimezone()
	localTime = testUTC.In(timezone)
	expectedHour = 6
	expectedMinute = 30 // UTC-5:30 should be 06:30
	if localTime.Hour() != expectedHour || localTime.Minute() != expectedMinute {
		t.Errorf("Expected UTC-5:30 to be 06:30, got %d:%d", localTime.Hour(), localTime.Minute())
	}
}

// Use the existing SetConfigByKey function from config_util.go
