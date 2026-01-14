package category_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

func Tree(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userId := r.Context().Value("user_id")
	if userId == nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"code":    http.StatusUnauthorized,
			"message": "user not authenticated",
		})
		return
	}

	userStrId, ok := userId.(string)
	if !ok {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"code":    http.StatusUnauthorized,
			"message": "invalid user ID format",
		})
		return
	}

	// Get category type from query parameter
	categoryType := r.URL.Query().Get("type")

	// Validate category type if provided
	if categoryType != "" && categoryType != "income" && categoryType != "expense" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "category type must be 'income' or 'expense'",
		})
		return
	}

	// Get category tree with user ID and type filter
	tree, err := category_service.GetCategoryTreeByUser(userStrId, categoryType)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"data": tree,
	})
}
