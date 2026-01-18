package statistic_service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDashboardOverviewForUser returns comprehensive dashboard data
func GetDashboardOverviewForUser(period, date, userId string) (*DashboardOverview, error) {
	// Get summary
	summary, err := GetSummaryForUser(period, date, userId)
	if err != nil {
		return nil, err
	}

	// Get breakdown for top categories
	breakdown, err := GetBreakdownForUser(period, date, userId)
	if err != nil {
		return nil, err
	}

	// Get top 5 expense categories
	topCategories := []CategoryBreakdownItem{}
	if len(breakdown.ExpenseCategories) > 0 {
		limit := 5
		if len(breakdown.ExpenseCategories) < 5 {
			limit = len(breakdown.ExpenseCategories)
		}
		topCategories = breakdown.ExpenseCategories[:limit]
	}

	// Get trends to determine recent trend
	trends, err := GetTrendsForUser(period, date, userId)
	if err != nil {
		return nil, err
	}

	// Calculate quick stats
	userObjectId, _ := primitive.ObjectIDFromHex(userId)
	fromDate, toDate := getDateRange(period, util.FormatDateFromStringWithoutDash(date))
	cashFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	quickStats := QuickStats{
		TotalTransactions: len(cashFlows),
	}

	if len(cashFlows) > 0 {
		// Calculate average daily spending
		days := toDate.Sub(fromDate).Hours() / 24
		if days > 0 {
			quickStats.AverageDaily = summary.Expense / days
		}

		// Find highest and lowest expenses
		highestExpense := 0.0
		lowestExpense := math.MaxFloat64
		for _, flow := range cashFlows {
			// Get category type
			category := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(flow.CategoryId.Hex(), userObjectId)
			categoryType := ""
			if !category.IsEmpty() {
				categoryType = category.Type
			}

			if strings.EqualFold(categoryType, "expense") {
				if flow.Amount > highestExpense {
					highestExpense = flow.Amount
				}
				if flow.Amount < lowestExpense {
					lowestExpense = flow.Amount
				}
			}
		}
		quickStats.HighestExpense = highestExpense
		if lowestExpense == math.MaxFloat64 {
			quickStats.LowestExpense = 0
		} else {
			quickStats.LowestExpense = lowestExpense
		}
	}

	dashboard := &DashboardOverview{
		Period:        date,
		PeriodType:    period,
		Summary:       summary,
		TopCategories: topCategories,
		RecentTrend:   trends.Trends.ExpenseTrend,
		QuickStats:    quickStats,
	}

	return dashboard, nil
}

// GetIncomeExpenseChartDataForUser returns time-series data for charts
func GetIncomeExpenseChartDataForUser(period, date, userId string) (*IncomeExpenseChartData, error) {
	trends, err := GetTrendsForUser(period, date, userId)
	if err != nil {
		return nil, err
	}

	chartData := &IncomeExpenseChartData{
		Labels:  []string{},
		Income:  []float64{},
		Expense: []float64{},
		Balance: []float64{},
		Period:  period,
	}

	for _, dp := range trends.DataPoints {
		chartData.Labels = append(chartData.Labels, dp.Date)
		chartData.Income = append(chartData.Income, dp.Income)
		chartData.Expense = append(chartData.Expense, dp.Expense)
		chartData.Balance = append(chartData.Balance, dp.Balance)
	}

	if len(chartData.Labels) > 0 {
		chartData.FromDate = chartData.Labels[0]
		chartData.ToDate = chartData.Labels[len(chartData.Labels)-1]
	}

	return chartData, nil
}

// GetCategoryDistributionForUser returns pie/donut chart data
func GetCategoryDistributionForUser(period, date, flowType, userId string) (*CategoryDistribution, error) {
	breakdown, err := GetBreakdownForUser(period, date, userId)
	if err != nil {
		return nil, err
	}

	distribution := &CategoryDistribution{
		Labels:      []string{},
		Values:      []float64{},
		Percentages: []float64{},
		Colors:      []string{},
		Type:        flowType,
	}

	var categories []CategoryBreakdownItem
	if flowType == "income" {
		categories = breakdown.IncomeCategories
		distribution.Total = breakdown.TotalIncome
	} else {
		categories = breakdown.ExpenseCategories
		distribution.Total = breakdown.TotalExpense
	}

	// Predefined color palette for consistency
	colorPalette := []string{
		"#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF",
		"#FF9F40", "#FF6384", "#C9CBCF", "#4BC0C0", "#FF6384",
	}

	for i, cat := range categories {
		distribution.Labels = append(distribution.Labels, cat.Category)
		distribution.Values = append(distribution.Values, cat.Amount)
		distribution.Percentages = append(distribution.Percentages, cat.Percentage)

		// Assign color from palette
		colorIndex := i % len(colorPalette)
		distribution.Colors = append(distribution.Colors, colorPalette[colorIndex])
	}

	return distribution, nil
}

// GetMonthlyComparisonForUser returns month-over-month comparison data
func GetMonthlyComparisonForUser(year, userId string) (*MonthlyComparison, error) {
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	comparison := &MonthlyComparison{
		Year:    year,
		Months:  []string{},
		Income:  []float64{},
		Expense: []float64{},
		Balance: []float64{},
	}

	// Parse year
	yearInt := 0
	fmt.Sscanf(year, "%d", &yearInt)
	if yearInt == 0 {
		return nil, errors.New("invalid year format")
	}

	// Get data for each month
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for month := 1; month <= 12; month++ {
		// Create date range for the month
		fromDate := time.Date(yearInt, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		toDate := fromDate.AddDate(0, 1, 0) // First day of next month

		// Get cash flows for this month
		cashFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

		monthIncome := 0.0
		monthExpense := 0.0

		for _, flow := range cashFlows {
			// Get category type
			category := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(flow.CategoryId.Hex(), userObjectId)
			categoryType := ""
			if !category.IsEmpty() {
				categoryType = category.Type
			}

			if strings.EqualFold(categoryType, "income") {
				monthIncome += flow.Amount
			} else if strings.EqualFold(categoryType, "expense") {
				monthExpense += flow.Amount
			}
		}

		comparison.Months = append(comparison.Months, monthNames[month-1])
		comparison.Income = append(comparison.Income, monthIncome)
		comparison.Expense = append(comparison.Expense, monthExpense)
		comparison.Balance = append(comparison.Balance, monthIncome-monthExpense)
	}

	return comparison, nil
}

// GetSpendingHeatmapForUser returns calendar heatmap data
func GetSpendingHeatmapForUser(year, userId string) (*SpendingHeatmap, error) {
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	heatmap := &SpendingHeatmap{
		Year: year,
		Data: []HeatmapDataPoint{},
		Max:  0,
		Min:  math.MaxFloat64,
	}

	// Parse year
	yearInt := 0
	fmt.Sscanf(year, "%d", &yearInt)
	if yearInt == 0 {
		return nil, errors.New("invalid year format")
	}

	// Create a map to store daily spending
	dailySpending := make(map[string]*HeatmapDataPoint)

	// Get all cash flows for the year
	fromDate := time.Date(yearInt, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := fromDate.AddDate(1, 0, 0) // First day of next year

	cashFlows := cash_flow_mapper.INSTANCE.GetCashFlowsByDateRangeAndUser(fromDate, toDate, userObjectId)

	// Aggregate spending by day
	for _, flow := range cashFlows {
		// Get category type
		category := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(flow.CategoryId.Hex(), userObjectId)
		categoryType := ""
		if !category.IsEmpty() {
			categoryType = category.Type
		}

		if strings.EqualFold(categoryType, "expense") {
			dateKey := util.FormatDateToStringWithDash(flow.BelongsDate)

			if dp, exists := dailySpending[dateKey]; exists {
				dp.Amount += flow.Amount
				dp.Count++
			} else {
				dailySpending[dateKey] = &HeatmapDataPoint{
					Date:   dateKey,
					Amount: flow.Amount,
					Count:  1,
				}
			}
		}
	}

	// Convert map to slice and find min/max
	for _, dp := range dailySpending {
		heatmap.Data = append(heatmap.Data, *dp)

		if dp.Amount > heatmap.Max {
			heatmap.Max = dp.Amount
		}
		if dp.Amount < heatmap.Min {
			heatmap.Min = dp.Amount
		}
	}

	// Sort by date
	sort.Slice(heatmap.Data, func(i, j int) bool {
		return heatmap.Data[i].Date < heatmap.Data[j].Date
	})

	// If no data, reset min
	if heatmap.Min == math.MaxFloat64 {
		heatmap.Min = 0
	}

	return heatmap, nil
}
