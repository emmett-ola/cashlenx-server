package statistic_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestStatisticHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		vars    map[string]string
	}{
		{"export", ExportData, http.MethodGet, "/statistic/export", nil},
		{"import", ImportData, http.MethodPost, "/statistic/import", nil},
		{"daily summary", GetDailySummary, http.MethodGet, "/statistic/summary/daily/2026-05-25", map[string]string{"date": "2026-05-25"}},
		{"monthly summary", GetMonthlySummary, http.MethodGet, "/statistic/summary/monthly/2026-05", map[string]string{"month": "2026-05"}},
		{"yearly summary", GetYearlySummary, http.MethodGet, "/statistic/summary/yearly/2026", map[string]string{"year": "2026"}},
		{"daily breakdown", GetDailyBreakdown, http.MethodGet, "/statistic/breakdown/daily/2026-05-25", map[string]string{"date": "2026-05-25"}},
		{"monthly breakdown", GetMonthlyBreakdown, http.MethodGet, "/statistic/breakdown/monthly/2026-05", map[string]string{"month": "2026-05"}},
		{"yearly breakdown", GetYearlyBreakdown, http.MethodGet, "/statistic/breakdown/yearly/2026", map[string]string{"year": "2026"}},
		{"daily trends", GetDailyTrends, http.MethodGet, "/statistic/trends/daily/2026-05-25", map[string]string{"date": "2026-05-25"}},
		{"monthly trends", GetMonthlyTrends, http.MethodGet, "/statistic/trends/monthly/2026-05", map[string]string{"month": "2026-05"}},
		{"yearly trends", GetYearlyTrends, http.MethodGet, "/statistic/trends/yearly/2026", map[string]string{"year": "2026"}},
		{"daily top", GetDailyTopExpenses, http.MethodGet, "/statistic/top/daily/2026-05-25", map[string]string{"date": "2026-05-25"}},
		{"monthly top", GetMonthlyTopExpenses, http.MethodGet, "/statistic/top/monthly/2026-05", map[string]string{"month": "2026-05"}},
		{"yearly top", GetYearlyTopExpenses, http.MethodGet, "/statistic/top/yearly/2026", map[string]string{"year": "2026"}},
		{"dashboard", GetDashboardOverview, http.MethodGet, "/statistic/dashboard/monthly/2026-05", map[string]string{"period": "monthly", "date": "2026-05"}},
		{"income expense chart", GetIncomeExpenseChartData, http.MethodGet, "/statistic/chart/income-expense/monthly/2026-05", map[string]string{"period": "monthly", "date": "2026-05"}},
		{"category distribution", GetCategoryDistributionData, http.MethodGet, "/statistic/chart/category-distribution/monthly/2026-05", map[string]string{"period": "monthly", "date": "2026-05"}},
		{"monthly comparison", GetMonthlyComparisonData, http.MethodGet, "/statistic/chart/monthly-comparison/2026", map[string]string{"year": "2026"}},
		{"spending heatmap", GetSpendingHeatmapData, http.MethodGet, "/statistic/chart/spending-heatmap/2026", map[string]string{"year": "2026"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestStatisticHandlersRejectMissingPathParameters(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		path    string
	}{
		{"daily summary", GetDailySummary, "/statistic/summary/daily/"},
		{"monthly summary", GetMonthlySummary, "/statistic/summary/monthly/"},
		{"yearly summary", GetYearlySummary, "/statistic/summary/yearly/"},
		{"daily breakdown", GetDailyBreakdown, "/statistic/breakdown/daily/"},
		{"monthly breakdown", GetMonthlyBreakdown, "/statistic/breakdown/monthly/"},
		{"yearly breakdown", GetYearlyBreakdown, "/statistic/breakdown/yearly/"},
		{"daily trends", GetDailyTrends, "/statistic/trends/daily/"},
		{"monthly trends", GetMonthlyTrends, "/statistic/trends/monthly/"},
		{"yearly trends", GetYearlyTrends, "/statistic/trends/yearly/"},
		{"daily top", GetDailyTopExpenses, "/statistic/top/daily/"},
		{"monthly top", GetMonthlyTopExpenses, "/statistic/top/monthly/"},
		{"yearly top", GetYearlyTopExpenses, "/statistic/top/yearly/"},
		{"dashboard", GetDashboardOverview, "/statistic/dashboard/"},
		{"income expense chart", GetIncomeExpenseChartData, "/statistic/chart/income-expense/"},
		{"category distribution", GetCategoryDistributionData, "/statistic/chart/category-distribution/"},
		{"monthly comparison", GetMonthlyComparisonData, "/statistic/chart/monthly-comparison/"},
		{"spending heatmap", GetSpendingHeatmapData, "/statistic/chart/spending-heatmap/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-id"))
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestStatisticExportRejectsInvalidFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/statistic/export?format=xml", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-id"))
	rec := httptest.NewRecorder()

	ExportData(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
