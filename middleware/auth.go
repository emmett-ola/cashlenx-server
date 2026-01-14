package middleware

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

// Auth middleware that uses the configured auth service for authentication
func Auth(next http.Handler) http.Handler {
	// Delegate to the auth service's middleware
	return auth.Service.Middleware(next)
}

// GenerateToken generates a new JWT token using the configured auth service
func GenerateToken(userID, username, role string) (string, error) {
	return auth.Service.GenerateToken(userID, username, role)
}

// Admin middleware that checks if the user has admin role
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user role from context (set by Auth middleware)
		roleValue := r.Context().Value("role")
		if roleValue == nil {
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user role not found in context"))
			return
		}

		role, ok := roleValue.(string)
		if !ok || role != "admin" {
			util.ComposeJSONResponse(w, http.StatusForbidden, errors.NewForbiddenError("admin role required"))
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
