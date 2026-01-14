package middleware

import (
	"net/http"
	"strings"

	"github.com/macar-x/cashlenx-server/util"
)

// CORS middleware to handle Cross-Origin Resource Sharing
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from configuration or use defaults
		allowedOrigins := util.GetConfigByKey("cors.origins")
		if allowedOrigins == "" {
			// Default allowed origins for development
			allowedOrigins = "http://localhost:3000,http://localhost:8080,http://localhost:4000,https://localhost:3000,https://localhost:8080"
		}

		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" {
			origins := strings.Split(allowedOrigins, ",")
			for _, allowedOrigin := range origins {
				if strings.TrimSpace(allowedOrigin) == origin || allowedOrigin == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
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
