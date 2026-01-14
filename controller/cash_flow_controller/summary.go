package cash_flow_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/macar-x/cashlenx-server/util"
)

// GetDailySummary returns summary for a specific day
func GetDailySummary(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	date := vars["date"]

	if date == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("date is required"))
		return
	}

	summary, err := cash_flow_service.GetSummaryForUser("daily", date, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, summary)
}

// GetMonthlySummary returns summary for a specific month (YYYYMM format)
func GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	month := vars["month"]

	if month == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("month is required (YYYYMM format)"))
		return
	}

	summary, err := cash_flow_service.GetSummaryForUser("monthly", month, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, summary)
}

// GetYearlySummary returns summary for a specific year (YYYY format)
func GetYearlySummary(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	year := vars["year"]

	if year == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("year is required (YYYY format)"))
		return
	}

	summary, err := cash_flow_service.GetSummaryForUser("yearly", year, userId)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, summary)
}
