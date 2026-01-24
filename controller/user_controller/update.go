package user_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Update updates an existing user (admin only)
func Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("user id is required"))
		return
	}

	var requestBody model.UserDTO
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	updatedUser, err := user_service.UpdateService(id, requestBody)
	if err != nil {
		if errors.IsNotFound(err) {
			util.ComposeJSONResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.IsAlreadyExistsError(err) {
			util.ComposeJSONResponse(w, http.StatusConflict, err)
			return
		}
		if errors.IsValidationError(err) {
			util.ComposeJSONResponse(w, http.StatusBadRequest, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedUser)
}
