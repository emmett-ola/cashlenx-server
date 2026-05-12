package auth_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
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
	// Check if registration is enabled
	registerEnabled := util.GetConfigByKey("auth.registration.enabled")
	if registerEnabled != "true" {
		util.ComposeJSONResponse(w, http.StatusForbidden, errors.NewForbiddenError("user registration is disabled"))
		return
	}

	var registerRequest model.UserRegistrationRequest
	if err := util.ParseJSONRequest(r, &registerRequest); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Validate required fields
	if registerRequest.Username == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("username is required"))
		return
	}
	if registerRequest.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("password is required"))
		return
	}

	// Validate password requirements
	if err := validation.ValidatePassword(registerRequest.Password); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, err)
		return
	}

	// Check if user already exists (including deleted users)
	existingUser := user_mapper.INSTANCE.GetUserByUsernameIncludeDeleted(registerRequest.Username)
	if !existingUser.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusConflict, errors.NewFieldAlreadyExistsError("username", "username already exists"))
		return
	}

	// Create user DTO
	userDTO := model.UserDTO{
		Username: registerRequest.Username,
		Password: registerRequest.Password,
	}

	// Create user
	// Pass nil as creatorId to indicate self-registration
	userId, err := user_service.CreateService(userDTO, nil)
	if err != nil {
		util.ComposeErrorResponse(w, err)
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
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse request body to check for refresh_token
	var logoutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := util.ParseJSONRequest(r, &logoutRequest)

	var logoutErr error
	var message string

	if err == nil && logoutRequest.RefreshToken != "" {
		// Refresh token provided - logout only from this session (local logout)
		logoutErr = refresh_token_service.RevokeRefreshToken(logoutRequest.RefreshToken, userID)
		message = "Successfully logged out from this device"
	} else {
		// No refresh token provided - logout from all sessions (logout everywhere)
		logoutErr = refresh_token_service.RevokeAllRefreshTokens(userID)
		message = "Successfully logged out from all devices"
	}

	if logoutErr != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, logoutErr)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": message,
	})
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
