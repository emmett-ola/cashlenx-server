package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var chartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Show chart data for visualizations",
	Long: `Show chart data used by dashboard visualizations.
Only includes your own transactions.`,
}

var (
	incomeExpensePeriod string
	incomeExpenseDate   string
	incomeExpenseUserId string

	categoryDistributionPeriod string
	categoryDistributionDate   string
	categoryDistributionType   string
	categoryDistributionUserId string

	monthlyComparisonYear   string
	monthlyComparisonUserId string

	spendingHeatmapYear   string
	spendingHeatmapUserId string
)

var incomeExpenseChartCmd = &cobra.Command{
	Use:   "income-expense",
	Short: "Show income/expense time-series chart data",
	Example: `  cashlenx statistic chart income-expense -p monthly -d 2026-01 -u <userId>
  cashlenx statistic chart income-expense -p yearly -d 2026 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if incomeExpenseUserId == "" {
			var err error
			incomeExpenseUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

		chartData, err := statistic_service.GetIncomeExpenseChartDataForUser(incomeExpensePeriod, incomeExpenseDate, incomeExpenseUserId)
		if err != nil {
			return fmt.Errorf("failed to get income/expense chart data: %w", err)
		}

		fmt.Printf("\n=== Income/Expense Chart (%s %s) ===\n", chartData.Period, incomeExpenseDate)
		fmt.Printf("Range: %s to %s\n\n", chartData.FromDate, chartData.ToDate)
		fmt.Println("Date       | Income    | Expense   | Balance")
		fmt.Println("-----------|-----------|-----------|-----------")
		for i, label := range chartData.Labels {
			fmt.Printf("%-10s | %9.2f | %9.2f | %9.2f\n",
				label, chartData.Income[i], chartData.Expense[i], chartData.Balance[i])
		}

		return nil
	},
}

var categoryDistributionChartCmd = &cobra.Command{
	Use:   "category-distribution",
	Short: "Show category distribution chart data",
	Example: `  cashlenx statistic chart category-distribution -p monthly -d 2026-01 -t expense -u <userId>
  cashlenx statistic chart category-distribution -p yearly -d 2026 -t income -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if categoryDistributionUserId == "" {
			var err error
			categoryDistributionUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

		distribution, err := statistic_service.GetCategoryDistributionForUser(categoryDistributionPeriod, categoryDistributionDate, categoryDistributionType, categoryDistributionUserId)
		if err != nil {
			return fmt.Errorf("failed to get category distribution chart data: %w", err)
		}

		fmt.Printf("\n=== Category Distribution (%s %s, %s) ===\n", categoryDistributionPeriod, categoryDistributionDate, distribution.Type)
		fmt.Printf("Total: %.2f\n\n", distribution.Total)
		fmt.Println("Category                  | Amount    | Percent | Color")
		fmt.Println("--------------------------|-----------|---------|----------")
		for i, label := range distribution.Labels {
			fmt.Printf("%-25s | %9.2f | %6.1f%% | %s\n",
				label, distribution.Values[i], distribution.Percentages[i], distribution.Colors[i])
		}

		return nil
	},
}

var monthlyComparisonChartCmd = &cobra.Command{
	Use:     "monthly-comparison",
	Short:   "Show monthly comparison chart data",
	Example: `  cashlenx statistic chart monthly-comparison -y 2026 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monthlyComparisonUserId == "" {
			var err error
			monthlyComparisonUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

		comparison, err := statistic_service.GetMonthlyComparisonForUser(monthlyComparisonYear, monthlyComparisonUserId)
		if err != nil {
			return fmt.Errorf("failed to get monthly comparison chart data: %w", err)
		}

		fmt.Printf("\n=== Monthly Comparison (%s) ===\n", comparison.Year)
		fmt.Println("Month | Income    | Expense   | Balance")
		fmt.Println("------|-----------|-----------|-----------")
		for i, month := range comparison.Months {
			fmt.Printf("%-5s | %9.2f | %9.2f | %9.2f\n",
				month, comparison.Income[i], comparison.Expense[i], comparison.Balance[i])
		}

		return nil
	},
}

var spendingHeatmapChartCmd = &cobra.Command{
	Use:     "spending-heatmap",
	Short:   "Show spending heatmap chart data",
	Example: `  cashlenx statistic chart spending-heatmap -y 2026 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if spendingHeatmapUserId == "" {
			var err error
			spendingHeatmapUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

		heatmap, err := statistic_service.GetSpendingHeatmapForUser(spendingHeatmapYear, spendingHeatmapUserId)
		if err != nil {
			return fmt.Errorf("failed to get spending heatmap chart data: %w", err)
		}

		fmt.Printf("\n=== Spending Heatmap (%s) ===\n", heatmap.Year)
		fmt.Printf("Min: %.2f | Max: %.2f | Active Days: %d\n\n", heatmap.Min, heatmap.Max, len(heatmap.Data))
		fmt.Println("Date       | Amount    | Count")
		fmt.Println("-----------|-----------|------")
		for _, dp := range heatmap.Data {
			fmt.Printf("%s | %9.2f | %5d\n", dp.Date, dp.Amount, dp.Count)
		}

		return nil
	},
}

func init() {
	chartCmd.AddCommand(incomeExpenseChartCmd)
	chartCmd.AddCommand(categoryDistributionChartCmd)
	chartCmd.AddCommand(monthlyComparisonChartCmd)
	chartCmd.AddCommand(spendingHeatmapChartCmd)

	incomeExpenseChartCmd.Flags().StringVarP(&incomeExpensePeriod, "period", "p", "monthly", "period type: daily, monthly, yearly")
	incomeExpenseChartCmd.Flags().StringVarP(&incomeExpenseDate, "date", "d", "", "date for chart (required)")
	incomeExpenseChartCmd.Flags().StringVarP(&incomeExpenseUserId, "user", "u", "", "user ID (required)")
	incomeExpenseChartCmd.MarkFlagRequired("date")

	categoryDistributionChartCmd.Flags().StringVarP(&categoryDistributionPeriod, "period", "p", "monthly", "period type: daily, monthly, yearly")
	categoryDistributionChartCmd.Flags().StringVarP(&categoryDistributionDate, "date", "d", "", "date for chart (required)")
	categoryDistributionChartCmd.Flags().StringVarP(&categoryDistributionType, "type", "t", "expense", "flow type: income or expense")
	categoryDistributionChartCmd.Flags().StringVarP(&categoryDistributionUserId, "user", "u", "", "user ID (required)")
	categoryDistributionChartCmd.MarkFlagRequired("date")

	monthlyComparisonChartCmd.Flags().StringVarP(&monthlyComparisonYear, "year", "y", "", "year for monthly comparison (required)")
	monthlyComparisonChartCmd.Flags().StringVarP(&monthlyComparisonUserId, "user", "u", "", "user ID (required)")
	monthlyComparisonChartCmd.MarkFlagRequired("year")

	spendingHeatmapChartCmd.Flags().StringVarP(&spendingHeatmapYear, "year", "y", "", "year for spending heatmap (required)")
	spendingHeatmapChartCmd.Flags().StringVarP(&spendingHeatmapUserId, "user", "u", "", "user ID (required)")
	spendingHeatmapChartCmd.MarkFlagRequired("year")
}
