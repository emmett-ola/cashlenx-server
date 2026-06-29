package user_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

// RequestPasswordReset handles password reset requests
func RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var requestBody model.PasswordResetRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if requestBody.EmailOrUsername == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("email or username is required"))
		return
	}

	ipAddress := util.GetClientIP(r)

	// Request password reset
	err := requestUserPasswordReset(requestBody.EmailOrUsername, ipAddress)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	// For security reasons, always return success regardless of whether user exists
	response := map[string]interface{}{
		"message": "If the user exists, a password reset verification code will be sent to their email",
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}

// ConfirmPasswordReset handles password reset confirmation
func ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var requestBody model.PasswordResetConfirmRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if requestBody.Token == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("password reset token is required"))
		return
	}

	if requestBody.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("new password is required"))
		return
	}

	// Confirm password reset
	err := confirmUserPasswordReset(requestBody.Token, requestBody.Password)
	if err != nil {
		if errors.IsValidationError(err) {
			util.ComposeJSONResponse(w, http.StatusBadRequest, err)
			return
		}
		util.ComposeErrorResponse(w, r, err)
		return
	}

	response := map[string]interface{}{
		"message": "Password has been successfully reset",
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
