package budget_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
)

func TestBudgetHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name               string
		handler            http.HandlerFunc
		method, path, body string
	}{
		{"create", Create, http.MethodPost, "/budget", `{}`},
		{"list", List, http.MethodGet, "/budget", ""},
		{"get", Get, http.MethodGet, "/budget/id", ""},
		{"update", Update, http.MethodPut, "/budget/id", `{}`},
		{"delete", Delete, http.MethodDelete, "/budget/id", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			tc.handler(recorder, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
			}
		})
	}
}

func TestCreatePassesAuthenticatedUserAndRequest(t *testing.T) {
	original := createBudget
	defer func() { createBudget = original }()
	var got model.UpsertBudgetRequest
	var gotUser string
	createBudget = func(request model.UpsertBudgetRequest, userID string) (model.BudgetView, error) {
		got, gotUser = request, userID
		return model.BudgetView{Id: "budget-id", CategoryId: request.CategoryId, Period: request.Period, LimitAmount: request.LimitAmount}, nil
	}
	request := httptest.NewRequest(http.MethodPost, "/budget", strings.NewReader(`{"category_id":"category-id","period":"2026-08","limit_amount":1200}`))
	request = request.WithContext(context.WithValue(request.Context(), "user_id", "user-id"))
	recorder := httptest.NewRecorder()
	Create(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if gotUser != "user-id" || got.CategoryId != "category-id" || got.Period != "2026-08" || got.LimitAmount != 1200 {
		t.Fatalf("service args = %#v, %q", got, gotUser)
	}
}

func TestCreateRejectsInvalidJSONAfterAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/budget", strings.NewReader(`{"period":`))
	request = request.WithContext(context.WithValue(request.Context(), "user_id", "user-id"))
	recorder := httptest.NewRecorder()
	Create(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}
