package auth_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/middleware"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"golang.org/x/crypto/bcrypt"
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

	// Get user by username
	user := user_service.GetUserByUsername(loginRequest.Username)
	if user.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("invalid username or password"))
		return
	}

	// Verify password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Password))
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("invalid username or password"))
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.Id.Hex(), user.Username, user.Role)
	if err != nil {
		util.Logger.Errorw("Failed to generate JWT token", "error", err)
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to generate authentication token", nil))
		return
	}

	// Return user info with token (without password hash)
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.Id.Hex(),
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
		"token": token,
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
