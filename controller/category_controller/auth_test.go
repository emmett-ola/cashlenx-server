package category_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestCategoryHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
		body    string
		vars    map[string]string
	}{
		{"create", Create, http.MethodPost, "/category", `{"name":"Food","type":"expense"}`, nil},
		{"list", ListAll, http.MethodGet, "/category", "", nil},
		{"tree", Tree, http.MethodGet, "/category/tree", "", nil},
		{"query by id", QueryById, http.MethodGet, "/category/id", "", map[string]string{"id": "id"}},
		{"query by name", QueryByName, http.MethodGet, "/category/name/Food", "", map[string]string{"name": "Food"}},
		{"query children", QueryChildren, http.MethodGet, "/category/id/children", "", map[string]string{"parent_id": "id"}},
		{"update", UpdateById, http.MethodPut, "/category/id", "", map[string]string{"id": "id"}},
		{"delete", DeleteById, http.MethodDelete, "/category/id", "", map[string]string{"id": "id"}},
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

func TestTreeRejectsInvalidCategoryTypeBeforeService(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/category/tree?type=invalid", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-id"))
	rec := httptest.NewRecorder()

	Tree(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCategoryCreateRejectsInvalidJSONBeforeAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/category", strings.NewReader(`{"name":`))
	rec := httptest.NewRecorder()

	Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCategoryUpdateRejectsMissingIDAndInvalidJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		vars map[string]string
	}{
		{name: "missing id", body: `{}`, vars: nil},
		{name: "invalid json", body: `{"name":`, vars: map[string]string{"id": "category-id"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/category/id", strings.NewReader(tc.body))
			req = req.WithContext(contextWithUserID(req.Context(), "user-id"))
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

func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}
