package statistic_cmd

import (
	"path/filepath"
	"testing"
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
