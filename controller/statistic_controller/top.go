package statistic_controller

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

// GetDailyTopExpenses returns top N expenses for a specific day
func GetDailyTopExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Parse limit from query parameter (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	topExpenses, err := getTopExpensesForUser(limit, "daily", date, userId)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, topExpenses)
}

// GetMonthlyTopExpenses returns top N expenses for a specific month
func GetMonthlyTopExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Parse limit from query parameter (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	topExpenses, err := getTopExpensesForUser(limit, "monthly", month, userId)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, topExpenses)
}

// GetYearlyTopExpenses returns top N expenses for a specific year
func GetYearlyTopExpenses(w http.ResponseWriter, r *http.Request) {
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

	// Parse limit from query parameter (default 10)
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	topExpenses, err := getTopExpensesForUser(limit, "yearly", year, userId)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, topExpenses)
}
