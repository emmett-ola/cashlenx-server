package auth_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/auth"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
)

// Login handles user login requests
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.UserLoginRequest
	if err := util.ParseJSONRequest(r, &loginRequest); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Validate required fields
	if loginRequest.Username == "" || loginRequest.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("username and password are required"))
		return
	}

	// Use auth service to authenticate
	accessToken, refreshToken, user, err := auth.Service.Authenticate(loginRequest.Username, loginRequest.Password)
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

// RefreshToken handles token refresh requests
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshRequest model.RefreshTokenRequest
	if err := util.ParseJSONRequest(r, &refreshRequest); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Validate required field
	if refreshRequest.RefreshToken == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("refresh_token is required"))
		return
	}

	// Use auth service to refresh token
	accessToken, newRefreshToken, user, err := auth.Service.RefreshToken(refreshRequest.RefreshToken)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Return user info with new tokens
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.Id.Hex(),
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreateTime,
			"updated_at": user.UpdateTime,
		},
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
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

	// Check if user already exists
	existingUser := user_service.GetUserByUsername(registerRequest.Username)
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
	userId, err := user_service.CreateService(userDTO)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
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
