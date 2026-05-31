package cash_flow_service

import (
	"errors"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Summary represents financial summary data
type Summary struct {
	TotalIncome       float64            `json:"total_income"`
	TotalExpense      float64            `json:"total_expense"`
	Balance           float64            `json:"balance"`
	TransactionCount  int                `json:"transaction_count"`
	CategoryBreakdown map[string]float64 `json:"category_breakdown"`
}

// GetSummary returns financial summary for a given period
func GetSummary(period, date string) (*Summary, error) {
	return defaultCashFlowService().GetSummary(period, date)
}

func (s *CashFlowService) GetSummary(period, date string) (*Summary, error) {
	validPeriods := map[string]bool{
		"daily":   true,
		"monthly": true,
		"yearly":  true,
	}

	if !validPeriods[period] {
		return nil, errors.New("invalid period: must be daily, monthly, or yearly")
	}

	var fromDate, toDate time.Time
	var err error

	// Parse date based on period
	switch period {
	case "daily":
		// Date format: YYYY-MM-DD
		fromDate = util.FormatDateFromStringWithoutDash(date)
		if fromDate.IsZero() {
			return nil, errors.New("invalid date format for daily, use YYYY-MM-DD")
		}
		toDate = fromDate
	case "monthly":
		// Date format: YYYY-MM
		parts := strings.Split(date, "-")
		if len(parts) != 2 {
			return nil, errors.New("invalid date format for monthly, use YYYY-MM")
		}
		fromDate, err = time.Parse("2006-01", date)
		if err != nil {
			return nil, errors.New("invalid date format for monthly, use YYYY-MM")
		}
		toDate = fromDate.AddDate(0, 1, -1) // Last day of month
	case "yearly":
		// Date format: YYYY
		fromDate, err = time.Parse("2006", date)
		if err != nil {
			return nil, errors.New("invalid date format for yearly, use YYYY")
		}
		toDate = fromDate.AddDate(1, 0, -1) // Last day of year
	}

	// Query transactions for period
	summary := &Summary{
		CategoryBreakdown: make(map[string]float64),
	}

	currentDate := fromDate
	for !currentDate.After(toDate) {
		dayResults := s.cashFlowMapper.GetCashFlowsByBelongsDate(currentDate)

		for _, cashFlow := range dayResults {
			summary.TransactionCount++

			// Get category to determine true type
			category := s.categoryMapper.GetCategoryByObjectId(cashFlow.CategoryId.Hex())

			isIncome := false
			if !category.IsEmpty() {
				if strings.EqualFold(category.Type, model.FlowTypeIncome) {
					isIncome = true
				} else if strings.EqualFold(category.Type, model.FlowTypeExpense) {
					isIncome = false
				}
			} else {
				// Fallback if category is missing: default to expense or skip?
				// Since we can't access FlowType anymore, we have to guess or assume expense.
				// Let's assume expense for safety (or income? Expense is more common to track).
				// Or log a warning.
				// For now, let's treat it as expense to be safe, or maybe skip?
				// If we treat as expense, it might skew expense.
				// Let's default to Expense as false.
				isIncome = false
			}

			if isIncome {
				summary.TotalIncome += cashFlow.Amount
			} else {
				summary.TotalExpense += cashFlow.Amount
				if !category.IsEmpty() {
					summary.CategoryBreakdown[category.Name] += cashFlow.Amount
				} else {
					summary.CategoryBreakdown["Unknown"] += cashFlow.Amount
				}
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	summary.Balance = summary.TotalIncome - summary.TotalExpense

	// Format float precision to 2 decimal places
	summary.TotalIncome = util.RoundFloat(summary.TotalIncome, 2)
	summary.TotalExpense = util.RoundFloat(summary.TotalExpense, 2)
	summary.Balance = util.RoundFloat(summary.Balance, 2)
	for k, v := range summary.CategoryBreakdown {
		summary.CategoryBreakdown[k] = util.RoundFloat(v, 2)
	}

	return summary, nil
}

// GetSummaryForUser returns financial summary for a given period for a specific user
func GetSummaryForUser(period, date string, userId string) (*Summary, error) {
	return defaultCashFlowService().GetSummaryForUser(period, date, userId)
}

// GetTotalSummaryForUser returns financial summary across all active cash flows for a user.
func GetTotalSummaryForUser(userId string) (*Summary, error) {
	return defaultCashFlowService().GetTotalSummaryForUser(userId)
}

func (s *CashFlowService) GetTotalSummaryForUser(userId string) (*Summary, error) {
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, errors.New("invalid user ID")
	}

	cashFlows, err := s.cashFlowMapper.GetCashFlowsByFilter(model.CashFlowFilter{
		UserId: userObjectId,
	})
	if err != nil {
		return nil, err
	}

	return s.buildSummary(cashFlows), nil
}

func (s *CashFlowService) GetSummaryForUser(period, date string, userId string) (*Summary, error) {
	validPeriods := map[string]bool{
		"daily":   true,
		"monthly": true,
		"yearly":  true,
	}

	if !validPeriods[period] {
		return nil, errors.New("invalid period: must be daily, monthly, or yearly")
	}

	// Validate and convert userId
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, errors.New("invalid user ID")
	}

	var fromDate, toDate time.Time
	var err error

	// Parse date based on period
	switch period {
	case "daily":
		// Date format: YYYY-MM-DD or YYYYMMDD
		parsedDate, err := util.ParseDate(date)
		if err != nil {
			return nil, errors.New("invalid date format for daily")
		}
		fromDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
		toDate = fromDate
	case "monthly":
		// Date format: YYYY-MM or YYYYMM
		var parsedDate time.Time
		if strings.Contains(date, "-") {
			parsedDate, err = time.Parse("2006-01", date)
		} else if len(date) == 6 {
			parsedDate, err = time.Parse("200601", date)
		} else {
			return nil, errors.New("invalid date format for monthly, use YYYY-MM or YYYYMM")
		}
		if err != nil {
			return nil, errors.New("invalid date format for monthly")
		}
		fromDate = time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		toDate = fromDate.AddDate(0, 1, -1) // Last day of month
	case "yearly":
		// Date format: YYYY
		parsedDate, err := time.Parse("2006", date)
		if err != nil {
			return nil, errors.New("invalid date format for yearly, use YYYY")
		}
		fromDate = time.Date(parsedDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		toDate = fromDate.AddDate(1, 0, -1) // Last day of year
	}

	// Use date range query for efficiency instead of iterating day by day
	cashFlows := s.cashFlowMapper.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	return s.buildSummary(cashFlows), nil
}

func (s *CashFlowService) buildSummary(cashFlows []model.CashFlowEntity) *Summary {
	summary := &Summary{
		CategoryBreakdown: make(map[string]float64),
	}

	for _, cashFlow := range cashFlows {
		summary.TransactionCount++

		// Get category to determine true type
		category := s.categoryMapper.GetCategoryByObjectId(cashFlow.CategoryId.Hex())

		isIncome := false
		if !category.IsEmpty() {
			if strings.EqualFold(category.Type, model.FlowTypeIncome) {
				isIncome = true
			} else if strings.EqualFold(category.Type, model.FlowTypeExpense) {
				isIncome = false
			}
		} else {
			isIncome = false
		}

		if isIncome {
			summary.TotalIncome += cashFlow.Amount
		} else {
			summary.TotalExpense += cashFlow.Amount
			if !category.IsEmpty() {
				summary.CategoryBreakdown[category.Name] += cashFlow.Amount
			} else {
				summary.CategoryBreakdown["Unknown"] += cashFlow.Amount
			}
		}
	}

	summary.Balance = summary.TotalIncome - summary.TotalExpense

	// Format float precision to 2 decimal places
	summary.TotalIncome = util.RoundFloat(summary.TotalIncome, 2)
	summary.TotalExpense = util.RoundFloat(summary.TotalExpense, 2)
	summary.Balance = util.RoundFloat(summary.Balance, 2)
	for k, v := range summary.CategoryBreakdown {
		summary.CategoryBreakdown[k] = util.RoundFloat(v, 2)
	}

	return summary
}
