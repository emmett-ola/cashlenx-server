package category_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

// QueryById retrieves a category by ID
func QueryById(w http.ResponseWriter, r *http.Request) {
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

	category, err := category_service.QueryByIdForUser(id, userId)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, category)
}

// QueryByName retrieves a category by name (exact match)
func QueryByName(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("name is required"))
		return
	}

	category, err := category_service.QueryByNameForUser(name, userId)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, category)
}

// QueryChildren retrieves children categories by parent ID
func QueryChildren(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	parentId := vars["parent_id"]

	if parentId == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("parent_id is required"))
		return
	}

	// Get category type filter from query parameter
	categoryType := r.URL.Query().Get("type")

	categories, err := category_service.GetChildCategoriesForUser(parentId, userId, categoryType)
	if err != nil {
		if err.Error() == "parent category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, categories)
}
