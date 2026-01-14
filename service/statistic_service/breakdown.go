package statistic_service

import (
	"errors"
	"sort"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetBreakdownForUser gets category breakdown for the specified period
// Only includes transactions belonging to the specified user
func GetBreakdownForUser(period, date, userId string) (*Breakdown, error) {
	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("invalid user ID")
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

	// Calculate breakdown
	breakdown := calculateBreakdown(date, cashFlows, userObjectId)

	return breakdown, nil
}

// calculateBreakdown groups cash flows by category and calculates percentages
func calculateBreakdown(period string, cashFlows []model.CashFlowEntity, userId primitive.ObjectID) *Breakdown {
	breakdown := &Breakdown{
		Period:            period,
		ExpenseCategories: []CategoryBreakdownItem{},
		IncomeCategories:  []CategoryBreakdownItem{},
	}

	// Maps to accumulate amounts and counts by category
	expenseMap := make(map[string]*CategoryBreakdownItem)
	incomeMap := make(map[string]*CategoryBreakdownItem)

	for _, flow := range cashFlows {
		categoryName := getCategoryName(flow.CategoryId, userId)

		if flow.FlowType == "income" {
			breakdown.TotalIncome += flow.Amount

			if item, exists := incomeMap[categoryName]; exists {
				item.Amount += flow.Amount
				item.Count++
			} else {
				incomeMap[categoryName] = &CategoryBreakdownItem{
					Category: categoryName,
					Amount:   flow.Amount,
					Count:    1,
				}
			}
		} else if flow.FlowType == "expense" {
			breakdown.TotalExpense += flow.Amount

			if item, exists := expenseMap[categoryName]; exists {
				item.Amount += flow.Amount
				item.Count++
			} else {
				expenseMap[categoryName] = &CategoryBreakdownItem{
					Category: categoryName,
					Amount:   flow.Amount,
					Count:    1,
				}
			}
		}
	}

	// Convert maps to slices and calculate percentages
	for _, item := range expenseMap {
		if breakdown.TotalExpense > 0 {
			item.Percentage = (item.Amount / breakdown.TotalExpense) * 100
		}
		breakdown.ExpenseCategories = append(breakdown.ExpenseCategories, *item)
	}

	for _, item := range incomeMap {
		if breakdown.TotalIncome > 0 {
			item.Percentage = (item.Amount / breakdown.TotalIncome) * 100
		}
		breakdown.IncomeCategories = append(breakdown.IncomeCategories, *item)
	}

	// Sort by amount descending
	sort.Slice(breakdown.ExpenseCategories, func(i, j int) bool {
		return breakdown.ExpenseCategories[i].Amount > breakdown.ExpenseCategories[j].Amount
	})

	sort.Slice(breakdown.IncomeCategories, func(i, j int) bool {
		return breakdown.IncomeCategories[i].Amount > breakdown.IncomeCategories[j].Amount
	})

	return breakdown
}
