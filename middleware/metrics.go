package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsRegistry = prometheus.NewRegistry()
	httpRequests    = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cashlenx",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of API requests.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cashlenx",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "API request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
	httpRequestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "cashlenx",
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Current number of API requests being served.",
	})
)

func init() {
	metricsRegistry.MustRegister(
		httpRequests,
		httpRequestDuration,
		httpRequestsInFlight,
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)
}

// Metrics records API request counts and durations using bounded route-template labels.
func Metrics(router *mux.Router, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		route := matchedRouteTemplate(router, r)
		status := strconv.Itoa(wrapped.statusCode)
		httpRequests.WithLabelValues(r.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route, status).Observe(time.Since(start).Seconds())
	})
}

func matchedRouteTemplate(router *mux.Router, r *http.Request) string {
	match := &mux.RouteMatch{}
	if !router.Match(r, match) || match.Route == nil {
		return "unmatched"
	}
	template, err := match.Route.GetPathTemplate()
	if err != nil || template == "" {
		return "unmatched"
	}
	return template
}

// MetricsHandler exposes application, Go runtime, and process metrics.
func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}
