package cash_flow_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

func TestQueryById(t *testing.T) {
	// Test only the error handling logic that can be verified without mocking
	// The actual database operations require integration testing

	// This test will fail if the function doesn't return an AppError for invalid IDs
	result, err := QueryById("invalid-id")
	if err == nil {
		t.Error("Expected error for invalid ID but got none")
	}

	// Verify it's an AppError
	if _, ok := err.(*errors.AppError); !ok {
		t.Error("Expected AppError for invalid ID but got different error type")
	}

	// Verify the result is empty
	if !result.IsEmpty() {
		t.Errorf("Expected empty result for invalid ID, got %+v", result)
	}
}

func TestQueryByDate(t *testing.T) {
	// Test with invalid date format to verify error handling
	result, err := QueryByDate("invalid-date")
	if err == nil {
		t.Error("Expected error for invalid date format but got none")
	}

	// Verify it's an AppError
	if _, ok := err.(*errors.AppError); !ok {
		t.Error("Expected AppError for invalid date format but got different error type")
	}

	// Verify the result is empty
	if len(result) > 0 {
		t.Errorf("Expected empty result for invalid date format, got %+v", result)
	}

	if util.GetConfigByKey("db.mongodb.url") == "" && util.GetConfigByKey("db.mysql.url") == "" {
		t.Skip("database not configured")
	}

	result, err = QueryByDate("20230101")
	if err != nil {
		if _, ok := err.(*errors.AppError); !ok {
			t.Errorf("Unexpected error type for valid date: %v", err)
		}
	}
}

func TestIsQueryFieldsConflicted(t *testing.T) {
	tests := []struct {
		name             string
		plainId          string
		belongsDate      string
		exactDescription string
		fuzzyDescription string
		expectConflicted bool
	}{
		{
			name:             "No fields filled",
			expectConflicted: true,
		},
		{
			name:             "Only plainId filled",
			plainId:          "123",
			expectConflicted: false,
		},
		{
			name:             "Only belongsDate filled",
			belongsDate:      "20230101",
			expectConflicted: false,
		},
		{
			name:             "Only exactDescription filled",
			exactDescription: "test",
			expectConflicted: false,
		},
		{
			name:             "Only fuzzyDescription filled",
			fuzzyDescription: "test",
			expectConflicted: false,
		},
		{
			name:             "Two fields filled",
			plainId:          "123",
			belongsDate:      "20230101",
			expectConflicted: true,
		},
		{
			name:             "Three fields filled",
			plainId:          "123",
			belongsDate:      "20230101",
			exactDescription: "test",
			expectConflicted: true,
		},
		{
			name:             "All fields filled",
			plainId:          "123",
			belongsDate:      "20230101",
			exactDescription: "test",
			fuzzyDescription: "test",
			expectConflicted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsQueryFieldsConflicted(tt.plainId, tt.belongsDate, tt.exactDescription, tt.fuzzyDescription)
			if result != tt.expectConflicted {
				t.Errorf("Expected conflicted=%v, got %v", tt.expectConflicted, result)
			}
		})
	}
}
