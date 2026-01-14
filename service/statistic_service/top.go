package statistic_service

import (
	"errors"
	"sort"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetTopExpensesForUser gets top N expenses for the specified period
// Only includes transactions belonging to the specified user
func GetTopExpensesForUser(limit int, period, date, userId string) (*TopExpenses, error) {
	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Validate limit
	if limit <= 0 {
		limit = 10 // Default to top 10
	}

	// Parse and validate period type
	if period != "daily" && period != "monthly" && period != "yearly" {
		return nil, errors.New("period must be 'daily', 'monthly', or 'yearly'")
	}

	// Parse the date string
	baseDate, err := util.ParseDate(date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYYMMDD or YYYY-MM-DD")
	}

	// Calculate date range based on period
	fromDate, toDate := getDateRange(period, baseDate)

	// Get all cash flows for user in this period
	cashFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	// Calculate top expenses
	topExpenses := calculateTopExpenses(date, limit, cashFlows, userObjectId)

	return topExpenses, nil
}

// calculateTopExpenses finds the top N expenses and calculates percentages
func calculateTopExpenses(period string, limit int, cashFlows []model.CashFlowEntity, userId primitive.ObjectID) *TopExpenses {
	topExpenses := &TopExpenses{
		Period:   period,
		Limit:    limit,
		Expenses: []TopExpense{},
	}

	// Filter and collect only expenses
	var expenses []TopExpense
	for _, flow := range cashFlows {
		if flow.FlowType == "expense" {
			topExpenses.TotalExpense += flow.Amount

			categoryName := getCategoryName(flow.CategoryId, userId)
			dateStr := util.FormatDateToStringWithoutDash(flow.BelongsDate)

			expenses = append(expenses, TopExpense{
				ID:          flow.Id.Hex(),
				Date:        dateStr,
				Category:    categoryName,
				Amount:      flow.Amount,
				Description: flow.Description,
			})
		}
	}

	// Sort by amount descending
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].Amount > expenses[j].Amount
	})

	// Take top N
	if len(expenses) > limit {
		expenses = expenses[:limit]
	}

	// Calculate percentages
	if topExpenses.TotalExpense > 0 {
		for i := range expenses {
			expenses[i].Percentage = (expenses[i].Amount / topExpenses.TotalExpense) * 100
		}
	}

	topExpenses.Expenses = expenses

	return topExpenses
}
