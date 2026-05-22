package user_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

func GetConfiguration(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	config, err := user_service.GetConfigurationService(userID)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, config)
}

func UpsertConfiguration(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	var requestBody model.UserConfigurationRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	config, err := user_service.UpsertConfigurationService(userID, requestBody)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, config)
}
