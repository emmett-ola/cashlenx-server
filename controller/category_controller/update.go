package category_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
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
	emoji, err := optionalString(requestBody, "emoji")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError(err.Error()))
		return
	}
	bgColor, err := optionalString(requestBody, "bg_color")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError(err.Error()))
		return
	}

	// Call user-specific service to update
	updatedCategory, err := updateCategoryForUser(plainId, categoryName, categoryType, remark, parentPlainId, userId, emoji, bgColor)
	if err != nil {
		if err.Error() == "category not found or access denied" {
			util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError(err.Error()))
		} else {
			util.ComposeErrorResponse(w, r, err)
		}
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, updatedCategory)
}

func optionalString(values map[string]interface{}, key string) (*string, error) {
	value, exists := values[key]
	if !exists {
		return nil, nil
	}
	text, ok := value.(string)
	if !ok {
		return nil, errors.NewInvalidInputError(key + " must be a string")
	}
	return &text, nil
}
