package cash_flow_controller

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/util"
)

// UpdateById updates a cash flow record by ID
func UpdateById(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	plainId := vars["id"]
	if plainId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("id is required"))
		return
	}

	// Parse JSON body for update fields
	var requestBody map[string]interface{}
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Extract optional fields
	belongsDate, _ := requestBody["belongs_date"].(string)
	categoryName, _ := requestBody["category_name"].(string)
	description, _ := requestBody["description"].(string)

	var amount float64
	if amountVal, ok := requestBody["amount"]; ok {
		switch v := amountVal.(type) {
		case float64:
			amount = v
		case string:
			amount, _ = strconv.ParseFloat(v, 64)
		}
	}

	// Call user-specific service to update
	updatedEntity, err := cash_flow_service.UpdateByIdForUser(plainId, belongsDate, categoryName, amount, description, userId)
	if err != nil {
		if err.Error() == "cash_flow not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedEntity)
}
