package cash_flow_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestCashHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		vars    map[string]string
	}{
		{"create expense", CreateExpense, http.MethodPost, "/cash/expense", `{"category_name":"Food","amount":10}`, nil},
		{"create income", CreateIncome, http.MethodPost, "/cash/income", `{"category_name":"Salary","amount":10}`, nil},
		{"list", ListAll, http.MethodGet, "/cash", "", nil},
		{"range", QueryByDateRange, http.MethodGet, "/cash/range", "", nil},
		{"query by date", QueryByDate, http.MethodGet, "/cash/date/2026-05-25", "", map[string]string{"date": "2026-05-25"}},
		{"query by id", QueryById, http.MethodGet, "/cash/id", "", map[string]string{"id": "id"}},
		{"update by id", UpdateById, http.MethodPut, "/cash/id", "", map[string]string{"id": "id"}},
		{"delete by id", DeleteById, http.MethodDelete, "/cash/id", "", map[string]string{"id": "id"}},
		{"delete by date", DeleteByDate, http.MethodDelete, "/cash/date/2026-05-25", "", map[string]string{"date": "2026-05-25"}},
		{"daily summary", GetDailySummary, http.MethodGet, "/cash/summary/daily/2026-05-25", "", map[string]string{"date": "2026-05-25"}},
		{"monthly summary", GetMonthlySummary, http.MethodGet, "/cash/summary/monthly/2026-05", "", map[string]string{"month": "2026-05"}},
		{"yearly summary", GetYearlySummary, http.MethodGet, "/cash/summary/yearly/2026", "", map[string]string{"year": "2026"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
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

func TestCashCreateRejectsInvalidInputBeforeAuth(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		body    string
	}{
		{"expense invalid json", CreateExpense, `{"category_name":`},
		{"expense missing required fields", CreateExpense, `{}`},
		{"income invalid json", CreateIncome, `{"category_name":`},
		{"income missing required fields", CreateIncome, `{}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/cash", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}

func TestCashUpdateRejectsMissingIDAndInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		vars map[string]string
	}{
		{name: "missing id", body: `{}`, vars: nil},
		{name: "invalid json", body: `{"amount":`, vars: map[string]string{"id": "cash-id"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/cash/id", strings.NewReader(tc.body))
			req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-id"))
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rec := httptest.NewRecorder()

			UpdateById(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		})
	}
}
