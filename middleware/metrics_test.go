package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestMetricsRecordsBoundedRouteTemplate(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/api/v0/cash/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}).Methods(http.MethodGet)

	handler := Metrics(router, router)
	req := httptest.NewRequest(http.MethodGet, "/api/v0/cash/507f1f77bcf86cd799439011", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricsRec := httptest.NewRecorder()
	MetricsHandler().ServeHTTP(metricsRec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricsRec.Body.String()
	if !strings.Contains(body, `cashlenx_http_requests_total{method="GET",route="/api/v0/cash/{id}",status="201"}`) {
		t.Fatalf("request metric missing route template: %s", body)
	}
	if strings.Contains(body, "507f1f77bcf86cd799439011") {
		t.Fatal("metrics must not contain concrete user-controlled path values")
	}
	if !strings.Contains(body, "cashlenx_http_request_duration_seconds") {
		t.Fatal("duration histogram is missing")
	}
}
