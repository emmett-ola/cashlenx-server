package user_controller

import (
	"net/http"
	"strconv"

	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ListAll returns all users with pagination
func ListAll(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
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

	// Get users from service
	users := user_service.GetAllUsers(limit, offset)
	count := user_service.CountAllUsers()

	response := map[string]interface{}{
		"data": users,
		"meta": map[string]interface{}{
			"total_count": count,
			"limit":       limit,
			"offset":      offset,
		},
	}

	util.ComposeJSONResponse(w, http.StatusOK, response)
}
