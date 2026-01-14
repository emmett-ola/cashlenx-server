package category_controller

import (
	"encoding/json"
	"net/http"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

func Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "invalid request body",
		})
		return
	}

	// Get user ID from request context
	userIdStr, ok := r.Context().Value("user_id").(string)
	if !ok || userIdStr == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"code":    http.StatusUnauthorized,
			"message": "user not authenticated",
		})
		return
	}

	// Create category using user-specific service
	createdCategory, err := category_service.CreateForUser(req.Name, req.Type, req.Remark, req.ParentId, userIdStr)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": createdCategory,
	})
}
