package category_controller

import (
	"encoding/json"
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Get user ID from request context
	userIdStr, ok := r.Context().Value("user_id").(string)
	if !ok || userIdStr == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Create category using user-specific service
	createdCategory, err := createCategoryForUser(req.Name, req.Type, req.Remark, req.ParentId, userIdStr)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusCreated, createdCategory)
}
