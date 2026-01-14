package service

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/auth/provider"
	"github.com/macar-x/cashlenx-server/model"
)

// AuthService defines the main authentication service interface
type AuthService interface {
	// Authenticate validates user credentials and returns tokens and user info
	Authenticate(username, password string) (string, string, model.UserEntity, error)

	// ValidateToken validates a JWT token and returns the claims if valid
	ValidateToken(tokenString string) (*provider.Claims, error)

	// GenerateToken generates a new JWT token for a user
	GenerateToken(userID, username, role string) (string, error)

	// RefreshToken generates new tokens using a refresh token
	RefreshToken(refreshToken string) (string, string, model.UserEntity, error)

	// GetUserFromToken extracts user information from a token
	GetUserFromToken(tokenString string) (model.UserEntity, error)

	// Middleware returns the authentication middleware
	Middleware(next http.Handler) http.Handler
}

// Service is the default implementation of AuthService
type Service struct {
	provider provider.AuthService
}

// NewAuthService creates a new instance of AuthService
func NewAuthService() *Service {
	return &Service{
		provider: provider.INSTANCE,
	}
}

// Authenticate delegates to the underlying provider
func (s *Service) Authenticate(username, password string) (string, string, model.UserEntity, error) {
	return s.provider.Authenticate(username, password)
}

// ValidateToken delegates to the underlying provider
func (s *Service) ValidateToken(tokenString string) (*provider.Claims, error) {
	return s.provider.ValidateToken(tokenString)
}

// GenerateToken delegates to the underlying provider
func (s *Service) GenerateToken(userID, username, role string) (string, error) {
	return s.provider.GenerateToken(userID, username, role)
}

// RefreshToken delegates to the underlying provider
func (s *Service) RefreshToken(refreshToken string) (string, string, model.UserEntity, error) {
	return s.provider.RefreshToken(refreshToken)
}

// GetUserFromToken delegates to the underlying provider
func (s *Service) GetUserFromToken(tokenString string) (model.UserEntity, error) {
	return s.provider.GetUserFromToken(tokenString)
}

// Middleware delegates to the underlying provider's middleware
func (s *Service) Middleware(next http.Handler) http.Handler {
	return s.provider.Middleware(next)
}
