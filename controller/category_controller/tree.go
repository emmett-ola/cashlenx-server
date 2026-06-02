package category_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

func Tree(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userId := r.Context().Value("user_id")
	if userId == nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	userStrId, ok := userId.(string)
	if !ok {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("invalid user ID format"))
		return
	}

	// Get category type from query parameter
	categoryType := r.URL.Query().Get("type")

	// Validate category type if provided
	if categoryType != "" && categoryType != "income" && categoryType != "expense" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("category type must be 'income' or 'expense'"))
		return
	}

	// Get category tree with user ID and type filter
	tree, err := queryCategoryTreeForUser(userStrId, categoryType)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, tree)
}
