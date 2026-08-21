package budget_controller

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	appErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

func userID(r *http.Request) (string, error) {
	value, ok := r.Context().Value("user_id").(string)
	if !ok || value == "" {
		return "", appErrors.NewUnauthorizedError("user not authenticated")
	}
	return value, nil
}

func decodeRequest(r *http.Request) (model.UpsertBudgetRequest, error) {
	var request model.UpsertBudgetRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return request, appErrors.NewInvalidInputError("invalid request body")
	}
	return request, nil
}

func Create(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	request, err := decodeRequest(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	view, err := createBudget(request, uid)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusCreated, view)
}

func List(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	views, err := listBudgets(r.URL.Query().Get("period"), uid)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": views})
}

func Get(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	view, err := getBudget(mux.Vars(r)["id"], uid)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusOK, view)
}

func Update(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	request, err := decodeRequest(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	view, err := updateBudget(mux.Vars(r)["id"], request, uid)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	util.ComposeJSONResponse(w, http.StatusOK, view)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	uid, err := userID(r)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	if err := deleteBudget(mux.Vars(r)["id"], uid); err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
