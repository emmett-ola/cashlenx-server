package cash_flow_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCashRootRequiresSubcommand(t *testing.T) {
	if err := CashCmd.RunE(CashCmd, nil); err == nil {
		t.Fatal("expected subcommand error")
	}
}

func TestCashCommandsRequireLoggedInUser(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "expense", run: func() error { return expenseCmd.RunE(expenseCmd, nil) }},
		{name: "income", run: func() error { return incomeCmd.RunE(incomeCmd, nil) }},
		{name: "list", run: func() error { return listCmd.RunE(listCmd, nil) }},
		{name: "query", run: func() error { return queryCmd.RunE(queryCmd, nil) }},
		{name: "range", run: func() error { return rangeCmd.RunE(rangeCmd, nil) }},
		{name: "summary", run: func() error { return summaryCmd.RunE(summaryCmd, nil) }},
		{name: "update", run: func() error { return updateCmd.RunE(updateCmd, nil) }},
		{name: "delete", run: func() error { return deleteCmd.RunE(deleteCmd, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cashUserId = ""
			if err := tt.run(); err == nil {
				t.Fatal("expected missing CLI session error")
			}
		})
	}
}

func TestExpenseCommandPassesFlagsAndAuthenticatedUserToService(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	var got struct {
		belongsDate  string
		categoryName string
		amount       float64
		description  string
		userID       string
	}

	originalRequire := requireCashUser
	originalSave := saveExpenseForUser
	requireCashUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	saveExpenseForUser = func(belongsDate, categoryName string, amount float64, description string, userID string) (model.CashFlowEntity, error) {
		got.belongsDate = belongsDate
		got.categoryName = categoryName
		got.amount = amount
		got.description = description
		got.userID = userID
		return model.CashFlowEntity{
			Id:           primitive.NewObjectID(),
			BelongsDate:  time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
			CategoryName: categoryName,
			CategoryType: model.FlowTypeExpense,
			Amount:       amount,
			Description:  description,
		}, nil
	}
	t.Cleanup(func() {
		requireCashUser = originalRequire
		saveExpenseForUser = originalSave
		cashUserId = ""
		_ = expenseCmd.Flags().Set("date", "")
		_ = expenseCmd.Flags().Set("category", "")
		_ = expenseCmd.Flags().Set("amount", "0")
		_ = expenseCmd.Flags().Set("description", "")
	})

	_ = expenseCmd.Flags().Set("date", "2026-05-25")
	_ = expenseCmd.Flags().Set("category", "Food")
	_ = expenseCmd.Flags().Set("amount", "12.34")
	_ = expenseCmd.Flags().Set("description", "lunch")

	if err := expenseCmd.RunE(expenseCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
	if got.userID != authenticatedUserID || got.belongsDate != "2026-05-25" || got.categoryName != "Food" || got.amount != 12.34 || got.description != "lunch" {
		t.Fatalf("service args = %+v", got)
	}
}

func TestListCommandUsesPageToCalculateOffset(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	var got struct {
		userID string
		limit  int
		offset int
	}

	originalRequire := requireCashUser
	originalQuery := queryCashForUser
	requireCashUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	queryCashForUser = func(userID, cashType, categoryID, description, exactDescription, fromDate, toDate string, limit, offset int) ([]*model.CashFlowEntity, int64, error) {
		got.userID = userID
		got.limit = limit
		got.offset = offset
		return []*model.CashFlowEntity{{Id: primitive.NewObjectID(), CategoryType: model.FlowTypeIncome, Amount: 100}}, 1, nil
	}
	t.Cleanup(func() {
		requireCashUser = originalRequire
		queryCashForUser = originalQuery
		cashUserId = ""
		_ = listCmd.Flags().Set("limit", "50")
		_ = listCmd.Flags().Set("offset", "0")
		_ = listCmd.Flags().Set("page", "0")
	})

	_ = listCmd.Flags().Set("limit", "25")
	_ = listCmd.Flags().Set("offset", "5")
	_ = listCmd.Flags().Set("page", "3")

	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("RunE returned error: %v", err)
	}
	if got.userID != authenticatedUserID || got.limit != 25 || got.offset != 50 {
		t.Fatalf("service args = %+v", got)
	}
}
