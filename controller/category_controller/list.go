package category_controller

import (
	"net/http"
	"strconv"

	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ListAll returns paginated list of all categories
func ListAll(w http.ResponseWriter, r *http.Request) {
	// Get current user ID from context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"message": "Invalid or missing user authentication",
		})
		return
	}

	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit for categories
	offset := 0 // Default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Get category type filter
	categoryType := r.URL.Query().Get("type")

	// Call user-specific service to get paginated results
	categories, totalCount, err := category_service.QueryAllForUser(userId, categoryType, limit, offset)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error":   err.Error(),
			"message": "Failed to retrieve categories",
		})
		return
	}

	// Return with pagination metadata
	util.ComposeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": categories,
		"meta": map[string]interface{}{
			"total_count": totalCount,
			"limit":       limit,
			"offset":      offset,
		},
	})
}
