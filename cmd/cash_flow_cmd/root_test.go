package cash_flow_cmd

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
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

func TestCashCommandsPassFlagsAndUserToServices(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	cashID := primitive.NewObjectID()
	var queryIDArgs, queryDateArgs, rangeArgs, deleteIDArgs, deleteDateArgs, updateArgs, incomeArgs, summaryArgs struct {
		id          string
		fromDate    string
		toDate      string
		date        string
		category    string
		amount      float64
		description string
		period      string
		userID      string
	}

	originalRequire := requireCashUser
	requireCashUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	originalQueryID := queryCashByIDForUser
	queryCashByIDForUser = func(id, userID string) (model.CashFlowEntity, error) {
		queryIDArgs.id = id
		queryIDArgs.userID = userID
		return testCashFlow(cashID, model.FlowTypeExpense, 12), nil
	}
	originalQueryDate := queryCashByDateForUser
	queryCashByDateForUser = func(date, userID string) ([]model.CashFlowEntity, error) {
		queryDateArgs.date = date
		queryDateArgs.userID = userID
		return []model.CashFlowEntity{testCashFlow(cashID, model.FlowTypeExpense, 12)}, nil
	}
	originalRange := queryCashRangeForUser
	queryCashRangeForUser = func(from, to, userID string) ([]model.CashFlowEntity, error) {
		rangeArgs.fromDate = from
		rangeArgs.toDate = to
		rangeArgs.userID = userID
		return []model.CashFlowEntity{testCashFlow(cashID, model.FlowTypeIncome, 100)}, nil
	}
	originalDeleteID := deleteCashByIDForUser
	deleteCashByIDForUser = func(id, userID string) (model.CashFlowEntity, error) {
		deleteIDArgs.id = id
		deleteIDArgs.userID = userID
		return testCashFlow(cashID, model.FlowTypeExpense, 12), nil
	}
	originalDeleteDate := deleteCashByDateForUser
	deleteCashByDateForUser = func(date, userID string) ([]model.CashFlowEntity, error) {
		deleteDateArgs.date = date
		deleteDateArgs.userID = userID
		return []model.CashFlowEntity{testCashFlow(cashID, model.FlowTypeExpense, 12)}, nil
	}
	originalUpdate := updateCashForUser
	updateCashForUser = func(id, date, category string, amount float64, description, userID string) (model.CashFlowEntity, error) {
		updateArgs.id = id
		updateArgs.date = date
		updateArgs.category = category
		updateArgs.amount = amount
		updateArgs.description = description
		updateArgs.userID = userID
		return testCashFlow(cashID, model.FlowTypeExpense, amount), nil
	}
	originalIncome := saveIncomeForUser
	saveIncomeForUser = func(date, category string, amount float64, description, userID string) (model.CashFlowEntity, error) {
		incomeArgs.date = date
		incomeArgs.category = category
		incomeArgs.amount = amount
		incomeArgs.description = description
		incomeArgs.userID = userID
		return testCashFlow(cashID, model.FlowTypeIncome, amount), nil
	}
	originalSummary := getCashSummaryForUser
	getCashSummaryForUser = func(period, date, userID string) (*cash_flow_service.Summary, error) {
		summaryArgs.period = period
		summaryArgs.date = date
		summaryArgs.userID = userID
		return &cash_flow_service.Summary{TotalIncome: 100, TotalExpense: 25, Balance: 75, TransactionCount: 2, CategoryBreakdown: map[string]float64{"Food": 25}}, nil
	}
	t.Cleanup(func() {
		requireCashUser = originalRequire
		queryCashByIDForUser = originalQueryID
		queryCashByDateForUser = originalQueryDate
		queryCashRangeForUser = originalRange
		deleteCashByIDForUser = originalDeleteID
		deleteCashByDateForUser = originalDeleteDate
		updateCashForUser = originalUpdate
		saveIncomeForUser = originalIncome
		getCashSummaryForUser = originalSummary
		resetCashCommandState()
		resetCashFlagValues()
	})

	_ = queryCmd.Flags().Set("id", cashID.Hex())
	if err := queryCmd.RunE(queryCmd, nil); err != nil {
		t.Fatalf("query by id RunE returned error: %v", err)
	}
	_ = queryCmd.Flags().Set("id", "")
	_ = queryCmd.Flags().Set("date", "2026-05-25")
	if err := queryCmd.RunE(queryCmd, nil); err != nil {
		t.Fatalf("query by date RunE returned error: %v", err)
	}

	fromDate, toDate = "2026-05-01", "2026-05-31"
	if err := rangeCmd.RunE(rangeCmd, nil); err != nil {
		t.Fatalf("range RunE returned error: %v", err)
	}

	_ = deleteCmd.Flags().Set("id", cashID.Hex())
	if err := deleteCmd.RunE(deleteCmd, nil); err != nil {
		t.Fatalf("delete by id RunE returned error: %v", err)
	}
	_ = deleteCmd.Flags().Set("id", "")
	_ = deleteCmd.Flags().Set("date", "2026-05-25")
	if err := deleteCmd.RunE(deleteCmd, nil); err != nil {
		t.Fatalf("delete by date RunE returned error: %v", err)
	}

	_ = updateCmd.Flags().Set("id", cashID.Hex())
	_ = updateCmd.Flags().Set("date", "2026-05-26")
	_ = updateCmd.Flags().Set("category", "Food")
	_ = updateCmd.Flags().Set("amount", "14.50")
	_ = updateCmd.Flags().Set("description", "dinner")
	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatalf("update RunE returned error: %v", err)
	}

	_ = incomeCmd.Flags().Set("date", "2026-05-27")
	_ = incomeCmd.Flags().Set("category", "Salary")
	_ = incomeCmd.Flags().Set("amount", "100")
	_ = incomeCmd.Flags().Set("description", "paycheck")
	if err := incomeCmd.RunE(incomeCmd, nil); err != nil {
		t.Fatalf("income RunE returned error: %v", err)
	}

	summaryPeriod, summaryDate = "monthly", "2026-05"
	if err := summaryCmd.RunE(summaryCmd, nil); err != nil {
		t.Fatalf("summary RunE returned error: %v", err)
	}

	if queryIDArgs.id != cashID.Hex() || queryIDArgs.userID != authenticatedUserID {
		t.Fatalf("query id args = %+v", queryIDArgs)
	}
	if queryDateArgs.date != "2026-05-25" || queryDateArgs.userID != authenticatedUserID {
		t.Fatalf("query date args = %+v", queryDateArgs)
	}
	if rangeArgs.fromDate != "2026-05-01" || rangeArgs.toDate != "2026-05-31" || rangeArgs.userID != authenticatedUserID {
		t.Fatalf("range args = %+v", rangeArgs)
	}
	if deleteIDArgs.id != cashID.Hex() || deleteIDArgs.userID != authenticatedUserID {
		t.Fatalf("delete id args = %+v", deleteIDArgs)
	}
	if deleteDateArgs.date != "2026-05-25" || deleteDateArgs.userID != authenticatedUserID {
		t.Fatalf("delete date args = %+v", deleteDateArgs)
	}
	if updateArgs.id != cashID.Hex() || updateArgs.date != "2026-05-26" || updateArgs.category != "Food" || updateArgs.amount != 14.50 || updateArgs.description != "dinner" || updateArgs.userID != authenticatedUserID {
		t.Fatalf("update args = %+v", updateArgs)
	}
	if incomeArgs.date != "2026-05-27" || incomeArgs.category != "Salary" || incomeArgs.amount != 100 || incomeArgs.description != "paycheck" || incomeArgs.userID != authenticatedUserID {
		t.Fatalf("income args = %+v", incomeArgs)
	}
	if summaryArgs.period != "monthly" || summaryArgs.date != "2026-05" || summaryArgs.userID != authenticatedUserID {
		t.Fatalf("summary args = %+v", summaryArgs)
	}
}

func testCashFlow(id primitive.ObjectID, flowType string, amount float64) model.CashFlowEntity {
	return model.CashFlowEntity{
		Id:           id,
		BelongsDate:  time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
		CategoryName: "Category",
		CategoryType: flowType,
		Amount:       amount,
		Description:  "description",
	}
}

func resetCashCommandState() {
	cashUserId = ""
	fromDate = ""
	toDate = ""
	summaryPeriod = ""
	summaryDate = ""
}

func resetCashFlagValues() {
	for _, reset := range []struct {
		cmd  string
		name string
	}{
		{"query", "id"},
		{"query", "date"},
		{"query", "exact"},
		{"query", "fuzzy"},
		{"delete", "id"},
		{"delete", "date"},
		{"update", "id"},
		{"update", "date"},
		{"update", "category"},
		{"update", "description"},
		{"income", "date"},
		{"income", "category"},
		{"income", "description"},
	} {
		switch reset.cmd {
		case "query":
			_ = queryCmd.Flags().Set(reset.name, "")
		case "delete":
			_ = deleteCmd.Flags().Set(reset.name, "")
		case "update":
			_ = updateCmd.Flags().Set(reset.name, "")
		case "income":
			_ = incomeCmd.Flags().Set(reset.name, "")
		}
	}
	_ = updateCmd.Flags().Set("amount", "0")
	_ = incomeCmd.Flags().Set("amount", "0")
}
