package controller

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/controller/auth_controller"
	"github.com/macar-x/cashlenx-server/controller/cash_flow_controller"
	"github.com/macar-x/cashlenx-server/controller/category_controller"
	"github.com/macar-x/cashlenx-server/controller/manage_controller"
	"github.com/macar-x/cashlenx-server/controller/statistic_controller"
	"github.com/macar-x/cashlenx-server/controller/user_controller"
	"github.com/macar-x/cashlenx-server/middleware"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

func StartServer(port int32) {
	// Explicitly load timezone at server startup to ensure it's configured
	// and logged immediately
	tz := util.GetTimezone()
	fmt.Printf("Loaded timezone: %v\n", tz)

	// Initialize Snowflake ID generator
	workerID := util.GetConfigInt("snowflake.worker_id", 0)
	if err := util.InitSnowflakeGenerator(workerID); err != nil {
		util.Logger.Warnf("Failed to initialize Snowflake generator with worker ID %d: %v, using default", workerID, err)
	} else {
		fmt.Printf("Snowflake ID generator initialized with worker ID: %d\n", workerID)
	}

	// Initialize admin user if needed
	user_service.InitAdminUser()

	r := mux.NewRouter()

	// Register routes with new structure
	registerOpenRoutes(r) // Public endpoints (no auth)

	// Create admin subrouter with Admin middleware
	adminRouter := r.PathPrefix("/api/admin").Subrouter()
	adminRouter.Use(middleware.Admin)
	registerAdminRoutes(adminRouter) // Admin-only endpoints

	registerUserRoutes(r)     // User-specific profile endpoints
	registerCashRoute(r)      // User-specific cash flow endpoints
	registerCategoryRoute(r)  // User-specific category endpoints
	registerStatisticRoute(r) // User-specific statistic endpoints

	// Apply middleware
	handler := middleware.Logging(middleware.Auth(middleware.SchemaValidation(middleware.CORS(r))))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("API server is running on http://localhost%s\n", addr)
	http.ListenAndServe(addr, handler)
}

// registerOpenRoutes registers public endpoints that don't require authentication
func registerOpenRoutes(r *mux.Router) {
	// System health and version
	r.HandleFunc("/api/open/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/open/version", versionInfo).Methods("GET")

	// Authentication routes
	r.HandleFunc("/api/open/auth/login", auth_controller.Login).Methods("POST")
	r.HandleFunc("/api/open/auth/register", auth_controller.Register).Methods("POST")
	r.HandleFunc("/api/open/auth/refresh", auth_controller.RefreshToken).Methods("POST")

	// Protected auth routes
	r.HandleFunc("/api/auth/logout", auth_controller.Logout).Methods("POST")

	// Password reset routes
	r.HandleFunc("/api/open/auth/reset-password", user_controller.RequestPasswordReset).Methods("POST")
	r.HandleFunc("/api/open/auth/reset-password/confirm", user_controller.ConfirmPasswordReset).Methods("POST")
}

// registerAdminRoutes registers admin-only endpoints
func registerAdminRoutes(r *mux.Router) {
	// User management - admin only
	r.HandleFunc("/user", user_controller.Create).Methods("POST")
	r.HandleFunc("/user", user_controller.ListAll).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Get).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Update).Methods("PUT")
	r.HandleFunc("/user/{id}", user_controller.Delete).Methods("DELETE")
	// User password management
	r.HandleFunc("/user/{id}/password", user_controller.SetPassword).Methods("PUT")

	// Database management - admin only
	r.HandleFunc("/manage/dump", manage_controller.DumpDatabase).Methods("GET")
	r.HandleFunc("/manage/restore", manage_controller.RestoreDatabase).Methods("POST")
	r.HandleFunc("/manage/export", manage_controller.ExportData).Methods("GET")
	r.HandleFunc("/manage/import", manage_controller.ImportData).Methods("POST")
}

// registerUserRoutes registers user-specific endpoints (authenticated users can access their own profiles)
func registerUserRoutes(r *mux.Router) {
	// User profile management
	r.HandleFunc("/api/user/profile", user_controller.UpdateProfile).Methods("PUT")
}

func registerCashRoute(r *mux.Router) {
	// Create
	r.HandleFunc("/api/cash/expense", cash_flow_controller.CreateExpense).Methods("POST")
	r.HandleFunc("/api/cash/income", cash_flow_controller.CreateIncome).Methods("POST")

	// Read
	r.HandleFunc("/api/cash", cash_flow_controller.ListAll).Methods("GET")
	r.HandleFunc("/api/cash/range", cash_flow_controller.QueryByDateRange).Methods("GET")
	r.HandleFunc("/api/cash/date/{date}", cash_flow_controller.QueryByDate).Methods("GET")
	r.HandleFunc("/api/cash/date/{date}", cash_flow_controller.DeleteByDate).Methods("DELETE")
	r.HandleFunc("/api/cash/{id}", cash_flow_controller.QueryById).Methods("GET")

	// Summary endpoints
	r.HandleFunc("/api/cash/summary/daily/{date}", cash_flow_controller.GetDailySummary).Methods("GET")
	r.HandleFunc("/api/cash/summary/monthly/{month}", cash_flow_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc("/api/cash/summary/yearly/{year}", cash_flow_controller.GetYearlySummary).Methods("GET")

	// Update
	r.HandleFunc("/api/cash/{id}", cash_flow_controller.UpdateById).Methods("PUT")

	// Delete
	r.HandleFunc("/api/cash/{id}", cash_flow_controller.DeleteById).Methods("DELETE")
}

func registerCategoryRoute(r *mux.Router) {
	// Create
	r.HandleFunc("/api/category", category_controller.Create).Methods("POST")
	// Read all categories with filtering
	r.HandleFunc("/api/category", category_controller.ListAll).Methods("GET")
	// Read children categories - RESTful design: parent/{id}/children
	r.HandleFunc("/api/category/name/{name}", category_controller.QueryByName).Methods("GET")
	r.HandleFunc("/api/category/{parent_id}/children", category_controller.QueryChildren).Methods("GET")
	r.HandleFunc("/api/category/tree", category_controller.Tree).Methods("GET")
	// Read specific category - must be after specific paths
	r.HandleFunc("/api/category/{id}", category_controller.QueryById).Methods("GET")

	// Update
	r.HandleFunc("/api/category/{id}", category_controller.UpdateById).Methods("PUT")

	// Delete
	r.HandleFunc("/api/category/{id}", category_controller.DeleteById).Methods("DELETE")
}

func registerStatisticRoute(r *mux.Router) {
	// Export/Import user-specific data
	r.HandleFunc("/api/statistic/export", statistic_controller.ExportData).Methods("GET")
	r.HandleFunc("/api/statistic/import", statistic_controller.ImportData).Methods("POST")

	// Summary endpoints
	r.HandleFunc("/api/statistic/summary/daily/{date}", statistic_controller.GetDailySummary).Methods("GET")
	r.HandleFunc("/api/statistic/summary/monthly/{month}", statistic_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc("/api/statistic/summary/yearly/{year}", statistic_controller.GetYearlySummary).Methods("GET")

	// Breakdown endpoints
	r.HandleFunc("/api/statistic/breakdown/daily/{date}", statistic_controller.GetDailyBreakdown).Methods("GET")
	r.HandleFunc("/api/statistic/breakdown/monthly/{month}", statistic_controller.GetMonthlyBreakdown).Methods("GET")
	r.HandleFunc("/api/statistic/breakdown/yearly/{year}", statistic_controller.GetYearlyBreakdown).Methods("GET")

	// Trends endpoints
	r.HandleFunc("/api/statistic/trends/daily/{date}", statistic_controller.GetDailyTrends).Methods("GET")
	r.HandleFunc("/api/statistic/trends/monthly/{month}", statistic_controller.GetMonthlyTrends).Methods("GET")
	r.HandleFunc("/api/statistic/trends/yearly/{year}", statistic_controller.GetYearlyTrends).Methods("GET")

	// Top expenses endpoints
	r.HandleFunc("/api/statistic/top/daily/{date}", statistic_controller.GetDailyTopExpenses).Methods("GET")
	r.HandleFunc("/api/statistic/top/monthly/{month}", statistic_controller.GetMonthlyTopExpenses).Methods("GET")
	r.HandleFunc("/api/statistic/top/yearly/{year}", statistic_controller.GetYearlyTopExpenses).Methods("GET")

	// Dashboard visualization endpoints
	r.HandleFunc("/api/statistic/dashboard/{period}/{date}", statistic_controller.GetDashboardOverview).Methods("GET")
	r.HandleFunc("/api/statistic/chart/income-expense/{period}/{date}", statistic_controller.GetIncomeExpenseChartData).Methods("GET")
	r.HandleFunc("/api/statistic/chart/category-distribution/{period}/{date}", statistic_controller.GetCategoryDistributionData).Methods("GET")
	r.HandleFunc("/api/statistic/chart/monthly-comparison/{year}", statistic_controller.GetMonthlyComparisonData).Methods("GET")
	r.HandleFunc("/api/statistic/chart/spending-heatmap/{year}", statistic_controller.GetSpendingHeatmapData).Methods("GET")
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "cashlenx-api",
		"message": "API is running",
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}

// Version info endpoint
func versionInfo(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"version":     model.Version,
		"name":        "CashLenX API",
		"description": "Personal finance management API",
		"endpoints": map[string][]string{
			"open": {
				"GET /api/open/health",
				"GET /api/open/version",
				"POST /api/open/auth/login",
				"POST /api/open/auth/register",
				"POST /api/open/auth/refresh",
				"POST /api/open/auth/reset-password",
				"POST /api/open/auth/reset-password/confirm",
			},
			"admin": {
				"POST /api/admin/user",
				"GET /api/admin/user",
				"GET /api/admin/user/{id}",
				"PUT /api/admin/user/{id}",
				"DELETE /api/admin/user/{id}",
				"PUT /api/admin/user/{id}/password",
				"GET /api/admin/manage/dump",
				"POST /api/admin/manage/restore",
				"GET /api/admin/manage/export",
				"POST /api/admin/manage/import",
			},
			"user": {
				"PUT /api/user/profile",
			},
			"cash_flow": {
				"POST /api/cash/expense",
				"POST /api/cash/income",
				"GET /api/cash",
				"GET /api/cash?limit=20&offset=0&type=income",
				"GET /api/cash/{id}",
				"GET /api/cash/date/{date}",
				"GET /api/cash/range?from=YYYYMMDD&to=YYYYMMDD",
				"GET /api/cash/summary/daily/{date}",
				"GET /api/cash/summary/monthly/{month}",
				"GET /api/cash/summary/yearly/{year}",
				"PUT /api/cash/{id}",
				"DELETE /api/cash/{id}",
				"DELETE /api/cash/date/{date}",
			},
			"category": {
				"POST /api/category",
				"GET /api/category",
				"GET /api/category?type=income&parent_id=XXX",
				"GET /api/category/{id}",
				"GET /api/category/name/{name}",
				"GET /api/category/{parent_id}/children",
				"GET /api/category/tree",
				"PUT /api/category/{id}",
				"DELETE /api/category/{id}",
			},
			"statistic": {
				"GET /api/statistic/export?from_date=YYYYMMDD&to_date=YYYYMMDD&format=xlsx|csv|pdf (binary download)",
				"POST /api/statistic/import?file_path=path",
				"GET /api/statistic/summary/daily/{date}",
				"GET /api/statistic/summary/monthly/{month}",
				"GET /api/statistic/summary/yearly/{year}",
				"GET /api/statistic/breakdown/daily/{date}",
				"GET /api/statistic/breakdown/monthly/{month}",
				"GET /api/statistic/breakdown/yearly/{year}",
				"GET /api/statistic/trends/daily/{date}",
				"GET /api/statistic/trends/monthly/{month}",
				"GET /api/statistic/trends/yearly/{year}",
				"GET /api/statistic/top/daily/{date}?limit=10",
				"GET /api/statistic/top/monthly/{month}?limit=10",
				"GET /api/statistic/top/yearly/{year}?limit=10",
				"GET /api/statistic/dashboard/{period}/{date}",
				"GET /api/statistic/chart/income-expense/{period}/{date}",
				"GET /api/statistic/chart/category-distribution/{period}/{date}?type=income|expense",
				"GET /api/statistic/chart/monthly-comparison/{year}",
				"GET /api/statistic/chart/spending-heatmap/{year}",
			},
		},
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
