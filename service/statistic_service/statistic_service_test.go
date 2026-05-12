package statistic_service

import (
	"testing"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetDateRange(t *testing.T) {
	baseDate := time.Date(2026, time.May, 12, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		period    string
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "daily",
			period:    "daily",
			wantStart: time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.May, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "monthly",
			period:    "monthly",
			wantStart: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "yearly",
			period:    "yearly",
			wantStart: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := getDateRange(tt.period, baseDate)
			if !gotStart.Equal(tt.wantStart) {
				t.Fatalf("start = %s, want %s", gotStart, tt.wantStart)
			}
			if !gotEnd.Equal(tt.wantEnd) {
				t.Fatalf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestAnalyzeTrendDirection(t *testing.T) {
	tests := []struct {
		name             string
		dataPoints       []TrendDataPoint
		expenseCount     int
		totalExpense     float64
		wantIncomeTrend  string
		wantExpenseTrend string
		wantAverage      float64
	}{
		{
			name: "increasing income and expense",
			dataPoints: []TrendDataPoint{
				{Income: 100, Expense: 50},
				{Income: 100, Expense: 50},
				{Income: 130, Expense: 70},
				{Income: 140, Expense: 80},
			},
			expenseCount:     4,
			totalExpense:     250,
			wantIncomeTrend:  "increasing",
			wantExpenseTrend: "increasing",
			wantAverage:      62.5,
		},
		{
			name: "decreasing income and expense",
			dataPoints: []TrendDataPoint{
				{Income: 150, Expense: 80},
				{Income: 140, Expense: 70},
				{Income: 100, Expense: 50},
				{Income: 100, Expense: 50},
			},
			expenseCount:     4,
			totalExpense:     250,
			wantIncomeTrend:  "decreasing",
			wantExpenseTrend: "decreasing",
			wantAverage:      62.5,
		},
		{
			name: "stable with single point",
			dataPoints: []TrendDataPoint{
				{Income: 100, Expense: 50},
			},
			expenseCount:     1,
			totalExpense:     50,
			wantIncomeTrend:  "stable",
			wantExpenseTrend: "stable",
			wantAverage:      50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzeTrendDirection(tt.dataPoints, tt.expenseCount, tt.totalExpense)
			if got.IncomeTrend != tt.wantIncomeTrend {
				t.Fatalf("IncomeTrend = %q, want %q", got.IncomeTrend, tt.wantIncomeTrend)
			}
			if got.ExpenseTrend != tt.wantExpenseTrend {
				t.Fatalf("ExpenseTrend = %q, want %q", got.ExpenseTrend, tt.wantExpenseTrend)
			}
			if got.AverageMonthlyExpense != tt.wantAverage {
				t.Fatalf("AverageMonthlyExpense = %v, want %v", got.AverageMonthlyExpense, tt.wantAverage)
			}
		})
	}
}

func TestStatisticServiceGetSummaryForUserAggregatesByCategoryType(t *testing.T) {
	userID := primitive.NewObjectID()
	incomeCategoryID := primitive.NewObjectID()
	expenseCategoryID := primitive.NewObjectID()
	cashMapper := &statisticCashFlowMapperFake{
		flows: []model.CashFlowEntity{
			{Id: primitive.NewObjectID(), BelongsUserId: userID, CategoryId: incomeCategoryID, BelongsDate: time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC), Amount: 1000},
			{Id: primitive.NewObjectID(), BelongsUserId: userID, CategoryId: expenseCategoryID, BelongsDate: time.Date(2026, time.May, 2, 0, 0, 0, 0, time.UTC), Amount: 25.125},
			{Id: primitive.NewObjectID(), BelongsUserId: userID, CategoryId: expenseCategoryID, BelongsDate: time.Date(2026, time.May, 3, 0, 0, 0, 0, time.UTC), Amount: 10.335},
		},
	}
	categoryMapper := &statisticCategoryMapperFake{
		categories: map[primitive.ObjectID]model.CategoryEntity{
			incomeCategoryID:  {Id: incomeCategoryID, BelongsUserId: userID, Name: "Salary", Type: model.FlowTypeIncome},
			expenseCategoryID: {Id: expenseCategoryID, BelongsUserId: userID, Name: "Food", Type: model.FlowTypeExpense},
		},
	}
	service := NewStatisticService(cashMapper, categoryMapper)

	summary, err := service.GetSummaryForUser("monthly", "2026-05-12", userID.Hex())
	if err != nil {
		t.Fatalf("GetSummaryForUser returned error: %v", err)
	}

	if summary.Income != 1000 {
		t.Fatalf("Income = %.2f, want 1000.00", summary.Income)
	}
	if summary.Expense != 35.46 {
		t.Fatalf("Expense = %.2f, want 35.46", summary.Expense)
	}
	if summary.Balance != 964.54 {
		t.Fatalf("Balance = %.2f, want 964.54", summary.Balance)
	}
	if summary.Categories["Food"] != 35.46 {
		t.Fatalf("Food category = %.2f, want 35.46", summary.Categories["Food"])
	}
}

type statisticCashFlowMapperFake struct {
	flows []model.CashFlowEntity
}

func (fake *statisticCashFlowMapperFake) GetCashFlowByObjectId(plainId string) model.CashFlowEntity {
	return model.CashFlowEntity{}
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByObjectIdArray(plainIdList []string) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) InsertCashFlowByEntity(newEntity model.CashFlowEntity) string {
	return newEntity.Id.Hex()
}
func (fake *statisticCashFlowMapperFake) BulkInsertCashFlows(entities []model.CashFlowEntity) ([]string, error) {
	return nil, nil
}
func (fake *statisticCashFlowMapperFake) UpdateCashFlowByEntity(plainId string, updatedEntity model.CashFlowEntity) model.CashFlowEntity {
	return updatedEntity
}
func (fake *statisticCashFlowMapperFake) GetAllCashFlows(limit, offset int) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) GetAllCashFlowsIncludeDeleted(limit, offset int) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) CountAllCashFlows() int64 { return int64(len(fake.flows)) }
func (fake *statisticCashFlowMapperFake) DeleteCashFlowByObjectId(plainId string) model.CashFlowEntity {
	return model.CashFlowEntity{}
}
func (fake *statisticCashFlowMapperFake) DeleteCashFlowByBelongsDate(belongsDate time.Time) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) TruncateCashFlows() error { return nil }
func (fake *statisticCashFlowMapperFake) GetCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	return model.CashFlowEntity{}
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByDateRangeAndUser(from, to time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	var flows []model.CashFlowEntity
	for _, flow := range fake.flows {
		if flow.BelongsUserId == userId && !flow.BelongsDate.Before(from) && flow.BelongsDate.Before(to) {
			flows = append(flows, flow)
		}
	}
	return flows
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) GetAllCashFlowsByUser(userId primitive.ObjectID, limit, offset int) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) CountAllCashFlowsByUser(userId primitive.ObjectID) int64 {
	return 0
}
func (fake *statisticCashFlowMapperFake) DeleteCashFlowByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CashFlowEntity {
	return model.CashFlowEntity{}
}
func (fake *statisticCashFlowMapperFake) DeleteCashFlowsByBelongsDateAndUser(belongsDate time.Time, userId primitive.ObjectID) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) DeleteCashFlowsByCategoryIdAndUser(categoryPlainId string, userId primitive.ObjectID) int64 {
	return 0
}
func (fake *statisticCashFlowMapperFake) UpdateCashFlowByEntityAndUser(plainId string, updatedEntity model.CashFlowEntity, userId primitive.ObjectID) model.CashFlowEntity {
	return updatedEntity
}
func (fake *statisticCashFlowMapperFake) GetCashFlowsByFilter(filter model.CashFlowFilter) ([]model.CashFlowEntity, error) {
	return nil, nil
}
func (fake *statisticCashFlowMapperFake) CountCashFlowsByFilter(filter model.CashFlowFilter) (int64, error) {
	return 0, nil
}
func (fake *statisticCashFlowMapperFake) GetAllCashFlowsByUserIncludeDeleted(userId primitive.ObjectID) []model.CashFlowEntity {
	return nil
}
func (fake *statisticCashFlowMapperFake) DeleteAllCashFlowsByUser(userId primitive.ObjectID) (int64, error) {
	return 0, nil
}

type statisticCategoryMapperFake struct {
	categories map[primitive.ObjectID]model.CategoryEntity
}

func (fake *statisticCategoryMapperFake) GetCategoryByObjectId(plainId string) model.CategoryEntity {
	id, _ := primitive.ObjectIDFromHex(plainId)
	return fake.categories[id]
}
func (fake *statisticCategoryMapperFake) GetCategoryByName(categoryName string) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) GetCategoryByParentId(parentPlainId string) []model.CategoryEntity {
	return nil
}
func (fake *statisticCategoryMapperFake) InsertCategoryByEntity(newEntity model.CategoryEntity) string {
	return newEntity.Id.Hex()
}
func (fake *statisticCategoryMapperFake) UpdateCategoryByEntity(plainId string, updatedEntity model.CategoryEntity) model.CategoryEntity {
	return updatedEntity
}
func (fake *statisticCategoryMapperFake) GetAllCategories(limit, offset int) []model.CategoryEntity {
	return nil
}
func (fake *statisticCategoryMapperFake) GetAllCategoriesIncludeDeleted(limit, offset int) []model.CategoryEntity {
	return nil
}
func (fake *statisticCategoryMapperFake) CountAllCategories() int64 {
	return int64(len(fake.categories))
}
func (fake *statisticCategoryMapperFake) CountCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) (int64, error) {
	return 0, nil
}
func (fake *statisticCategoryMapperFake) DeleteCategoryByObjectId(plainId string) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) TruncateCategories() error { return nil }
func (fake *statisticCategoryMapperFake) GetCategoriesByUserAndType(userId primitive.ObjectID, categoryType string, limit, offset int) ([]model.CategoryEntity, error) {
	return nil, nil
}
func (fake *statisticCategoryMapperFake) GetRootCategoriesByUser(userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return nil, nil
}
func (fake *statisticCategoryMapperFake) GetRootCategoriesByUserAndType(userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return nil, nil
}
func (fake *statisticCategoryMapperFake) GetCategoriesByParentIdAndUser(parentId primitive.ObjectID, userId primitive.ObjectID) ([]model.CategoryEntity, error) {
	return nil, nil
}
func (fake *statisticCategoryMapperFake) GetCategoriesByParentIdUserAndType(parentId primitive.ObjectID, userId primitive.ObjectID, categoryType string) ([]model.CategoryEntity, error) {
	return nil, nil
}
func (fake *statisticCategoryMapperFake) GetCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	category := fake.GetCategoryByObjectId(plainId)
	if category.BelongsUserId == userId {
		return category
	}
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) GetCategoryByNameAndUser(categoryName string, userId primitive.ObjectID) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) GetCategoryByNameUserAndType(categoryName string, userId primitive.ObjectID, categoryType string) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) GetCategoryByNameUserTypeAndParent(categoryName string, userId primitive.ObjectID, categoryType string, parentId primitive.ObjectID) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) DeleteCategoryByObjectIdAndUser(plainId string, userId primitive.ObjectID) model.CategoryEntity {
	return model.CategoryEntity{}
}
func (fake *statisticCategoryMapperFake) UpdateCategoryByEntityAndUser(plainId string, updatedEntity model.CategoryEntity, userId primitive.ObjectID) model.CategoryEntity {
	return updatedEntity
}
func (fake *statisticCategoryMapperFake) GetAllCategoriesByUser(userId primitive.ObjectID, limit, offset int) []model.CategoryEntity {
	return nil
}
func (fake *statisticCategoryMapperFake) CountAllCategoriesByUser(userId primitive.ObjectID) int64 {
	return 0
}
func (fake *statisticCategoryMapperFake) GetAllCategoriesByUserIncludeDeleted(userId primitive.ObjectID) []model.CategoryEntity {
	return nil
}
func (fake *statisticCategoryMapperFake) DeleteAllCategoriesByUser(userId primitive.ObjectID) (int64, error) {
	return 0, nil
}
