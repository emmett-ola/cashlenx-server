package statistic_service

import (
	"errors"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetSummaryForUser gets financial summary for the specified period
// Only includes transactions belonging to the specified user
func GetSummaryForUser(period, date, userId string) (*Summary, error) {
	return defaultStatisticService().GetSummaryForUser(period, date, userId)
}

func (s *StatisticService) GetSummaryForUser(period, date, userId string) (*Summary, error) {
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
	baseDate, err := parsePeriodDate(period, date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYYMMDD or YYYY-MM-DD")
	}

	// Calculate date range based on period
	fromDate, toDate := getDateRange(period, baseDate)

	// Get all cash flows for user in this period
	cashFlows := s.cashFlowMapper.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	// Calculate summary statistics
	summary := s.calculateSummary(period, date, cashFlows, userObjectId)

	return summary, nil
}

// getDateRange calculates the start and end date for a given period
func getDateRange(period string, baseDate time.Time) (time.Time, time.Time) {
	year, month, day := baseDate.Date()

	switch period {
	case "daily":
		// Single day: from start of day to end of day
		start := time.Date(year, month, day, 0, 0, 0, 0, baseDate.Location())
		end := start.AddDate(0, 0, 1) // Next day (exclusive)
		return start, end

	case "monthly":
		// Entire month: from first day to last day
		start := time.Date(year, month, 1, 0, 0, 0, 0, baseDate.Location())
		end := start.AddDate(0, 1, 0) // First day of next month (exclusive)
		return start, end

	case "yearly":
		// Entire year: from Jan 1 to Dec 31
		start := time.Date(year, 1, 1, 0, 0, 0, 0, baseDate.Location())
		end := start.AddDate(1, 0, 0) // Jan 1 of next year (exclusive)
		return start, end

	default:
		return baseDate, baseDate
	}
}

// calculateSummary computes all summary statistics from cash flows
func calculateSummary(period, date string, cashFlows []model.CashFlowEntity, userId primitive.ObjectID) *Summary {
	return defaultStatisticService().calculateSummary(period, date, cashFlows, userId)
}

func (s *StatisticService) calculateSummary(period, date string, cashFlows []model.CashFlowEntity, userId primitive.ObjectID) *Summary {
	summary := &Summary{
		Period:     date,
		PeriodType: period,
		Categories: make(map[string]float64),
	}

	if len(cashFlows) == 0 {
		return summary
	}

	totalAmount := 0.0

	for _, flow := range cashFlows {
		// Count transaction
		summary.TransactionCount++

		// Get category for grouping and type determination
		category := s.categoryMapper.GetCategoryByObjectIdAndUser(flow.CategoryId.Hex(), userId)
		categoryName := "Unknown"
		categoryType := ""
		if !category.IsEmpty() {
			categoryName = category.Name
			categoryType = category.Type
		}

		if strings.EqualFold(categoryType, "income") {
			summary.Income += flow.Amount
			summary.IncomeCount++
			// Track income by category
			summary.Categories[categoryName] += flow.Amount
		} else if strings.EqualFold(categoryType, "expense") {
			summary.Expense += flow.Amount
			summary.ExpenseCount++
			// Track expense by category (positive for display)
			summary.Categories[categoryName] += flow.Amount
		}

		totalAmount += flow.Amount
	}

	// Calculate balance
	summary.Balance = summary.Income - summary.Expense

	// Format float precision to 2 decimal places
	summary.Income = util.RoundFloat(summary.Income, 2)
	summary.Expense = util.RoundFloat(summary.Expense, 2)
	summary.Balance = util.RoundFloat(summary.Balance, 2)
	for k, v := range summary.Categories {
		summary.Categories[k] = util.RoundFloat(v, 2)
	}

	// Calculate average transaction
	if summary.TransactionCount > 0 {
		summary.AverageTransaction = util.RoundFloat(totalAmount/float64(summary.TransactionCount), 2)
	}

	return summary
}
