package controller

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/macar-x/cashlenx-server/controller/auth_controller"
	"github.com/macar-x/cashlenx-server/controller/cash_flow_controller"
	"github.com/macar-x/cashlenx-server/controller/category_controller"
	"github.com/macar-x/cashlenx-server/controller/manage_controller"
	"github.com/macar-x/cashlenx-server/controller/statistic_controller"
	"github.com/macar-x/cashlenx-server/controller/user_controller"
	"github.com/macar-x/cashlenx-server/controller/verification_controller"
	"github.com/macar-x/cashlenx-server/middleware"
	dbmigrations "github.com/macar-x/cashlenx-server/migrations"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

func StartServer(port int32) {
	tz := util.GetTimezone()
	fmt.Printf("Loaded timezone: %v\n", tz)

	workerID := util.GetConfigInt("snowflake.worker_id", 0)
	if err := util.InitSnowflakeGenerator(workerID); err != nil {
		util.Logger.Warnf("Failed to initialize Snowflake generator with worker ID %d: %v, using default", workerID, err)
	} else {
		fmt.Printf("Snowflake ID generator initialized with worker ID: %d\n", workerID)
	}

	if util.GetConfigByKey("db.type") == "mysql" {
		if err := dbmigrations.Run(database.GetMySqlConnection()); err != nil {
			util.Logger.Fatalw("MySQL migration failed", "error", err)
			return
		}
	}

	user_service.InitAdminUser()
	if err := manage_service.CreateIndexes(); err != nil {
		util.Logger.Warnw("Database index reconciliation failed", "error", err)
	}

	r := mux.NewRouter()

	apiVersion := util.GetConfigByKey("api.version")
	if apiVersion == "" {
		apiVersion = "v0"
	}
	apiPrefix := "/api/" + apiVersion

	registerOpenRoutes(r, apiPrefix)

	adminRouter := r.PathPrefix(apiPrefix + "/admin").Subrouter()
	adminRouter.Use(middleware.Admin)
	registerAdminRoutes(adminRouter)

	registerUserRoutes(r, apiPrefix)
	registerCashRoute(r, apiPrefix)
	registerCategoryRoute(r, apiPrefix)
	registerStatisticRoute(r, apiPrefix)

	handler := buildHTTPHandler(r)

	host := util.GetConfigByKey("server.host")
	addr := fmt.Sprintf(":%d", port)
	displayHost := "localhost"
	if host != "" {
		addr = fmt.Sprintf("%s:%d", host, port)
		displayHost = host
	}
	fmt.Printf("API server is running on http://%s:%d\n", displayHost, port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		util.Logger.Fatalw("API server stopped", "error", err)
	}
}

func buildHTTPHandler(r *mux.Router) http.Handler {
	// CORS stays outermost for API traffic so browser preflight requests are
	// answered before auth or OpenAPI validation can reject them.
	apiHandler := middleware.CORS(
		middleware.Logging(
			middleware.Metrics(r,
				middleware.Auth(
					middleware.SchemaValidation(r),
				),
			),
		),
	)

	root := mux.NewRouter()
	root.Handle("/metrics", middleware.MetricsHandler()).Methods(http.MethodGet)
	if util.GetConfigByKey("env") == "dev" {
		root.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}
	root.PathPrefix("/").Handler(apiHandler)
	return root
}

func registerOpenRoutes(r *mux.Router, prefix string) {
	r.HandleFunc(prefix+"/open/health", healthCheck).Methods("GET")
	r.HandleFunc(prefix+"/open/version", versionInfo).Methods("GET")

	r.HandleFunc(prefix+"/open/auth/login", auth_controller.Login).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/register", auth_controller.Register).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/logout", auth_controller.Logout).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/reset-password", user_controller.RequestPasswordReset).Methods("POST")
	r.HandleFunc(prefix+"/open/auth/reset-password/confirm", user_controller.ConfirmPasswordReset).Methods("POST")

	r.HandleFunc(prefix+"/open/verification/code", verification_controller.SendCode).Methods("POST")
	r.HandleFunc(prefix+"/open/verification/verify", verification_controller.VerifyCode).Methods("POST")

	r.HandleFunc(prefix+"/auth/tokens", auth_controller.GetTokens).Methods("GET")
}

func registerAdminRoutes(r *mux.Router) {
	r.HandleFunc("/user", user_controller.Create).Methods("POST")
	r.HandleFunc("/user", user_controller.ListAll).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Get).Methods("GET")
	r.HandleFunc("/user/{id}", user_controller.Update).Methods("PUT")
	r.HandleFunc("/user/{id}", user_controller.Delete).Methods("DELETE")

	r.HandleFunc("/database/backup", manage_controller.DumpDatabase).Methods("GET")
	r.HandleFunc("/database/restore", manage_controller.RestoreDatabase).Methods("POST")
}

func registerUserRoutes(r *mux.Router, prefix string) {
	r.HandleFunc(prefix+"/user/profile", user_controller.GetProfile).Methods("GET")
	r.HandleFunc(prefix+"/user/profile", user_controller.UpdateProfile).Methods("PUT")
	r.HandleFunc(prefix+"/user/configuration", user_controller.GetConfiguration).Methods("GET")
	r.HandleFunc(prefix+"/user/configuration", user_controller.UpsertConfiguration).Methods("POST")
	r.HandleFunc(prefix+"/user/configuration", user_controller.UpsertConfiguration).Methods("PUT")
	r.HandleFunc(prefix+"/user/password", user_controller.ChangePassword).Methods("PUT")
	r.HandleFunc(prefix+"/user/email/change", user_controller.RequestEmailChange).Methods("POST")
	r.HandleFunc(prefix+"/user/email/confirm", user_controller.ConfirmEmailChange).Methods("POST")
	r.HandleFunc(prefix+"/user/account", user_controller.DeleteAccount).Methods("DELETE")

	r.HandleFunc(prefix+"/user/database/backup", manage_controller.ExportUserData).Methods("GET")
	r.HandleFunc(prefix+"/user/database/restore", manage_controller.ImportUserData).Methods("POST")
}

func registerCashRoute(r *mux.Router, prefix string) {
	r.HandleFunc(prefix+"/cash/expense", cash_flow_controller.CreateExpense).Methods("POST")
	r.HandleFunc(prefix+"/cash/income", cash_flow_controller.CreateIncome).Methods("POST")

	r.HandleFunc(prefix+"/cash", cash_flow_controller.ListAll).Methods("GET")
	r.HandleFunc(prefix+"/cash/range", cash_flow_controller.QueryByDateRange).Methods("GET")
	r.HandleFunc(prefix+"/cash/date/{date}", cash_flow_controller.QueryByDate).Methods("GET")
	r.HandleFunc(prefix+"/cash/date/{date}", cash_flow_controller.DeleteByDate).Methods("DELETE")
	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.QueryById).Methods("GET")

	r.HandleFunc(prefix+"/cash/summary/total", cash_flow_controller.GetTotalSummary).Methods("GET")
	r.HandleFunc(prefix+"/cash/summary/daily/{date}", cash_flow_controller.GetDailySummary).Methods("GET")
	r.HandleFunc(prefix+"/cash/summary/monthly/{month}", cash_flow_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc(prefix+"/cash/summary/yearly/{year}", cash_flow_controller.GetYearlySummary).Methods("GET")

	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.UpdateById).Methods("PUT")
	r.HandleFunc(prefix+"/cash/{id}", cash_flow_controller.DeleteById).Methods("DELETE")
}

func registerCategoryRoute(r *mux.Router, prefix string) {
	r.HandleFunc(prefix+"/category", category_controller.Create).Methods("POST")
	r.HandleFunc(prefix+"/category", category_controller.ListAll).Methods("GET")
	r.HandleFunc(prefix+"/category/name/{name}", category_controller.QueryByName).Methods("GET")
	r.HandleFunc(prefix+"/category/{parent_id}/children", category_controller.QueryChildren).Methods("GET")
	r.HandleFunc(prefix+"/category/tree", category_controller.Tree).Methods("GET")
	r.HandleFunc(prefix+"/category/{id}", category_controller.QueryById).Methods("GET")
	r.HandleFunc(prefix+"/category/{id}", category_controller.UpdateById).Methods("PUT")
	r.HandleFunc(prefix+"/category/{id}", category_controller.DeleteById).Methods("DELETE")
}

func registerStatisticRoute(r *mux.Router, prefix string) {
	r.HandleFunc(prefix+"/statistic/export", statistic_controller.ExportData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/import", statistic_controller.ImportData).Methods("POST")

	r.HandleFunc(prefix+"/statistic/summary/daily/{date}", statistic_controller.GetDailySummary).Methods("GET")
	r.HandleFunc(prefix+"/statistic/summary/monthly/{month}", statistic_controller.GetMonthlySummary).Methods("GET")
	r.HandleFunc(prefix+"/statistic/summary/yearly/{year}", statistic_controller.GetYearlySummary).Methods("GET")

	r.HandleFunc(prefix+"/statistic/breakdown/daily/{date}", statistic_controller.GetDailyBreakdown).Methods("GET")
	r.HandleFunc(prefix+"/statistic/breakdown/monthly/{month}", statistic_controller.GetMonthlyBreakdown).Methods("GET")
	r.HandleFunc(prefix+"/statistic/breakdown/yearly/{year}", statistic_controller.GetYearlyBreakdown).Methods("GET")

	r.HandleFunc(prefix+"/statistic/trends/daily/{date}", statistic_controller.GetDailyTrends).Methods("GET")
	r.HandleFunc(prefix+"/statistic/trends/monthly/{month}", statistic_controller.GetMonthlyTrends).Methods("GET")
	r.HandleFunc(prefix+"/statistic/trends/yearly/{year}", statistic_controller.GetYearlyTrends).Methods("GET")

	r.HandleFunc(prefix+"/statistic/top/daily/{date}", statistic_controller.GetDailyTopExpenses).Methods("GET")
	r.HandleFunc(prefix+"/statistic/top/monthly/{month}", statistic_controller.GetMonthlyTopExpenses).Methods("GET")
	r.HandleFunc(prefix+"/statistic/top/yearly/{year}", statistic_controller.GetYearlyTopExpenses).Methods("GET")

	r.HandleFunc(prefix+"/statistic/dashboard/{period}/{date}", statistic_controller.GetDashboardOverview).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/income-expense/{period}/{date}", statistic_controller.GetIncomeExpenseChartData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/category-distribution/{period}/{date}", statistic_controller.GetCategoryDistributionData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/monthly-comparison/{year}", statistic_controller.GetMonthlyComparisonData).Methods("GET")
	r.HandleFunc(prefix+"/statistic/chart/spending-heatmap/{year}", statistic_controller.GetSpendingHeatmapData).Methods("GET")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "cashlenx-api",
		"message": "API is running",
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}

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
				"POST " + apiPrefix + "/open/auth/logout",
				"POST " + apiPrefix + "/open/verification/code",
				"POST " + apiPrefix + "/open/verification/verify",
				"POST " + apiPrefix + "/open/auth/reset-password",
				"POST " + apiPrefix + "/open/auth/reset-password/confirm",
			},
			"auth": {
				"GET " + apiPrefix + "/auth/tokens",
			},
			"admin": {
				"POST " + apiPrefix + "/admin/user",
				"GET " + apiPrefix + "/admin/user",
				"GET " + apiPrefix + "/admin/user/{id}",
				"PUT " + apiPrefix + "/admin/user/{id}",
				"DELETE " + apiPrefix + "/admin/user/{id}",
				"GET " + apiPrefix + "/admin/database/backup",
				"POST " + apiPrefix + "/admin/database/restore",
			},
			"user": {
				"GET " + apiPrefix + "/user/profile",
				"PUT " + apiPrefix + "/user/profile",
				"GET " + apiPrefix + "/user/configuration",
				"POST " + apiPrefix + "/user/configuration",
				"PUT " + apiPrefix + "/user/configuration",
				"PUT " + apiPrefix + "/user/password",
				"POST " + apiPrefix + "/user/email/change",
				"POST " + apiPrefix + "/user/email/confirm",
				"DELETE " + apiPrefix + "/user/account",
				"GET " + apiPrefix + "/user/database/backup",
				"POST " + apiPrefix + "/user/database/restore",
			},
			"cash_flow": {
				"POST " + apiPrefix + "/cash/expense",
				"POST " + apiPrefix + "/cash/income",
				"GET " + apiPrefix + "/cash",
				"GET " + apiPrefix + "/cash?limit=20&offset=0&type=income",
				"GET " + apiPrefix + "/cash/{id}",
				"GET " + apiPrefix + "/cash/date/{date}",
				"GET " + apiPrefix + "/cash/range?from=YYYYMMDD&to=YYYYMMDD",
				"GET " + apiPrefix + "/cash/summary/total",
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
