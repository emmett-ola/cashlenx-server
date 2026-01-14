package util

import (
	"fmt"
	"reflect"
	"time"
)

var (
	defaultDateFormatInString  = "20060102"
	dateFormatInStringWithDash = "2006-01-02"
	timezone                   *time.Location
)

// init is called after all other package initialization
// We don't load timezone here to avoid initialization order issues
// Instead, we ensure timezone is loaded when first used

// loadTimezone loads the timezone from configuration
func loadTimezone() {
	tzName := GetConfigByKey("timezone")
	if tzName == "" {
		tzName = "UTC" // Default to UTC
	}

	var err error

	// Try to parse as UTC offset first (e.g., UTC+1, UTC-5:30, UTC+0)
	if tzName == "UTC" {
		// Special case for UTC
		timezone = time.UTC
	} else if len(tzName) > 3 && tzName[:3] == "UTC" {
		// Handle UTC offset format like UTC+1, UTC-5:30, UTC+0
		offsetStr := tzName[3:]
		var offsetSeconds int

		// Parse offset string (e.g., "+1", "-5:30", "+0")
		if len(offsetStr) > 0 {
			// Determine sign
			sign := 1
			if offsetStr[0] == '-' {
				sign = -1
				offsetStr = offsetStr[1:]
			} else if offsetStr[0] == '+' {
				offsetStr = offsetStr[1:]
			}

			// Split into hours and minutes if colon exists
			hours := 0
			minutes := 0

			colonIndex := -1
			for i, c := range offsetStr {
				if c == ':' {
					colonIndex = i
					break
				}
			}

			if colonIndex != -1 {
				// Has minutes component (e.g., "1:30", "5:30")
				hoursStr := offsetStr[:colonIndex]
				minutesStr := offsetStr[colonIndex+1:]

				// Parse hours
				if _, err := fmt.Sscanf(hoursStr, "%d", &hours); err != nil {
					hours = 0
				}

				// Parse minutes
				if _, err := fmt.Sscanf(minutesStr, "%d", &minutes); err != nil {
					minutes = 0
				}
			} else {
				// Only hours component (e.g., "1", "5", "0")
				if _, err := fmt.Sscanf(offsetStr, "%d", &hours); err != nil {
					hours = 0
				}
			}

			// Calculate total seconds
			offsetSeconds = sign * (hours*3600 + minutes*60)

			// Create timezone from offset
			timezone = time.FixedZone(tzName, offsetSeconds)
			Logger.Infow("Loaded UTC offset timezone", "timezone", tzName, "offset_seconds", offsetSeconds)
			return
		}
	}

	// If not a UTC offset or parsing failed, try as named timezone
	timezone, err = time.LoadLocation(tzName)
	if err != nil {
		Logger.Errorw("Failed to load timezone, using UTC instead", "timezone", tzName, "error", err)
		timezone = time.UTC
	} else {
		Logger.Infow("Loaded named timezone", "timezone", tzName)
	}
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
