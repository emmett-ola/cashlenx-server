package auth

import (
	"github.com/macar-x/cashlenx-server/auth/service"
)

// Service is the global authentication service instance
var Service *service.Service

func init() {
	// Initialize the authentication service
	Service = service.NewAuthService()
}
