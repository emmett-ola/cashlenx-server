package util

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	defaultDateFormatInString  = "20060102"
	dateFormatInStringWithDash = "2006-01-02"
	timezone                   *time.Location
	timezoneNamePattern        = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._+-]*(/[A-Za-z][A-Za-z0-9._+-]*)+$`)
)

// init is called after all other package initialization
// We don't load timezone here to avoid initialization order issues
// Instead, we ensure timezone is loaded when first used

var errInvalidTimezone = errors.New("must be UTC or a region-based IANA timezone name")

func configuredTimezoneName() string {
	tzName := GetConfigByKey("timezone")
	if tzName == "" {
		return "UTC"
	}
	return tzName
}

func parseTimezone(tzName string) (*time.Location, error) {
	if tzName == "UTC" {
		return time.UTC, nil
	}

	if !timezoneNamePattern.MatchString(tzName) || strings.HasPrefix(tzName, "Etc/GMT") {
		return nil, errInvalidTimezone
	}

	location, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, errInvalidTimezone
	}
	return location, nil
}

// ValidateConfiguredTimezone verifies the shared application/container contract.
func ValidateConfiguredTimezone() error {
	_, err := parseTimezone(configuredTimezoneName())
	return err
}

// loadTimezone loads the configured timezone and falls back to UTC for callers
// that do not enter through the validated API start command.
func loadTimezone() {
	location, err := parseTimezone(configuredTimezoneName())
	if err != nil {
		Logger.Errorw("Invalid TIMEZONE; using UTC instead", "key", "TIMEZONE", "error", err)
		timezone = time.UTC
		return
	}

	timezone = location
	Logger.Infow("Loaded timezone", "timezone", timezone.String())
}

// GetTimezone returns the configured timezone
func GetTimezone() *time.Location {
	// Lazy load timezone if not already loaded
	if timezone == nil {
		loadTimezone()
	}
	return timezone
}

// ToUTC converts time to UTC for storage
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ToTimezone converts time to the configured timezone for display
func ToTimezone(t time.Time) time.Time {
	// Lazy load timezone if not already loaded
	if timezone == nil {
		loadTimezone()
	}
	return t.In(timezone)
}

func FormatDateFromStringWithoutDash(dateString string) time.Time {
	// Try to parse with the specified format first
	date, err := time.Parse(defaultDateFormatInString, dateString)
	if err == nil {
		return date
	}

	// Try other formats
	return parseAnyDateFormat(dateString)
}

func FormatDateFromStringWithDash(dateString string) time.Time {
	// Try to parse with the specified format first
	date, err := time.Parse(dateFormatInStringWithDash, dateString)
	if err == nil {
		return date
	}

	// Try other formats
	return parseAnyDateFormat(dateString)
}

// parseAnyDateFormat tries to parse a date string in any supported format
func parseAnyDateFormat(dateString string) time.Time {
	// Try YYYYMMDD
	date, err := time.Parse(defaultDateFormatInString, dateString)
	if err == nil {
		return date
	}

	// Try YYYY-MM-DD
	date, err = time.Parse(dateFormatInStringWithDash, dateString)
	if err == nil {
		return date
	}

	// Try YYYY/MM/DD
	date, err = time.Parse("2006/01/02", dateString)
	if err == nil {
		return date
	}

	Logger.Errorw("Failed to parse date", "date", dateString, "error", err)
	return time.Time{}
}

func formatDateFromString(dateString, format string) time.Time {
	date, err := time.Parse(format, dateString)
	if err != nil {
		Logger.Errorln(err)
		// Try other formats if the specified one fails
		return parseAnyDateFormat(dateString)
	}
	return date
}

func FormatDateToStringWithoutDash(date time.Time) string {
	return formatDateToString(date, defaultDateFormatInString)
}

func FormatDateToStringWithDash(date time.Time) string {
	return formatDateToString(date, dateFormatInStringWithDash)
}

func formatDateToString(date time.Time, format string) string {
	return date.Format(format)
}

func IsDateTimeEmpty(dateTime time.Time) bool {
	return reflect.DeepEqual(dateTime, time.Time{})
}

// ParseDate parses a date string in YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD format and returns a time.Time
// Returns an error if the date string is invalid
func ParseDate(dateStr string) (time.Time, error) {
	// Try parsing without dash first (YYYYMMDD)
	date, err := time.Parse(defaultDateFormatInString, dateStr)
	if err == nil {
		return date, nil
	}

	// Try parsing with dash (YYYY-MM-DD)
	date, err = time.Parse(dateFormatInStringWithDash, dateStr)
	if err == nil {
		return date, nil
	}

	// Try parsing with slash (YYYY/MM/DD)
	date, err = time.Parse("2006/01/02", dateStr)
	if err == nil {
		return date, nil
	}

	// All formats failed
	return time.Time{}, err
}

// GetCurrentTime returns the current time in UTC for storage purposes
func GetCurrentTime() time.Time {
	return time.Now().UTC()
}
