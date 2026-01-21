package provider

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
	"golang.org/x/crypto/bcrypt"
)

// LocalAuthService implements AuthService using local JWT authentication
type LocalAuthService struct{}

// NewLocalAuthService creates a new instance of LocalAuthService
func NewLocalAuthService() *LocalAuthService {
	return &LocalAuthService{}
}

// jwtClaims extends Claims with JWT registered claims
type jwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

// Authenticate validates user credentials and returns a token and user info
func (s *LocalAuthService) Authenticate(username, password string) (string, string, model.UserEntity, error) {
	// Get user by username
	user := user_service.GetUserByUsername(username)
	if user.Id.IsZero() {
		return "", "", model.UserEntity{}, errors.NewUnauthorizedError("invalid username or password")
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", model.UserEntity{}, errors.NewUnauthorizedError("invalid username or password")
	}

	// Generate JWT token
	token, err := s.GenerateToken(user.Id.Hex(), user.Username, user.Role)
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	// Generate refresh token
	refreshToken, err := refresh_token_service.CreateRefreshToken(user.Id.Hex())
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	return token, refreshToken, user, nil
}

// ValidateToken validates a JWT token and returns the claims if valid
func (s *LocalAuthService) ValidateToken(tokenString string) (*Claims, error) {
	jwtSecret := util.GetConfigByKey("auth.jwt.secret")
	if jwtSecret == "" {
		return nil, errors.NewInternalError("JWT secret is not configured", nil)
	}

	// Parse and validate token
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewUnauthorizedError("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.NewUnauthorizedError("invalid or expired token")
	}

	// Check token expiration
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.NewUnauthorizedError("token has expired")
	}

	return &claims.Claims, nil
}

// GenerateToken generates a new JWT token for a user
func (s *LocalAuthService) GenerateToken(userID, username, role string) (string, error) {
	jwtSecret := util.GetConfigByKey("auth.jwt.secret")
	if jwtSecret == "" {
		return "", errors.NewInternalError("JWT secret is not configured", nil)
	}

	// Get JWT expiration minutes from configuration
	expMinutesStr := util.GetConfigByKey("auth.jwt.expiration_minutes")
	expMinutes := 30 // Default to 30 minutes
	if expMinutesStr != "" {
		if parsedMinutes, err := strconv.Atoi(expMinutesStr); err == nil {
			expMinutes = parsedMinutes
		}
	}
	// Set token expiration time
	expirationTime := time.Now().Add(time.Duration(expMinutes) * time.Minute)

	// Create claims
	claims := &jwtClaims{
		Claims: Claims{
			UserID:   userID,
			Username: username,
			Role:     role,
		},
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

// RefreshToken generates a new access token using a refresh token
func (s *LocalAuthService) RefreshToken(refreshToken string) (string, string, model.UserEntity, error) {
	// Validate refresh token
	token, err := refresh_token_service.GetRefreshTokenByToken(refreshToken)
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	// Get user by ID
	user := user_service.GetUserByObjectId(token.UserId)
	if user.Id.IsZero() {
		return "", "", model.UserEntity{}, errors.NewUnauthorizedError("user not found")
	}

	// Revoke old refresh token
	err = refresh_token_service.RevokeRefreshToken(refreshToken, user.Id.Hex())
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	// Generate new access token
	accessToken, err := s.GenerateToken(user.Id.Hex(), user.Username, user.Role)
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	// Generate new refresh token
	newRefreshToken, err := refresh_token_service.CreateRefreshToken(user.Id.Hex())
	if err != nil {
		return "", "", model.UserEntity{}, err
	}

	return accessToken, newRefreshToken, user, nil
}

// GetUserFromToken extracts user information from a token
func (s *LocalAuthService) GetUserFromToken(tokenString string) (model.UserEntity, error) {
	// Validate the token
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return model.UserEntity{}, err
	}

	// Get user by ID
	user := user_service.GetUserByObjectId(claims.UserID)
	if user.Id.IsZero() {
		return model.UserEntity{}, errors.NewUnauthorizedError("user not found")
	}

	return user, nil
}

// Middleware returns the authentication middleware for the service
func (s *LocalAuthService) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authentication is always enabled

		// Skip authentication for all public endpoints under /api/open/*
		path := r.URL.Path
		if strings.HasPrefix(path, "/api/open/") {
			next.ServeHTTP(w, r)
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

		// Validate token
		claims, err := s.ValidateToken(tokenString)
		if err != nil {
			util.Logger.Errorw("Invalid JWT token", "error", err)
			util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("Invalid or expired token"))
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
