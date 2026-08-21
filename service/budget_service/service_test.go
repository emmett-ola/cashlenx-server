package budget_service

import (
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateRequest(t *testing.T) {
	categoryID := primitive.NewObjectID().Hex()
	tests := []struct {
		name      string
		request   model.UpsertBudgetRequest
		wantError bool
	}{
		{"valid", model.UpsertBudgetRequest{CategoryId: categoryID, Period: "2026-08", LimitAmount: 1200}, false},
		{"bad category", model.UpsertBudgetRequest{CategoryId: "bad", Period: "2026-08", LimitAmount: 1200}, true},
		{"bad period", model.UpsertBudgetRequest{CategoryId: categoryID, Period: "2026-13", LimitAmount: 1200}, true},
		{"zero limit", model.UpsertBudgetRequest{CategoryId: categoryID, Period: "2026-08", LimitAmount: 0}, true},
		{"negative limit", model.UpsertBudgetRequest{CategoryId: categoryID, Period: "2026-08", LimitAmount: -1}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRequest(tc.request)
			if (err != nil) != tc.wantError {
				t.Fatalf("validateRequest() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}
