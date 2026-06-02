package cash_flow_controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		{"total summary", GetTotalSummary, http.MethodGet, "/cash/summary/total", "", nil},
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

func TestCreateExpensePassesAuthenticatedUserAndRequestToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	cashID := primitive.NewObjectID()
	var got struct {
		belongsDate  string
		categoryName string
		amount       float64
		description  string
		userID       string
	}

	original := saveExpenseForUser
	saveExpenseForUser = func(belongsDate, categoryName string, amount float64, description string, serviceUserID string) (model.CashFlowEntity, error) {
		got.belongsDate = belongsDate
		got.categoryName = categoryName
		got.amount = amount
		got.description = description
		got.userID = serviceUserID
		return model.CashFlowEntity{
			Id:           cashID,
			CategoryName: categoryName,
			CategoryType: model.FlowTypeExpense,
			Amount:       amount,
			Description:  description,
			BelongsDate:  time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
		}, nil
	}
	t.Cleanup(func() { saveExpenseForUser = original })

	req := httptest.NewRequest(http.MethodPost, "/cash/expense", strings.NewReader(`{"belongs_date":"2026-05-25","category_name":"Food","amount":12.34,"description":"lunch"}`))
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	rec := httptest.NewRecorder()

	CreateExpense(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got.userID != userID || got.belongsDate != "2026-05-25" || got.categoryName != "Food" || got.amount != 12.34 || got.description != "lunch" {
		t.Fatalf("service args = %+v", got)
	}
}

func TestListAllPassesFiltersAndPaginationToService(t *testing.T) {
	userID := primitive.NewObjectID().Hex()
	var got struct {
		userID           string
		cashType         string
		categoryID       string
		description      string
		exactDescription string
		fromDate         string
		toDate           string
		limit            int
		offset           int
	}

	original := queryCashForUser
	queryCashForUser = func(userID, cashType, categoryID, description, exactDescription, fromDate, toDate string, limit, offset int) ([]*model.CashFlowEntity, int64, error) {
		got.userID = userID
		got.cashType = cashType
		got.categoryID = categoryID
		got.description = description
		got.exactDescription = exactDescription
		got.fromDate = fromDate
		got.toDate = toDate
		got.limit = limit
		got.offset = offset
		return []*model.CashFlowEntity{{Id: primitive.NewObjectID(), Amount: 20, CategoryType: model.FlowTypeIncome}}, 7, nil
	}
	t.Cleanup(func() { queryCashForUser = original })

	req := httptest.NewRequest(http.MethodGet, "/cash?type=income&category_id=cat-id&description=foo&exact_description=bar&from_date=2026-05-01&to_date=2026-05-31&limit=5&page=3", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	rec := httptest.NewRecorder()

	ListAll(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got.userID != userID || got.cashType != "income" || got.categoryID != "cat-id" || got.description != "foo" || got.exactDescription != "bar" || got.fromDate != "2026-05-01" || got.toDate != "2026-05-31" || got.limit != 5 || got.offset != 10 {
		t.Fatalf("service args = %+v", got)
	}

	var response struct {
		Meta map[string]interface{} `json:"meta"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON decode failed: %v", err)
	}
	if response.Meta["total_count"].(float64) != 7 || response.Meta["limit"].(float64) != 5 || response.Meta["offset"].(float64) != 10 {
		t.Fatalf("meta = %+v", response.Meta)
	}
}
