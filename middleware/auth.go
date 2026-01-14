package middleware

import (
	"context"
	std_errors "errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

// JWT claims structure
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Auth middleware that handles JWT authentication and role-based access control
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is always enabled

		// Skip authentication for all public endpoints under /api/open/*
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/open/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get JWT secret from environment
		jwtSecret := util.GetConfigByKey("auth.jwt.secret")
		if jwtSecret == "" {
			util.Logger.Error("JWT_SECRET is not configured but AUTH_ENABLED is true")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("Authorization header is required"))
			return
		}

		// Check if header starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("Invalid authorization header format"))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, std_errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			util.Logger.Errorw("Invalid JWT token", "error", err)
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("Invalid or expired token"))
			return
		}

		// Check token expiration
		if claims.ExpiresAt.Before(time.Now()) {
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("Token has expired"))
			return
		}

		// Add user information to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)
		r = r.WithContext(ctx)

		// Check if admin role is required for admin-only routes
		if strings.HasPrefix(path, "/api/admin/") {
			if claims.Role != "admin" {
				util.ComposeJSONResponse(w, http.StatusForbidden, errors.NewForbiddenError("admin role required"))
				return
			}
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID, username, role string) (string, error) {
	jwtSecret := util.GetConfigByKey("auth.jwt.secret")
	if jwtSecret == "" {
		return "", std_errors.New("JWT_SECRET is not configured")
	}

	// Set token expiration time
	expirationTime := time.Now().Add(model.JWTExpirationHours * time.Hour)

	// Create claims
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "cashlenx-server",
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Admin middleware that checks if the user has admin role
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is always enabled

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
