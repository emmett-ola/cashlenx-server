package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/macar-x/cashlenx-server/util"
)

const defaultDevAllowedOrigins = "http://localhost:*,http://127.0.0.1:*,https://localhost:*,https://127.0.0.1:*"

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from configuration or use defaults
		allowedOrigins := util.GetConfigByKey("cors.origins")
		if allowedOrigins == "" {
			allowedOrigins = defaultDevAllowedOrigins
		}

		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && shouldAllowOrigin(origin, allowedOrigins, util.GetConfigByKey("env")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func shouldAllowOrigin(origin string, allowedOrigins string, env string) bool {
	if isAllowedOrigin(origin, allowedOrigins) {
		return true
	}

	// Flutter web and browser-based local clients often use random localhost ports.
	// Keep that ergonomic in dev/test, but require explicit origins outside those modes.
	if env == "" || env == "dev" || env == "test" {
		return isLoopbackOrigin(origin)
	}

	return false
}

func isAllowedOrigin(origin string, allowedOrigins string) bool {
	origins := strings.Split(allowedOrigins, ",")
	for _, allowedOrigin := range origins {
		if originMatchesRule(origin, strings.TrimSpace(allowedOrigin)) {
			return true
		}
	}

	return false
}

func originMatchesRule(origin string, allowedOrigin string) bool {
	if allowedOrigin == "" {
		return false
	}

	if allowedOrigin == "*" || allowedOrigin == origin {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Support patterns like http://localhost:* or https://127.0.0.1:* for dev clients.
	if strings.HasSuffix(allowedOrigin, ":*") {
		allowedBase := strings.TrimSuffix(allowedOrigin, ":*")
		allowedURL, err := url.Parse(allowedBase)
		if err != nil {
			return false
		}

		if originURL.Scheme != allowedURL.Scheme {
			return false
		}

		allowedHost := allowedURL.Hostname()
		originHost := originURL.Hostname()
		return sameLoopbackHost(originHost, allowedHost)
	}

	return false
}

func isLoopbackOrigin(origin string) bool {
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if originURL.Scheme != "http" && originURL.Scheme != "https" {
		return false
	}

	host := originURL.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func sameLoopbackHost(originHost string, allowedHost string) bool {
	if strings.EqualFold(originHost, allowedHost) {
		return true
	}

	originIP := net.ParseIP(originHost)
	allowedIP := net.ParseIP(allowedHost)
	if originIP != nil && allowedIP != nil {
		return originIP.Equal(allowedIP)
	}

	return false
}
