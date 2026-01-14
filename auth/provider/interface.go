package provider

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/model"
)

// AuthService defines the interface for authentication services
type AuthService interface {
	// Authenticate validates user credentials and returns a token and user info
	Authenticate(username, password string) (string, model.UserEntity, error)

	// ValidateToken validates a JWT token and returns the claims if valid
	ValidateToken(tokenString string) (*Claims, error)

	// GenerateToken generates a new JWT token for a user
	GenerateToken(userID, username, role string) (string, error)

	// GetUserFromToken extracts user information from a token
	GetUserFromToken(tokenString string) (model.UserEntity, error)

	// Middleware returns the authentication middleware for the service
	Middleware(next http.Handler) http.Handler
}

// Claims represents the JWT claims structure used across auth services
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
