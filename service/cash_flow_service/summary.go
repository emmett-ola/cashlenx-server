package cash_flow_service

import (
	"errors"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
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
		dayResults := cash_flow_mapper.INSTANCE.GetCashFlowsByBelongsDate(currentDate)

		for _, cashFlow := range dayResults {
			summary.TransactionCount++

			if cashFlow.FlowType == model.FlowTypeIncome {
				summary.TotalIncome += cashFlow.Amount
			} else {
				summary.TotalExpense += cashFlow.Amount
				// Get category name for breakdown
				category := category_mapper.INSTANCE.GetCategoryByObjectId(cashFlow.CategoryId.Hex())
				if !category.IsEmpty() {
					summary.CategoryBreakdown[category.Name] += cashFlow.Amount
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

	// Query transactions for period using user-specific methods
	summary := &Summary{
		CategoryBreakdown: make(map[string]float64),
	}

	// Use date range query for efficiency instead of iterating day by day
	cashFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	for _, cashFlow := range cashFlows {
		summary.TransactionCount++

		if cashFlow.FlowType == model.FlowTypeIncome {
			summary.TotalIncome += cashFlow.Amount
		} else {
			summary.TotalExpense += cashFlow.Amount
			// Get category name for breakdown
			category := category_mapper.INSTANCE.GetCategoryByObjectId(cashFlow.CategoryId.Hex())
			if !category.IsEmpty() {
				summary.CategoryBreakdown[category.Name] += cashFlow.Amount
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

	return summary, nil
}
