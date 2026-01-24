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
	apiVersion := util.GetConfigByKey("api.version")
	if apiVersion == "" {
		apiVersion = "v0"
	}
	apiPrefix := "/api/" + apiVersion

	registerOpenRoutes(r, apiPrefix) // Public endpoints (no auth)

	// Create admin subrouter with Admin middleware
	adminRouter := r.PathPrefix(apiPrefix + "/admin").Subrouter()
	adminRouter.Use(middleware.Admin)
	registerAdminRoutes(adminRouter) // Admin-only endpoints

	registerUserRoutes(r, apiPrefix)     // User-specific profile endpoints
	registerCashRoute(r, apiPrefix)      // User-specific cash flow endpoints
	registerCategoryRoute(r, apiPrefix)  // User-specific category endpoints
	registerStatisticRoute(r, apiPrefix) // User-specific statistic endpoints

	// Apply middleware
	handler := middleware.Logging(middleware.Auth(middleware.SchemaValidation(middleware.CORS(r))))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("API server is running on http://localhost%s\n", addr)
	http.ListenAndServe(addr, handler)
}

// registerOpenRoutes registers public endpoints that don't require authentication
func registerOpenRoutes(r *mux.Router, prefix string) {
	// System health and version
	r.HandleFunc(prefix+"/open/health", healthCheck).Methods("GET")
	r.HandleFunc(prefix+"/open/version", versionInfo).Methods("GET")

	// Authentication routes
	r.HandleFunc(prefix+"/open/auth/login", auth_controller.Login).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/register", auth_controller.Register).Methods("POST")

	// Protected auth routes (technically these shouldn't be in 'open' but kept for grouping logic)
	// Note: They are protected by middleware checking for /api/open/ prefix exception
	r.HandleFunc(prefix+"/auth/logout", auth_controller.Logout).Methods("POST")
	r.HandleFunc(prefix+"/auth/tokens", auth_controller.GetTokens).Methods("GET")

	// Password reset routes
	r.HandleFunc(prefix+"/open/auth/reset-password", user_controller.RequestPasswordReset).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/reset-password/confirm", user_controller.ConfirmPasswordReset).Methods("POST")
}

// registerAdminRoutes registers admin-only endpoints
func registerAdminRoutes(r *mux.Router) {
	// User management - admin only
	r.HandleFunc("/user", user_controller.Create).Methods("POST")
	r.HandleFunc("/user", user_controller.ListAll).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Get).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Delete).Methods("DELETE")


	// Database management - admin only
	r.HandleFunc("/manage/dump", manage_controller.DumpDatabase).Methods("GET")
	r.HandleFunc("/manage/restore", manage_controller.RestoreDatabase).Methods("POST")
}

// registerUserRoutes registers user-specific endpoints (authenticated users can access their own profiles)
func registerUserRoutes(r *mux.Router, prefix string) {
	// User profile management
	r.HandleFunc(prefix+"/user/profile", user_controller.GetProfile).Methods("GET")
	r.HandleFunc(prefix+"/user/profile", user_controller.UpdateProfile).Methods("PUT")
	r.HandleFunc(prefix+"/user/password", user_controller.ChangePassword).Methods("PUT")
	r.HandleFunc(prefix+"/user/account", user_controller.DeleteAccount).Methods("DELETE")
}

func registerCashRoute(r *mux.Router, prefix string) {
	// Create
	r.HandleFunc(prefix+"/cash/expense", cash_flow_controller.CreateExpense).Methods("POST")
	r.HandleFunc(prefix+"/cash/income", cash_flow_controller.CreateIncome).Methods("POST")

	// Read
	r.HandleFunc(prefix+"/cash", cash_flow_controller.ListAll).Methods("GET")
	r.HandleFunc(prefix+"/cash/range", cash_flow_controller.QueryByDateRange).Methods("GET")
	r.HandleFunc(prefix+"/cash/date/{date}", cash_flow_controller.QueryByDate).Methods("GET")
	r.HandleFunc(prefix+"/cash/date/{date}", cash_flow_controller.DeleteByDate).Methods("DELETE")
	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.QueryById).Methods("GET")

	// Summary endpoints
	r.HandleFunc(prefix+"/cash/summary/daily/{date}", cash_flow_controller.GetDailySummary).Methods("GET")
	r.HandleFunc(prefix+"/cash/summary/monthly/{month}", cash_flow_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc(prefix+"/cash/summary/yearly/{year}", cash_flow_controller.GetYearlySummary).Methods("GET")

	// Update
	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.UpdateById).Methods("PUT")

	// Delete
	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.DeleteById).Methods("DELETE")
}

func registerCategoryRoute(r *mux.Router, prefix string) {
	// Create
	r.HandleFunc(prefix+"/category", category_controller.Create).Methods("POST")
	// Read all categories with filtering
	r.HandleFunc(prefix+"/category", category_controller.ListAll).Methods("GET")
	// Read children categories - RESTful design: parent/{id}/children
	r.HandleFunc(prefix+"/category/name/{name}", category_controller.QueryByName).Methods("GET")
	r.HandleFunc(prefix+"/category/{parent_id}/children", category_controller.QueryChildren).Methods("GET")
	r.HandleFunc(prefix+"/category/tree", category_controller.Tree).Methods("GET")
	// Read specific category - must be after specific paths
	r.HandleFunc(prefix+"/category/{id}", category_controller.QueryById).Methods("GET")

	// Update
	r.HandleFunc(prefix+"/category/{id}", category_controller.UpdateById).Methods("PUT")

	// Delete
	r.HandleFunc(prefix+"/category/{id}", category_controller.DeleteById).Methods("DELETE")
}

func registerStatisticRoute(r *mux.Router, prefix string) {
	// Export/Import user-specific data
	r.HandleFunc(prefix+"/statistic/export", statistic_controller.ExportData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/import", statistic_controller.ImportData).Methods("POST")

	// Summary endpoints
	r.HandleFunc(prefix+"/statistic/summary/daily/{date}", statistic_controller.GetDailySummary).Methods("GET")
	r.HandleFunc(prefix+"/statistic/summary/monthly/{month}", statistic_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc(prefix+"/statistic/summary/yearly/{year}", statistic_controller.GetYearlySummary).Methods("GET")

	// Breakdown endpoints
	r.HandleFunc(prefix+"/statistic/breakdown/daily/{date}", statistic_controller.GetDailyBreakdown).Methods("GET")
	r.HandleFunc(prefix+"/statistic/breakdown/monthly/{month}", statistic_controller.GetMonthlyBreakdown).Methods("GET")
	r.HandleFunc(prefix+"/statistic/breakdown/yearly/{year}", statistic_controller.GetYearlyBreakdown).Methods("GET")

	// Trends endpoints
	r.HandleFunc(prefix+"/statistic/trends/daily/{date}", statistic_controller.GetDailyTrends).Methods("GET")
	r.HandleFunc(prefix+"/statistic/trends/monthly/{month}", statistic_controller.GetMonthlyTrends).Methods("GET")
	r.HandleFunc(prefix+"/statistic/trends/yearly/{year}", statistic_controller.GetYearlyTrends).Methods("GET")

	// Top expenses endpoints
	r.HandleFunc(prefix+"/statistic/top/daily/{date}", statistic_controller.GetDailyTopExpenses).Methods("GET")
	r.HandleFunc(prefix+"/statistic/top/monthly/{month}", statistic_controller.GetMonthlyTopExpenses).Methods("GET")
	r.HandleFunc(prefix+"/statistic/top/yearly/{year}", statistic_controller.GetYearlyTopExpenses).Methods("GET")

	// Dashboard visualization endpoints
	r.HandleFunc(prefix+"/statistic/dashboard/{period}/{date}", statistic_controller.GetDashboardOverview).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/income-expense/{period}/{date}", statistic_controller.GetIncomeExpenseChartData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/category-distribution/{period}/{date}", statistic_controller.GetCategoryDistributionData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/monthly-comparison/{year}", statistic_controller.GetMonthlyComparisonData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/spending-heatmap/{year}", statistic_controller.GetSpendingHeatmapData).Methods("GET")
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
	apiVersion := util.GetConfigByKey("api.version")
	if apiVersion == "" {
		apiVersion = "v0"
	}
	apiPrefix := "/api/" + apiVersion

	response := map[string]interface{}{
		"version":     model.Version,
		"name":        "CashLenX API",
		"description": "Personal finance management API",
		"endpoints": map[string][]string{
			"open": {
				"GET " + apiPrefix + "/open/health",
				"GET " + apiPrefix + "/open/version",
				"POST " + apiPrefix + "/open/auth/login",
				"POST " + apiPrefix + "/open/auth/register",
				"POST " + apiPrefix + "/open/auth/reset-password",
				"POST " + apiPrefix + "/open/auth/reset-password/confirm",
			},
			"admin": {
				"POST " + apiPrefix + "/admin/user",
				"GET " + apiPrefix + "/admin/user",
				"GET " + apiPrefix + "/admin/user/{id}",
				"DELETE " + apiPrefix + "/admin/user/{id}",
				"GET " + apiPrefix + "/admin/manage/dump",
				"POST " + apiPrefix + "/admin/manage/restore",
			},
			"user": {
				"GET " + apiPrefix + "/user/profile",
				"PUT " + apiPrefix + "/user/profile",
				"PUT " + apiPrefix + "/user/password",
				"DELETE " + apiPrefix + "/user/account",
			},
			"cash_flow": {
				"POST " + apiPrefix + "/cash/expense",
				"POST " + apiPrefix + "/cash/income",
				"GET " + apiPrefix + "/cash",
				"GET " + apiPrefix + "/cash?limit=20&offset=0&type=income",
				"GET " + apiPrefix + "/cash/{id}",
				"GET " + apiPrefix + "/cash/date/{date}",
				"GET " + apiPrefix + "/cash/range?from=YYYYMMDD&to=YYYYMMDD",
				"GET " + apiPrefix + "/cash/summary/daily/{date}",
				"GET " + apiPrefix + "/cash/summary/monthly/{month}",
				"GET " + apiPrefix + "/cash/summary/yearly/{year}",
				"PUT " + apiPrefix + "/cash/{id}",
				"DELETE " + apiPrefix + "/cash/{id}",
				"DELETE " + apiPrefix + "/cash/date/{date}",
			},
			"category": {
				"POST " + apiPrefix + "/category",
				"GET " + apiPrefix + "/category",
				"GET " + apiPrefix + "/category?type=income&parent_id=XXX",
				"GET " + apiPrefix + "/category/{id}",
				"GET " + apiPrefix + "/category/name/{name}",
				"GET " + apiPrefix + "/category/{parent_id}/children",
				"GET " + apiPrefix + "/category/tree",
				"PUT " + apiPrefix + "/category/{id}",
				"DELETE " + apiPrefix + "/category/{id}",
			},
			"statistic": {
				"GET " + apiPrefix + "/statistic/export?from_date=YYYYMMDD&to_date=YYYYMMDD&format=xlsx|csv|pdf (binary download)",
				"POST " + apiPrefix + "/statistic/import?file_path=path",
				"GET " + apiPrefix + "/statistic/summary/daily/{date}",
				"GET " + apiPrefix + "/statistic/summary/monthly/{month}",
				"GET " + apiPrefix + "/statistic/summary/yearly/{year}",
				"GET " + apiPrefix + "/statistic/breakdown/daily/{date}",
				"GET " + apiPrefix + "/statistic/breakdown/monthly/{month}",
				"GET " + apiPrefix + "/statistic/breakdown/yearly/{year}",
				"GET " + apiPrefix + "/statistic/trends/daily/{date}",
				"GET " + apiPrefix + "/statistic/trends/monthly/{month}",
				"GET " + apiPrefix + "/statistic/trends/yearly/{year}",
				"GET " + apiPrefix + "/statistic/top/daily/{date}?limit=10",
				"GET " + apiPrefix + "/statistic/top/monthly/{month}?limit=10",
				"GET " + apiPrefix + "/statistic/top/yearly/{year}?limit=10",
				"GET " + apiPrefix + "/statistic/dashboard/{period}/{date}",
				"GET " + apiPrefix + "/statistic/chart/income-expense/{period}/{date}",
				"GET " + apiPrefix + "/statistic/chart/category-distribution/{period}/{date}?type=income|expense",
				"GET " + apiPrefix + "/statistic/chart/monthly-comparison/{year}",
				"GET " + apiPrefix + "/statistic/chart/spending-heatmap/{year}",
			},
		},
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
