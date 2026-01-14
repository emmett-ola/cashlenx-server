package statistic_service

// Summary represents a financial summary for a period
type Summary struct {
	Period             string             `json:"period"`
	PeriodType         string             `json:"period_type"`
	Income             float64            `json:"income"`
	Expense            float64            `json:"expense"`
	Balance            float64            `json:"balance"`
	TransactionCount   int                `json:"transaction_count"`
	IncomeCount        int                `json:"income_count"`
	ExpenseCount       int                `json:"expense_count"`
	AverageTransaction float64            `json:"average_transaction"`
	Categories         map[string]float64 `json:"categories"`
}

// CategoryBreakdownItem represents a single category in the breakdown
type CategoryBreakdownItem struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	Count      int     `json:"count"`
}

// Breakdown represents category breakdown analysis
type Breakdown struct {
	Period            string                  `json:"period"`
	TotalExpense      float64                 `json:"total_expense"`
	TotalIncome       float64                 `json:"total_income"`
	ExpenseCategories []CategoryBreakdownItem `json:"expense_categories"`
	IncomeCategories  []CategoryBreakdownItem `json:"income_categories"`
}

// TrendDataPoint represents a single data point in the trend
type TrendDataPoint struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

// TrendAnalysis represents the trend analysis results
type TrendAnalysis struct {
	IncomeTrend           string  `json:"income_trend"`
	ExpenseTrend          string  `json:"expense_trend"`
	AverageMonthlyExpense float64 `json:"average_monthly_expense"`
}

// Trends represents spending trends over time
type Trends struct {
	Period     string           `json:"period"`
	PeriodType string           `json:"period_type"`
	DataPoints []TrendDataPoint `json:"data_points"`
	Trends     TrendAnalysis    `json:"trends"`
}

// TopExpense represents a single top expense
type TopExpense struct {
	ID          string  `json:"id"`
	Date        string  `json:"date"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Percentage  float64 `json:"percentage"`
}

// TopExpenses represents the top N expenses
type TopExpenses struct {
	Period       string       `json:"period"`
	Limit        int          `json:"limit"`
	TotalExpense float64      `json:"total_expense"`
	Expenses     []TopExpense `json:"expenses"`
}

// DashboardOverview represents comprehensive dashboard data
type DashboardOverview struct {
	Period        string                  `json:"period"`
	PeriodType    string                  `json:"period_type"`
	Summary       *Summary                `json:"summary"`
	TopCategories []CategoryBreakdownItem `json:"top_categories"`
	RecentTrend   string                  `json:"recent_trend"`
	QuickStats    QuickStats              `json:"quick_stats"`
}

// QuickStats represents quick statistics for dashboard
type QuickStats struct {
	TotalTransactions int     `json:"total_transactions"`
	AverageDaily      float64 `json:"average_daily"`
	HighestExpense    float64 `json:"highest_expense"`
	LowestExpense     float64 `json:"lowest_expense"`
}

// IncomeExpenseChartData represents time-series data for charts
type IncomeExpenseChartData struct {
	Labels   []string  `json:"labels"`
	Income   []float64 `json:"income"`
	Expense  []float64 `json:"expense"`
	Balance  []float64 `json:"balance"`
	Period   string    `json:"period"`
	FromDate string    `json:"from_date"`
	ToDate   string    `json:"to_date"`
}

// CategoryDistribution represents pie/donut chart data
type CategoryDistribution struct {
	Labels      []string  `json:"labels"`
	Values      []float64 `json:"values"`
	Percentages []float64 `json:"percentages"`
	Colors      []string  `json:"colors"`
	Total       float64   `json:"total"`
	Type        string    `json:"type"` // income or expense
}

// MonthlyComparison represents month-over-month comparison data
type MonthlyComparison struct {
	Year    string    `json:"year"`
	Months  []string  `json:"months"`
	Income  []float64 `json:"income"`
	Expense []float64 `json:"expense"`
	Balance []float64 `json:"balance"`
}

// SpendingHeatmap represents calendar heatmap data
type SpendingHeatmap struct {
	Year string             `json:"year"`
	Data []HeatmapDataPoint `json:"data"`
	Max  float64            `json:"max"`
	Min  float64            `json:"min"`
}

// HeatmapDataPoint represents a single day's data point
type HeatmapDataPoint struct {
	Date   string  `json:"date"`   // YYYY-MM-DD
	Amount float64 `json:"amount"` // Total spending for the day
	Count  int     `json:"count"`  // Number of transactions
}
