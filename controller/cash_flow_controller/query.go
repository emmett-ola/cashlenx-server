package cash_flow_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

func QueryById(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("id is required"))
		return
	}

	cashFlowEntity, err := queryCashByIDForUser(id, userId)
	if err != nil {
		if err.Error() == "cash_flow not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeErrorResponse(w, r, err)
		}
		return
	}
	util.ComposeJSONResponse(w, http.StatusOK, cashFlowEntity)
}

func QueryByDate(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	date := mux.Vars(r)["date"]
	if date == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("date is required"))
		return
	}

	cashFlowEntityList, err := queryCashByDateForUser(date, userId)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusOK, cashFlowEntityList)
}
