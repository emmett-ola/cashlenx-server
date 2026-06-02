package statistic_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
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

func TestStatisticExportImportAndTopPassInputsToServices(t *testing.T) {
	const userID = "507f1f77bcf86cd799439011"
	var exportArgs struct {
		fromDate string
		toDate   string
		filePath string
		userID   string
	}
	var importArgs struct {
		filePath string
		userID   string
	}
	var topCalls []struct {
		limit  int
		period string
		date   string
		userID string
	}

	originalCSV := exportStatisticCSVForUser
	exportStatisticCSVForUser = func(fromDate, toDate, filePath, serviceUserID string) error {
		exportArgs.fromDate = fromDate
		exportArgs.toDate = toDate
		exportArgs.filePath = filePath
		exportArgs.userID = serviceUserID
		return os.WriteFile(filePath, []byte("id,amount\n1,10\n"), 0600)
	}
	originalImport := importStatisticForUser
	importStatisticForUser = func(filePath, serviceUserID string) error {
		importArgs.filePath = filePath
		importArgs.userID = serviceUserID
		return nil
	}
	originalTop := getTopExpensesForUser
	getTopExpensesForUser = func(limit int, period, date, serviceUserID string) (*statistic_service.TopExpenses, error) {
		topCalls = append(topCalls, struct {
			limit  int
			period string
			date   string
			userID string
		}{limit: limit, period: period, date: date, userID: serviceUserID})
		return &statistic_service.TopExpenses{
			Period: date, Limit: limit, TotalExpense: 10,
			Expenses: []statistic_service.TopExpense{{Date: date, Category: "Food", Amount: 10, Percentage: 100}},
		}, nil
	}
	t.Cleanup(func() {
		exportStatisticCSVForUser = originalCSV
		importStatisticForUser = originalImport
		getTopExpensesForUser = originalTop
	})

	exportReq := httptest.NewRequest(http.MethodGet, "/statistic/export?format=csv&from_date=2026-05-01&to_date=2026-05-31", nil)
	exportReq = exportReq.WithContext(context.WithValue(exportReq.Context(), "user_id", userID))
	exportRec := httptest.NewRecorder()
	ExportData(exportRec, exportReq)
	if exportRec.Code != http.StatusOK {
		t.Fatalf("export status = %d, want %d; body=%s", exportRec.Code, http.StatusOK, exportRec.Body.String())
	}
	if exportArgs.fromDate != "2026-05-01" || exportArgs.toDate != "2026-05-31" || exportArgs.userID != userID || exportArgs.filePath == "" {
		t.Fatalf("export args = %+v", exportArgs)
	}
	if exportRec.Header().Get("Content-Type") != "text/csv" {
		t.Fatalf("content type = %q", exportRec.Header().Get("Content-Type"))
	}

	importReq := httptest.NewRequest(http.MethodPost, "/statistic/import?file_path=/tmp/import.csv", nil)
	importReq = importReq.WithContext(context.WithValue(importReq.Context(), "user_id", userID))
	importRec := httptest.NewRecorder()
	ImportData(importRec, importReq)
	if importRec.Code != http.StatusOK {
		t.Fatalf("import status = %d, want %d; body=%s", importRec.Code, http.StatusOK, importRec.Body.String())
	}
	if importArgs.filePath != "/tmp/import.csv" || importArgs.userID != userID {
		t.Fatalf("import args = %+v", importArgs)
	}

	cases := []struct {
		handler http.HandlerFunc
		path    string
		vars    map[string]string
	}{
		{GetDailyTopExpenses, "/statistic/top/daily/2026-05-25?limit=3", map[string]string{"date": "2026-05-25"}},
		{GetMonthlyTopExpenses, "/statistic/top/monthly/2026-05?limit=4", map[string]string{"month": "2026-05"}},
		{GetYearlyTopExpenses, "/statistic/top/yearly/2026?limit=5", map[string]string{"year": "2026"}},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
		req = mux.SetURLVars(req, tc.vars)
		rec := httptest.NewRecorder()
		tc.handler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("top status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
	}
	want := []struct {
		limit  int
		period string
		date   string
		userID string
	}{
		{3, "daily", "2026-05-25", userID},
		{4, "monthly", "2026-05", userID},
		{5, "yearly", "2026", userID},
	}
	if len(topCalls) != len(want) {
		t.Fatalf("top calls = %+v", topCalls)
	}
	for i := range want {
		if topCalls[i] != want[i] {
			t.Fatalf("top call[%d] = %+v, want %+v", i, topCalls[i], want[i])
		}
	}
}
