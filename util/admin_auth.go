package util

import (
	"errors"
	"net/http"
	"strings"
)

// AdminTokenError represents an error when ADMIN_TOKEN verification fails
var AdminTokenError = errors.New("invalid or missing ADMIN_TOKEN")

// VerifyAdminToken checks if the provided token matches the configured ADMIN_TOKEN
func VerifyAdminToken(token string) error {
	// Get configured ADMIN_TOKEN from environment
	adminToken := GetConfigByKey("ADMIN_TOKEN")

	// If no ADMIN_TOKEN is configured, we allow the operation for backward compatibility
	// but log a warning
	if adminToken == "" {
		Logger.Warn("ADMIN_TOKEN is not configured, allowing dangerous operation without authentication")
		return nil
	}

	// Check if provided token matches configured token
	if token == "" || token != adminToken {
		Logger.Errorw("Invalid ADMIN_TOKEN provided", "provided_token", token)
		return AdminTokenError
	}

	Logger.Info("ADMIN_TOKEN verified successfully")
	return nil
}

// ExtractAdminTokenFromRequest extracts ADMIN_TOKEN from HTTP request headers or query parameters
func ExtractAdminTokenFromRequest(r *http.Request) string {
	// Check Authorization header first (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Check X-Admin-Token header
	if adminToken := r.Header.Get("X-Admin-Token"); adminToken != "" {
		return adminToken
	}

	// Check admin_token query parameter
	if adminToken := r.URL.Query().Get("admin_token"); adminToken != "" {
		return adminToken
	}

	return ""
}

// VerifyAdminTokenFromRequest verifies ADMIN_TOKEN from HTTP request
func VerifyAdminTokenFromRequest(r *http.Request) error {
	token := ExtractAdminTokenFromRequest(r)
	return VerifyAdminToken(token)
}
