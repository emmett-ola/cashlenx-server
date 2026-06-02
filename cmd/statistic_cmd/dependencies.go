package statistic_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
)

var (
	requireStatisticUser           = cli_auth.RequireUserID
	getStatisticSummaryForUser     = statistic_service.GetSummaryForUser
	getStatisticBreakdownForUser   = statistic_service.GetBreakdownForUser
	getStatisticTrendsForUser      = statistic_service.GetTrendsForUser
	getStatisticTopExpensesForUser = statistic_service.GetTopExpensesForUser
	getStatisticDashboardForUser   = statistic_service.GetDashboardOverviewForUser
	getIncomeExpenseChartForUser   = statistic_service.GetIncomeExpenseChartDataForUser
	getCategoryDistributionForUser = statistic_service.GetCategoryDistributionForUser
	getMonthlyComparisonForUser    = statistic_service.GetMonthlyComparisonForUser
	getSpendingHeatmapForUser      = statistic_service.GetSpendingHeatmapForUser
)
