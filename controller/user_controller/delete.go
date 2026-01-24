package user_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Delete deletes a user by ID
func Delete(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path parameters
	vars := mux.Vars(r)
	userId := vars["id"]

	if userId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("user ID is required"))
		return
	}

	// Delete user via service
	if err := user_service.DeleteService(userId); err != nil {
		if errors.IsNotFound(err) {
			util.ComposeJSONResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.IsForbiddenError(err) {
			util.ComposeJSONResponse(w, http.StatusForbidden, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Return success response
	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "User deleted successfully",
	})
}
