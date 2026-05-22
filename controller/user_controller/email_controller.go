package user_controller

import (
	"encoding/json"
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// RequestEmailChange handles the request to change email
func RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	// Parse request body
	var req model.UserChangeEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewBadRequestError("invalid request body"))
		return
	}

	// Validate request
	if req.NewEmail == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewBadRequestError("new email is required"))
		return
	}

	// Call service
	err := user_service.RequestEmailChange(userID, req.NewEmail, req.VerificationToken)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	// Return success
	response := map[string]string{
		"message": "Email address updated successfully",
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}

// ConfirmEmailChange handles the confirmation of email change
func ConfirmEmailChange(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	// Parse request body
	var req model.UserConfirmEmailChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewBadRequestError("invalid request body"))
		return
	}

	// Validate request
	if req.Token == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewBadRequestError("token is required"))
		return
	}
	if req.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewBadRequestError("password is required"))
		return
	}

	// Call service
	err := user_service.ConfirmEmailChange(userID, req.Token, req.Password)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Email address updated successfully",
	})
}
