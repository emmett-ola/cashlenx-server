package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/util"
)

func TestHealthCheckReturnsHealthyResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v0/open/health", nil)
	rec := httptest.NewRecorder()

	healthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body util.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("response code = %q, want OK", body.Code)
	}
	if body.Message != "API is running" {
		t.Fatalf("message = %q, want API is running", body.Message)
	}
}

func TestVersionInfoIncludesCurrentRouteGroups(t *testing.T) {
	originalVersion := util.GetConfigByKey("api.version")
	defer util.SetConfigByKey("api.version", originalVersion)
	util.SetConfigByKey("api.version", "vtest")

	req := httptest.NewRequest(http.MethodGet, "/api/vtest/open/version", nil)
	rec := httptest.NewRecorder()

	versionInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	body := decodeResponseData[map[string]interface{}](t, rec)
	endpoints, ok := body["endpoints"].(map[string]interface{})
	if !ok {
		t.Fatalf("endpoints payload has type %T", body["endpoints"])
	}
	assertEndpointListed(t, endpoints, "open", "POST /api/vtest/open/auth/logout")
	assertEndpointListed(t, endpoints, "auth", "GET /api/vtest/auth/tokens")
	assertEndpointListed(t, endpoints, "admin", "PUT /api/vtest/admin/user/{id}")
	assertEndpointListed(t, endpoints, "user", "GET /api/vtest/user/configuration")
}

func TestRouteRegistrationMatchesExpectedEndpoints(t *testing.T) {
	r := mux.NewRouter()
	registerOpenRoutes(r, "/api/v0")
	adminRouter := r.PathPrefix("/api/v0/admin").Subrouter()
	registerAdminRoutes(adminRouter)
	registerUserRoutes(r, "/api/v0")
	registerCashRoute(r, "/api/v0")
	registerCategoryRoute(r, "/api/v0")
	registerStatisticRoute(r, "/api/v0")

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v0/open/auth/logout"},
		{http.MethodGet, "/api/v0/auth/tokens"},
		{http.MethodPut, "/api/v0/admin/user/123"},
		{http.MethodGet, "/api/v0/user/configuration"},
		{http.MethodDelete, "/api/v0/cash/date/2026-05-25"},
		{http.MethodGet, "/api/v0/category/parent-id/children"},
		{http.MethodGet, "/api/v0/statistic/chart/spending-heatmap/2026"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		match := &mux.RouteMatch{}
		if !r.Match(req, match) {
			t.Fatalf("%s %s did not match registered routes", tc.method, tc.path)
		}
	}
}

func decodeResponseData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var body struct {
		Code string          `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response wrapper: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("response code = %q, want OK; body=%s", body.Code, rec.Body.String())
	}
	var data T
	if err := json.Unmarshal(body.Data, &data); err != nil {
		t.Fatalf("failed to decode response data: %v", err)
	}
	return data
}

func assertEndpointListed(t *testing.T, endpoints map[string]interface{}, group string, expected string) {
	t.Helper()
	values, ok := endpoints[group].([]interface{})
	if !ok {
		t.Fatalf("endpoint group %q has type %T", group, endpoints[group])
	}
	for _, value := range values {
		if value == expected {
			return
		}
	}
	t.Fatalf("endpoint %q not listed in group %q: %#v", expected, group, values)
}
