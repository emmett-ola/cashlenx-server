package user_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Update updates an existing user
func Update(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path parameters
	vars := mux.Vars(r)
	userId := vars["id"]

	if userId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("user ID is required"))
		return
	}

	var requestBody model.UserDTO
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Update user via service
	updatedUser, err := user_service.UpdateService(userId, requestBody)
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			util.ComposeJSONResponse(w, http.StatusConflict, err)
			return
		}
		if errors.IsNotFound(err) {
			util.ComposeJSONResponse(w, http.StatusNotFound, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedUser)
}

// SetPassword allows users to set or reset their password
func SetPassword(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path parameters
	vars := mux.Vars(r)
	userId := vars["id"]

	if userId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("user ID is required"))
		return
	}

	// Parse request body to get password
	type PasswordRequest struct {
		Password string `json:"password"`
	}

	var requestBody PasswordRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if requestBody.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("password is required"))
		return
	}

	// Set password via service
	updatedUser, err := user_service.SetPasswordService(userId, requestBody.Password)
	if err != nil {
		if errors.IsNotFound(err) {
			util.ComposeJSONResponse(w, http.StatusNotFound, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedUser)
}

// UpdateProfile updates user profile information (nickname, avatar_url, email_address, gender)
// This endpoint allows users to update their own profile
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT context (middleware sets this)
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	var requestBody model.UserDTO
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Update profile via service
	updatedUser, err := user_service.UpdateProfileService(userId, requestBody)
	if err != nil {
		if errors.IsNotFound(err) {
			util.ComposeJSONResponse(w, http.StatusNotFound, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedUser)
}
