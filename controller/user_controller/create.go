package user_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// Create creates a new user
func Create(w http.ResponseWriter, r *http.Request) {
	var requestBody model.UserDTO
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if requestBody.Username == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("username is required"))
		return
	}

	if requestBody.Password == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("password is required"))
		return
	}

	plainId, err := user_service.CreateService(requestBody)
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			util.ComposeJSONResponse(w, http.StatusConflict, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Get the created user entity
	createdUser := user_service.GetUserByObjectId(plainId)
	if createdUser.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to retrieve created user", nil))
		return
	}

	util.ComposeJSONResponse(w, http.StatusCreated, createdUser)
}
