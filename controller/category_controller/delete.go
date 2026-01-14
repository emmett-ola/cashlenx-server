package category_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

// DeleteById deletes a category by ID
func DeleteById(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("id is required"))
		return
	}

	// Delete the category using user-specific service
	deletedCategory, err := category_service.DeleteByIdForUser(id, userId)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Return the deleted category
	util.ComposeJSONResponse(w, http.StatusOK, deletedCategory)
}
