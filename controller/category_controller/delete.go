package category_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
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

	// Check for force flag in query parameters
	force := r.URL.Query().Get("force") == "true"

	// Delete the category using user-specific service
	deletedCategory, err := deleteCategoryForUser(id, userId, force)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeErrorResponse(w, r, err)
		}
		return
	}

	// Return the deleted category
	util.ComposeJSONResponse(w, http.StatusOK, deletedCategory)
}
