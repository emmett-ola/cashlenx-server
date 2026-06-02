package statistic_cmd

import (
	"path/filepath"
	"testing"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
)

func TestStatisticCommandsRequireLoggedInUser(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "summary", run: func() error { return summaryCmd.RunE(summaryCmd, nil) }},
		{name: "breakdown", run: func() error { return breakdownCmd.RunE(breakdownCmd, nil) }},
		{name: "trends", run: func() error { return trendsCmd.RunE(trendsCmd, nil) }},
		{name: "top", run: func() error { return topCmd.RunE(topCmd, nil) }},
		{name: "dashboard", run: func() error { return dashboardCmd.RunE(dashboardCmd, nil) }},
		{name: "export", run: func() error { return exportCmd.RunE(exportCmd, nil) }},
		{name: "import", run: func() error { return importCmd.RunE(importCmd, nil) }},
		{name: "income expense chart", run: func() error { return incomeExpenseChartCmd.RunE(incomeExpenseChartCmd, nil) }},
		{name: "category distribution chart", run: func() error { return categoryDistributionChartCmd.RunE(categoryDistributionChartCmd, nil) }},
		{name: "monthly comparison chart", run: func() error { return monthlyComparisonChartCmd.RunE(monthlyComparisonChartCmd, nil) }},
		{name: "spending heatmap chart", run: func() error { return spendingHeatmapChartCmd.RunE(spendingHeatmapChartCmd, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetStatisticUserIDs()
			if err := tt.run(); err == nil {
				t.Fatal("expected missing CLI session error")
			}
		})
	}
}

func resetStatisticUserIDs() {
	summaryUserId = ""
	breakdownUserId = ""
	trendsUserId = ""
	topUserId = ""
	dashboardUserId = ""
	exportUserId = ""
	importUserId = ""
	incomeExpenseUserId = ""
	categoryDistributionUserId = ""
	monthlyComparisonUserId = ""
	spendingHeatmapUserId = ""
}

func TestStatisticSummaryBreakdownTrendsTopAndDashboardPassFlagsToServices(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	type periodCall struct {
		period string
		date   string
		userID string
	}

	originalRequire := requireStatisticUser
	requireStatisticUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	var summaryArgs, breakdownArgs, trendsArgs, dashboardArgs periodCall
	var topArgs struct {
		limit  int
		period string
		date   string
		userID string
	}
	originalSummary := getStatisticSummaryForUser
	getStatisticSummaryForUser = func(period, date, userID string) (*statistic_service.Summary, error) {
		summaryArgs.period = period
		summaryArgs.date = date
		summaryArgs.userID = userID
		return &statistic_service.Summary{Period: date, PeriodType: period, Income: 100, IncomeCount: 1, TransactionCount: 1, Categories: map[string]float64{"Salary": 100}}, nil
	}
	originalBreakdown := getStatisticBreakdownForUser
	getStatisticBreakdownForUser = func(period, date, userID string) (*statistic_service.Breakdown, error) {
		breakdownArgs.period = period
		breakdownArgs.date = date
		breakdownArgs.userID = userID
		return &statistic_service.Breakdown{
			Period:       date,
			TotalExpense: 25,
			ExpenseCategories: []statistic_service.CategoryBreakdownItem{
				{Category: "Food", Amount: 25, Percentage: 100, Count: 1},
			},
		}, nil
	}
	originalTrends := getStatisticTrendsForUser
	getStatisticTrendsForUser = func(period, date, userID string) (*statistic_service.Trends, error) {
		trendsArgs.period = period
		trendsArgs.date = date
		trendsArgs.userID = userID
		return &statistic_service.Trends{
			Period: date, PeriodType: period,
			DataPoints: []statistic_service.TrendDataPoint{{Date: "2026-01", Income: 100, Expense: 25, Balance: 75}},
			Trends:     statistic_service.TrendAnalysis{IncomeTrend: "up", ExpenseTrend: "flat", AverageMonthlyExpense: 25},
		}, nil
	}
	originalTop := getStatisticTopExpensesForUser
	getStatisticTopExpensesForUser = func(limit int, period, date, userID string) (*statistic_service.TopExpenses, error) {
		topArgs.limit = limit
		topArgs.period = period
		topArgs.date = date
		topArgs.userID = userID
		return &statistic_service.TopExpenses{
			Period: date, Limit: limit, TotalExpense: 30,
			Expenses: []statistic_service.TopExpense{{Date: "2026-01-02", Category: "Food", Amount: 30, Percentage: 100, Description: "weekly groceries"}},
		}, nil
	}
	originalDashboard := getStatisticDashboardForUser
	getStatisticDashboardForUser = func(period, date, userID string) (*statistic_service.DashboardOverview, error) {
		dashboardArgs.period = period
		dashboardArgs.date = date
		dashboardArgs.userID = userID
		return &statistic_service.DashboardOverview{
			Period: date, PeriodType: period,
			Summary:     &statistic_service.Summary{Income: 100, IncomeCount: 1, Expense: 25, ExpenseCount: 1, Balance: 75, TransactionCount: 2},
			QuickStats:  statistic_service.QuickStats{TotalTransactions: 2, AverageDaily: 10, HighestExpense: 25, LowestExpense: 5},
			RecentTrend: "flat",
			TopCategories: []statistic_service.CategoryBreakdownItem{
				{Category: "Food", Amount: 25, Percentage: 100, Count: 1},
			},
		}, nil
	}
	t.Cleanup(func() {
		requireStatisticUser = originalRequire
		getStatisticSummaryForUser = originalSummary
		getStatisticBreakdownForUser = originalBreakdown
		getStatisticTrendsForUser = originalTrends
		getStatisticTopExpensesForUser = originalTop
		getStatisticDashboardForUser = originalDashboard
		resetStatisticUserIDs()
		resetStatisticFlagValues()
	})

	summaryPeriod, summaryDate = "monthly", "2026-01"
	if err := summaryCmd.RunE(summaryCmd, nil); err != nil {
		t.Fatalf("summary RunE returned error: %v", err)
	}
	breakdownPeriod, breakdownDate = "yearly", "2026"
	if err := breakdownCmd.RunE(breakdownCmd, nil); err != nil {
		t.Fatalf("breakdown RunE returned error: %v", err)
	}
	trendsPeriod, trendsDate = "yearly", "2026"
	if err := trendsCmd.RunE(trendsCmd, nil); err != nil {
		t.Fatalf("trends RunE returned error: %v", err)
	}
	topLimit, topPeriod, topDate = 7, "monthly", "2026-01"
	if err := topCmd.RunE(topCmd, nil); err != nil {
		t.Fatalf("top RunE returned error: %v", err)
	}
	dashboardPeriod, dashboardDate = "monthly", "2026-01"
	if err := dashboardCmd.RunE(dashboardCmd, nil); err != nil {
		t.Fatalf("dashboard RunE returned error: %v", err)
	}

	if summaryArgs != (periodCall{"monthly", "2026-01", authenticatedUserID}) {
		t.Fatalf("summary args = %+v", summaryArgs)
	}
	if breakdownArgs != (periodCall{"yearly", "2026", authenticatedUserID}) {
		t.Fatalf("breakdown args = %+v", breakdownArgs)
	}
	if trendsArgs != (periodCall{"yearly", "2026", authenticatedUserID}) {
		t.Fatalf("trends args = %+v", trendsArgs)
	}
	if topArgs.limit != 7 || topArgs.period != "monthly" || topArgs.date != "2026-01" || topArgs.userID != authenticatedUserID {
		t.Fatalf("top args = %+v", topArgs)
	}
	if dashboardArgs != (periodCall{"monthly", "2026-01", authenticatedUserID}) {
		t.Fatalf("dashboard args = %+v", dashboardArgs)
	}
}

func TestStatisticChartCommandsPassFlagsToServices(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"

	originalRequire := requireStatisticUser
	requireStatisticUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	var incomeExpenseArgs, distributionArgs struct {
		period string
		date   string
		typ    string
		userID string
	}
	var monthlyArgs, heatmapArgs struct {
		year   string
		userID string
	}
	originalIncomeExpense := getIncomeExpenseChartForUser
	getIncomeExpenseChartForUser = func(period, date, userID string) (*statistic_service.IncomeExpenseChartData, error) {
		incomeExpenseArgs.period = period
		incomeExpenseArgs.date = date
		incomeExpenseArgs.userID = userID
		return &statistic_service.IncomeExpenseChartData{
			Labels: []string{"Jan"}, Income: []float64{100}, Expense: []float64{25}, Balance: []float64{75}, Period: period, FromDate: "2026-01-01", ToDate: "2026-01-31",
		}, nil
	}
	originalDistribution := getCategoryDistributionForUser
	getCategoryDistributionForUser = func(period, date, flowType, userID string) (*statistic_service.CategoryDistribution, error) {
		distributionArgs.period = period
		distributionArgs.date = date
		distributionArgs.typ = flowType
		distributionArgs.userID = userID
		return &statistic_service.CategoryDistribution{Labels: []string{"Food"}, Values: []float64{25}, Percentages: []float64{100}, Colors: []string{"#000000"}, Total: 25, Type: flowType}, nil
	}
	originalMonthly := getMonthlyComparisonForUser
	getMonthlyComparisonForUser = func(year, userID string) (*statistic_service.MonthlyComparison, error) {
		monthlyArgs.year = year
		monthlyArgs.userID = userID
		return &statistic_service.MonthlyComparison{Year: year, Months: []string{"Jan"}, Income: []float64{100}, Expense: []float64{25}, Balance: []float64{75}}, nil
	}
	originalHeatmap := getSpendingHeatmapForUser
	getSpendingHeatmapForUser = func(year, userID string) (*statistic_service.SpendingHeatmap, error) {
		heatmapArgs.year = year
		heatmapArgs.userID = userID
		return &statistic_service.SpendingHeatmap{Year: year, Data: []statistic_service.HeatmapDataPoint{{Date: "2026-01-02", Amount: 25, Count: 1}}, Min: 0, Max: 25}, nil
	}
	t.Cleanup(func() {
		requireStatisticUser = originalRequire
		getIncomeExpenseChartForUser = originalIncomeExpense
		getCategoryDistributionForUser = originalDistribution
		getMonthlyComparisonForUser = originalMonthly
		getSpendingHeatmapForUser = originalHeatmap
		resetStatisticUserIDs()
		resetStatisticFlagValues()
	})

	incomeExpensePeriod, incomeExpenseDate = "monthly", "2026-01"
	if err := incomeExpenseChartCmd.RunE(incomeExpenseChartCmd, nil); err != nil {
		t.Fatalf("income-expense RunE returned error: %v", err)
	}
	categoryDistributionPeriod, categoryDistributionDate, categoryDistributionType = "monthly", "2026-01", "expense"
	if err := categoryDistributionChartCmd.RunE(categoryDistributionChartCmd, nil); err != nil {
		t.Fatalf("category-distribution RunE returned error: %v", err)
	}
	monthlyComparisonYear = "2026"
	if err := monthlyComparisonChartCmd.RunE(monthlyComparisonChartCmd, nil); err != nil {
		t.Fatalf("monthly-comparison RunE returned error: %v", err)
	}
	spendingHeatmapYear = "2026"
	if err := spendingHeatmapChartCmd.RunE(spendingHeatmapChartCmd, nil); err != nil {
		t.Fatalf("spending-heatmap RunE returned error: %v", err)
	}

	if incomeExpenseArgs.period != "monthly" || incomeExpenseArgs.date != "2026-01" || incomeExpenseArgs.userID != authenticatedUserID {
		t.Fatalf("income expense args = %+v", incomeExpenseArgs)
	}
	if distributionArgs.period != "monthly" || distributionArgs.date != "2026-01" || distributionArgs.typ != "expense" || distributionArgs.userID != authenticatedUserID {
		t.Fatalf("distribution args = %+v", distributionArgs)
	}
	if monthlyArgs.year != "2026" || monthlyArgs.userID != authenticatedUserID {
		t.Fatalf("monthly args = %+v", monthlyArgs)
	}
	if heatmapArgs.year != "2026" || heatmapArgs.userID != authenticatedUserID {
		t.Fatalf("heatmap args = %+v", heatmapArgs)
	}
}

func resetStatisticFlagValues() {
	summaryPeriod, summaryDate = "monthly", ""
	breakdownPeriod, breakdownDate = "monthly", ""
	trendsPeriod, trendsDate = "yearly", ""
	topLimit, topPeriod, topDate = 10, "monthly", ""
	dashboardPeriod, dashboardDate = "monthly", ""
	incomeExpensePeriod, incomeExpenseDate = "monthly", ""
	categoryDistributionPeriod, categoryDistributionDate, categoryDistributionType = "monthly", "", "expense"
	monthlyComparisonYear = ""
	spendingHeatmapYear = ""
}
