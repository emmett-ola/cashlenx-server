package statistic_controller

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/util"
)

// GetDashboardOverview returns comprehensive dashboard data for visualization
func GetDashboardOverview(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	period := vars["period"] // daily, monthly, yearly
	date := vars["date"]

	if period == "" || date == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("period and date are required"))
		return
	}

	dashboard, err := statistic_service.GetDashboardOverviewForUser(period, date, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, dashboard)
}

// GetIncomeExpenseChartData returns time-series data for income vs expense chart
func GetIncomeExpenseChartData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	period := vars["period"]
	date := vars["date"]

	if period == "" || date == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("period and date are required"))
		return
	}

	chartData, err := statistic_service.GetIncomeExpenseChartDataForUser(period, date, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, chartData)
}

// GetCategoryDistributionData returns pie/donut chart data for category distribution
func GetCategoryDistributionData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	period := vars["period"]
	date := vars["date"]
	flowType := r.URL.Query().Get("type") // income or expense

	if period == "" || date == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("period and date are required"))
		return
	}

	// Default to expense if not specified
	if flowType == "" {
		flowType = "expense"
	}

	distributionData, err := statistic_service.GetCategoryDistributionForUser(period, date, flowType, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, distributionData)
}

// GetMonthlyComparisonData returns bar chart data for month-over-month comparison
func GetMonthlyComparisonData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	year := vars["year"]

	if year == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("year is required"))
		return
	}

	comparisonData, err := statistic_service.GetMonthlyComparisonForUser(year, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, comparisonData)
}

// GetSpendingHeatmapData returns calendar heatmap data for spending visualization
func GetSpendingHeatmapData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	vars := mux.Vars(r)
	year := vars["year"]

	if year == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("year is required"))
		return
	}

	heatmapData, err := statistic_service.GetSpendingHeatmapForUser(year, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, heatmapData)
}
