package validation

import (
	"regexp"
	"time"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return errors.NewFieldValidationError(field, message)
}

// ValidateDate validates date string format (YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD)
func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return NewValidationError("date", "cannot be empty")
	}

	// Check format YYYYMMDD (8 digits)
	if matched, _ := regexp.MatchString(`^\d{8}$`, dateStr); matched {
		// Parse to verify it's a valid date
		_, err := time.Parse("20060102", dateStr)
		if err != nil {
			return NewValidationError("date", "invalid date format, use YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD")
		}
		return nil
	}

	// Check format YYYY-MM-DD
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); matched {
		_, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return NewValidationError("date", "invalid date format, use YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD")
		}
		return nil
	}

	// Check format YYYY/MM/DD
	if matched, _ := regexp.MatchString(`^\d{4}/\d{2}/\d{2}$`, dateStr); matched {
		_, err := time.Parse("2006/01/02", dateStr)
		if err != nil {
			return NewValidationError("date", "invalid date format, use YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD")
		}
		return nil
	}

	return NewValidationError("date", "invalid date format, use YYYYMMDD, YYYY-MM-DD, or YYYY/MM/DD")
}

// ValidateDateRange validates a date range (from must be before or equal to to)
func ValidateDateRange(fromStr, toStr string) error {
	if err := ValidateDate(fromStr); err != nil {
		return err
	}

	if err := ValidateDate(toStr); err != nil {
		return err
	}

	// Parse dates for comparison
	var from, to time.Time
	var err error

	// Parse from date
	if len(fromStr) == 8 {
		from, err = time.Parse("20060102", fromStr)
	} else if matched, _ := regexp.MatchString(`^\d{4}/\d{2}/\d{2}$`, fromStr); matched {
		from, err = time.Parse("2006/01/02", fromStr)
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
	}
	if err != nil {
		return NewValidationError("from_date", "failed to parse date")
	}

	// Parse to date
	if len(toStr) == 8 {
		to, err = time.Parse("20060102", toStr)
	} else if matched, _ := regexp.MatchString(`^\d{4}/\d{2}/\d{2}$`, toStr); matched {
		to, err = time.Parse("2006/01/02", toStr)
	} else {
		to, err = time.Parse("2006-01-02", toStr)
	}
	if err != nil {
		return NewValidationError("to_date", "failed to parse date")
	}

	if from.After(to) {
		return NewValidationError("date_range", "from date must be before or equal to to date")
	}

	return nil
}

// ValidateAmount validates monetary amount (must be positive)
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return NewValidationError("amount", "must be positive")
	}

	if amount > 999999999.99 {
		return NewValidationError("amount", "exceeds maximum allowed value")
	}

	return nil
}

// ValidateID validates ObjectID format
func ValidateID(id string) error {
	if id == "" {
		return NewValidationError("id", "cannot be empty")
	}

	if len(id) != 24 {
		return NewValidationError("id", "invalid ID format (must be 24 characters)")
	}

	_, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return NewValidationError("id", "invalid ID format")
	}

	return nil
}

// ValidateCategoryName validates category name
func ValidateCategoryName(name string) error {
	if name == "" {
		return NewValidationError("category", "cannot be empty")
	}

	if len(name) > 100 {
		return NewValidationError("category", "name too long (max 100 characters)")
	}

	// Check for valid characters (alphanumeric, spaces, and common punctuation)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_&]+$`, name); !matched {
		return NewValidationError("category", "contains invalid characters")
	}

	return nil
}

// ValidateDescription validates description text
func ValidateDescription(desc string) error {
	if len(desc) > 500 {
		return NewValidationError("description", "too long (max 500 characters)")
	}

	return nil
}

// ValidateFlowType validates flow type (income or expense)
func ValidateFlowType(flowType string) error {
	if flowType != model.FlowTypeIncome && flowType != model.FlowTypeExpense {
		return NewValidationError("flow_type", "must be income or expense")
	}

	return nil
}

// ValidateRequired validates that a string field is not empty
func ValidateRequired(field, value string) error {
	if value == "" {
		return NewValidationError(field, "is required")
	}

	return nil
}

// ValidatePassword validates password requirements
func ValidatePassword(password string) error {
	if password == "" {
		return NewValidationError("password", "is required")
	}

	if len(password) < 6 {
		return NewValidationError("password", "must be 6 characters or more")
	}

	if len(password) > 100 {
		return NewValidationError("password", "must be 100 characters or less")
	}

	return nil
}
