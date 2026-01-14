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
	if err := user_service.UpdateService(userId, requestBody); err != nil {
		if errors.IsAlreadyExistsError(err) {
			util.ComposeJSONResponse(w, http.StatusConflict, err)
			return
		}
		if err.Error() == "user not found" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError(err.Error(), nil))
		return
	}

	// Get the updated user
	updatedUser := user_service.GetUserByObjectId(userId)
	if updatedUser.IsEmpty() {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to retrieve updated user", nil))
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedUser)
}
