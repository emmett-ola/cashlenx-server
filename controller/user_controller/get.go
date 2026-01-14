package user_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Get retrieves a user by ID
func Get(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path parameters
	vars := mux.Vars(r)
	userId := vars["id"]

	if userId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("user ID is required"))
		return
	}

	// Get user from service
	user := user_service.GetUserByObjectId(userId)
	if user.IsEmpty() {
		util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError("user not found"))
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, user)
}
