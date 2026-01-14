package category_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

// UpdateById updates a category by ID
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
	parentPlainId, _ := requestBody["parent_id"].(string)
	categoryName, _ := requestBody["name"].(string)
	categoryType, _ := requestBody["type"].(string)
	remark, _ := requestBody["remark"].(string)

	// Call user-specific service to update
	updatedCategory, err := category_service.UpdateByIdForUser(plainId, categoryName, categoryType, remark, parentPlainId, userId)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedCategory)
}
