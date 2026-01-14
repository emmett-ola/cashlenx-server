package provider

var INSTANCE AuthService

func init() {
	// Initialize local auth service by default
	INSTANCE = NewLocalAuthService()
}
