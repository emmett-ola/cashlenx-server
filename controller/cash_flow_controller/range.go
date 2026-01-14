package cash_flow_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/util"
)

// QueryByDateRange queries cash flows between two dates
func QueryByDateRange(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse query parameters
	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")

	if fromDate == "" || toDate == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("from and to dates are required"))
		return
	}

	// Call user-specific service to get records in range
	cashFlowEntities, err := cash_flow_service.QueryByDateRangeForUser(fromDate, toDate, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, cashFlowEntities)
}
