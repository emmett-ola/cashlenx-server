package cash_flow_controller

import (
	"net/http"
	"strconv"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ListAll returns paginated list of all cash flows with filtering
func ListAll(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context (set by Auth middleware)
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		response := model.NewErrorResponse("UNAUTHORIZED", "user not authenticated")
		util.ComposeJSONResponse(w, http.StatusUnauthorized, response)
		return
	}

	// Parse query parameters for pagination and filtering
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	pageStr := r.URL.Query().Get("page")

	// Filter parameters
	cashType := r.URL.Query().Get("type") // INCOME or EXPENSE
	categoryId := r.URL.Query().Get("category_id")
	description := r.URL.Query().Get("description") // Fuzzy search
	exactDescription := r.URL.Query().Get("exact_description")
	fromDate := r.URL.Query().Get("from_date") // YYYYMMDD or YYYY-MM-DD
	toDate := r.URL.Query().Get("to_date")     // YYYYMMDD or YYYY-MM-DD

	// Pagination defaults
	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if pageStr != "" {
		// Calculate offset from page number
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	} else if offsetStr != "" {
		// Use offset directly if provided
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Call user-specific service to get paginated and filtered results
	cashFlows, totalCount, err := cash_flow_service.QueryAllForUser(
		userId,
		cashType,
		categoryId,
		description,
		exactDescription,
		fromDate,
		toDate,
		limit,
		offset,
	)

	if err != nil {
		response := model.NewErrorResponse("INTERNAL_ERROR", err.Error())
		util.ComposeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	meta := map[string]interface{}{
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	// Return with pagination metadata
	response := map[string]interface{}{
		"data": cashFlows,
		"meta": meta,
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
