package validation

import (
	"testing"
)

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"Valid YYYYMMDD", "20241205", false},
		{"Valid YYYY-MM-DD", "2024-12-05", false},
		{"Empty date", "", true},
		{"Valid YYYY/MM/DD", "2024/12/05", false},
		{"Invalid date", "20241301", true},
		{"Too short", "2024120", true},
		{"Too long", "202412055", true},
		{"Non-numeric", "abcd1205", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{"Valid range", "20241201", "20241205", false},
		{"Same date", "20241205", "20241205", false},
		{"Invalid range", "20241205", "20241201", true},
		{"Invalid from date", "invalid", "20241205", true},
		{"Invalid to date", "20241201", "invalid", true},
		{"Mixed formats", "20241201", "2024-12-05", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"Valid amount", 100.50, false},
		{"Zero amount", 0, true},
		{"Negative amount", -50.00, true},
		{"Very large amount", 1000000000.00, true},
		{"Small positive", 0.01, false},
		{"Maximum valid", 999999999.99, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"Valid ID", "507f1f77bcf86cd799439011", false},
		{"Empty ID", "", true},
		{"Too short", "507f1f77bcf86cd79943901", true},
		{"Too long", "507f1f77bcf86cd7994390111", true},
		{"Invalid characters", "507f1f77bcf86cd79943901g", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCategoryName(t *testing.T) {
	tests := []struct {
		name     string
		category string
		wantErr  bool
	}{
		{"Valid name", "Food & Dining", false},
		{"Valid with dash", "Food-Dining", false},
		{"Valid with underscore", "Food_Dining", false},
		{"Empty name", "", true},
		{"Too long", string(make([]byte, 101)), true},
		{"Invalid characters", "Food@Dining", true},
		{"Valid alphanumeric", "Category123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategoryName(tt.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategoryName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		desc    string
		wantErr bool
	}{
		{"Valid description", "Test description", false},
		{"Empty description", "", false},
		{"Too long", string(make([]byte, 501)), true},
		{"Maximum length", string(make([]byte, 500)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFlowType(t *testing.T) {
	tests := []struct {
		name     string
		flowType string
		wantErr  bool
	}{
		{"Valid income", "income", false},
		{"Valid expense", "expense", false},
		{"Invalid uppercase", "INCOME", true},
		{"Invalid type", "OUTCOME", true},
		{"Empty type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlowType(tt.flowType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFlowType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
