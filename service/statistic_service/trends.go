package statistic_service

import (
	"errors"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetTrendsForUser gets spending trends for the specified period
// Only includes transactions belonging to the specified user
func GetTrendsForUser(period, date, userId string) (*Trends, error) {
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

	// Calculate trends
	trends := calculateTrends(period, date, fromDate, toDate, cashFlows)

	return trends, nil
}

// calculateTrends groups data by sub-periods and analyzes trends
func calculateTrends(period, dateStr string, fromDate, toDate time.Time, cashFlows []model.CashFlowEntity) *Trends {
	trends := &Trends{
		Period:     dateStr,
		PeriodType: period,
		DataPoints: []TrendDataPoint{},
	}

	// Create map to group cash flows by date
	flowsByDate := make(map[string][]model.CashFlowEntity)
	for _, flow := range cashFlows {
		var dateKey string
		if period == "daily" {
			// For daily view, show hourly breakdown (simplified to daily)
			dateKey = util.FormatDateToStringWithoutDash(flow.BelongsDate)
		} else if period == "monthly" {
			// For monthly view, show daily breakdown
			dateKey = util.FormatDateToStringWithoutDash(flow.BelongsDate)
		} else {
			// For yearly view, show monthly breakdown
			dateKey = flow.BelongsDate.Format("200601") // YYYYMM format
		}
		flowsByDate[dateKey] = append(flowsByDate[dateKey], flow)
	}

	// Generate data points for each sub-period
	currentDate := fromDate
	var dataPoints []TrendDataPoint
	totalExpense := 0.0
	expenseCount := 0

	for currentDate.Before(toDate) {
		var dateKey string
		var displayDate string

		if period == "yearly" {
			// Monthly breakdown for yearly view
			dateKey = currentDate.Format("200601")
			displayDate = currentDate.Format("2006-01")
			currentDate = currentDate.AddDate(0, 1, 0) // Next month
		} else {
			// Daily breakdown for monthly/daily view
			dateKey = util.FormatDateToStringWithoutDash(currentDate)
			displayDate = util.FormatDateToStringWithDash(currentDate)
			currentDate = currentDate.AddDate(0, 0, 1) // Next day
		}

		income := 0.0
		expense := 0.0

		if flows, exists := flowsByDate[dateKey]; exists {
			for _, flow := range flows {
				if flow.FlowType == "income" {
					income += flow.Amount
				} else if flow.FlowType == "expense" {
					expense += flow.Amount
				}
			}
		}

		// Only add data point if there's any activity
		if income > 0 || expense > 0 {
			dataPoints = append(dataPoints, TrendDataPoint{
				Date:    displayDate,
				Income:  income,
				Expense: expense,
				Balance: income - expense,
			})

			totalExpense += expense
			expenseCount++
		}
	}

	trends.DataPoints = dataPoints

	// Analyze trends
	trends.Trends = analyzeTrendDirection(dataPoints, expenseCount, totalExpense)

	return trends
}

// analyzeTrendDirection determines if spending is increasing, decreasing, or stable
func analyzeTrendDirection(dataPoints []TrendDataPoint, expenseCount int, totalExpense float64) TrendAnalysis {
	analysis := TrendAnalysis{
		IncomeTrend:  "stable",
		ExpenseTrend: "stable",
	}

	if expenseCount > 0 {
		analysis.AverageMonthlyExpense = totalExpense / float64(expenseCount)
	}

	if len(dataPoints) < 2 {
		return analysis
	}

	// Calculate trend for income and expense
	// Compare first half average vs second half average
	midPoint := len(dataPoints) / 2

	// Income trend
	firstHalfIncome := 0.0
	secondHalfIncome := 0.0
	for i := 0; i < midPoint; i++ {
		firstHalfIncome += dataPoints[i].Income
	}
	for i := midPoint; i < len(dataPoints); i++ {
		secondHalfIncome += dataPoints[i].Income
	}

	if midPoint > 0 {
		firstHalfIncomeAvg := firstHalfIncome / float64(midPoint)
		secondHalfIncomeAvg := secondHalfIncome / float64(len(dataPoints)-midPoint)

		if secondHalfIncomeAvg > firstHalfIncomeAvg*1.1 {
			analysis.IncomeTrend = "increasing"
		} else if secondHalfIncomeAvg < firstHalfIncomeAvg*0.9 {
			analysis.IncomeTrend = "decreasing"
		}
	}

	// Expense trend
	firstHalfExpense := 0.0
	secondHalfExpense := 0.0
	for i := 0; i < midPoint; i++ {
		firstHalfExpense += dataPoints[i].Expense
	}
	for i := midPoint; i < len(dataPoints); i++ {
		secondHalfExpense += dataPoints[i].Expense
	}

	if midPoint > 0 {
		firstHalfExpenseAvg := firstHalfExpense / float64(midPoint)
		secondHalfExpenseAvg := secondHalfExpense / float64(len(dataPoints)-midPoint)

		if secondHalfExpenseAvg > firstHalfExpenseAvg*1.1 {
			analysis.ExpenseTrend = "increasing"
		} else if secondHalfExpenseAvg < firstHalfExpenseAvg*0.9 {
			analysis.ExpenseTrend = "decreasing"
		}
	}

	return analysis
}
