package auth_controller

import (
	"io"
	"net/http"
	"strings"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Login handles user login requests (via username/password or refresh_token)
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.UserLoginRequest
	if err := util.ParseJSONRequest(r, &loginRequest); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	deviceId, deviceName, ipAddress, userAgent := getDeviceInfo(r)
	var accessToken, refreshToken string
	var user model.UserEntity
	var err error

	// Check if this is a refresh token request
	if loginRequest.RefreshToken != "" {
		accessToken, refreshToken, user, err = auth.Service.RefreshToken(loginRequest.RefreshToken,
			deviceId, deviceName, ipAddress, userAgent)
	} else {
		// Normal login request
		// Validate required fields
		if loginRequest.Username == "" || loginRequest.Password == "" {
			util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("username and password (or refresh_token) are required"))
			return
		}

		// Use auth service to authenticate
		accessToken, refreshToken, user, err = auth.Service.Authenticate(loginRequest.Username, loginRequest.Password,
			deviceId, deviceName, ipAddress, userAgent)
	}

	if err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Return user info with tokens (without password hash)
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.Id.Hex(),
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreateTime,
			"updated_at": user.UpdateTime,
		},
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	util.ComposeJSONResponse(w, http.StatusOK, response)
}

// Register handles user registration requests
func Register(w http.ResponseWriter, r *http.Request) {
	var registerRequest model.UserRegistrationRequest
	if err := util.ParseJSONRequest(r, &registerRequest); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	userId, err := user_service.RegisterPublicUser(
		registerRequest.Username,
		registerRequest.Password,
		registerRequest.EmailAddress,
		registerRequest.VerificationToken,
	)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	// Get the created user
	createdUser := user_service.GetUserByObjectId(userId)
	if createdUser.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to create user", nil))
		return
	}

	util.ComposeJSONResponse(w, http.StatusCreated, createdUser)
}

// Logout handles user logout requests
func Logout(w http.ResponseWriter, r *http.Request) {
	// Parse request body to check for refresh_token
	var logoutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := util.ParseJSONRequest(r, &logoutRequest)

	message := "Logout accepted"

	if err == nil && logoutRequest.RefreshToken != "" {
		deviceID, deviceName, ipAddress, userAgent := getDeviceInfo(r)
		refreshToken, tokenErr := refresh_token_service.GetRefreshTokenByToken(logoutRequest.RefreshToken, deviceID, deviceName, ipAddress, userAgent)
		if tokenErr != nil {
			util.Logger.Warnw("Logout refresh token ignored", "error", tokenErr, "request_id", util.RequestIDFromContext(r.Context()))
			util.ComposeJSONResponse(w, http.StatusOK, map[string]string{"message": message})
			return
		}
		if revokeErr := refresh_token_service.RevokeRefreshToken(logoutRequest.RefreshToken, refreshToken.UserId); revokeErr != nil {
			util.ComposeErrorResponse(w, r, revokeErr)
			return
		}
		util.ComposeJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully logged out from this device"})
		return
	}

	if err != nil && err != io.EOF {
		util.Logger.Warnw("Logout request body ignored", "error", err, "request_id", util.RequestIDFromContext(r.Context()))
	}

	userID := userIDFromAuthorizationHeader(r)
	if userID == "" {
		util.ComposeJSONResponse(w, http.StatusOK, map[string]string{"message": message})
		return
	}

	if revokeErr := refresh_token_service.RevokeAllRefreshTokens(userID); revokeErr != nil {
		util.ComposeErrorResponse(w, r, revokeErr)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Successfully logged out from all devices",
	})
}

func userIDFromAuthorizationHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		util.Logger.Warnw("Logout authorization header ignored", "request_id", util.RequestIDFromContext(r.Context()))
		return ""
	}

	claims, err := auth.Service.ValidateToken(parts[1])
	if err != nil || claims == nil {
		util.Logger.Warnw("Logout access token ignored", "error", err, "request_id", util.RequestIDFromContext(r.Context()))
		return ""
	}

	return claims.UserID
}

// GetTokens handles requests to list all user's refresh tokens
func GetTokens(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Get all refresh tokens for the user
	tokens := refresh_token_service.GetUserRefreshTokens(userID)

	util.ComposeJSONResponse(w, http.StatusOK, tokens)
}

func getDeviceInfo(r *http.Request) (string, string, string, string) {
	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "Unknown"
	}

	// Simple parsing logic
	deviceName := "Unknown Device"
	if len(userAgent) > 0 {
		if len(userAgent) > 50 {
			deviceName = userAgent[:50] + "..."
		} else {
			deviceName = userAgent
		}
	}

	ipAddress := util.GetClientIP(r)
	deviceId := "" // Let service handle or generate if needed

	return deviceId, deviceName, ipAddress, userAgent
}
