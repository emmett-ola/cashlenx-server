package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPeerRateLimiterUsesBurstAndRefills(t *testing.T) {
	now := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	limiter := &peerRateLimiter{
		buckets:       make(map[string]*peerBucket),
		tokensPerNano: 2 / float64(time.Minute),
		burst:         2,
		idleTTL:       2 * time.Minute,
		now:           func() time.Time { return now },
	}
	handler := limiter.wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for requestNumber := 1; requestNumber <= 3; requestNumber++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v0/open/health", nil)
		request.RemoteAddr = "192.0.2.10:12345"
		handler.ServeHTTP(recorder, request)
		want := http.StatusNoContent
		if requestNumber == 3 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("request %d status = %d, want %d", requestNumber, recorder.Code, want)
		}
	}

	now = now.Add(30 * time.Second)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v0/open/health", nil)
	request.RemoteAddr = "192.0.2.10:54321"
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("refilled request status = %d", recorder.Code)
	}
}
