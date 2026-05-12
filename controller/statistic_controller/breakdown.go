package statistic_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/util"
)

// GetDailyBreakdown returns category breakdown for a specific day
func GetDailyBreakdown(w http.ResponseWriter, r *http.Request) {
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

	breakdown, err := statistic_service.GetBreakdownForUser("daily", date, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, breakdown)
}

// GetMonthlyBreakdown returns category breakdown for a specific month
func GetMonthlyBreakdown(w http.ResponseWriter, r *http.Request) {
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

	breakdown, err := statistic_service.GetBreakdownForUser("monthly", month, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, breakdown)
}

// GetYearlyBreakdown returns category breakdown for a specific year
func GetYearlyBreakdown(w http.ResponseWriter, r *http.Request) {
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

	breakdown, err := statistic_service.GetBreakdownForUser("yearly", year, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, breakdown)
}
