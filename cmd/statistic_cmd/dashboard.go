package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var (
	dashboardPeriod string
	dashboardDate   string
	dashboardUserId string
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show dashboard overview",
	Long: `Display dashboard overview data for the specified period.
Only includes your own transactions.`,
	Example: `  cashlenx statistic dashboard -p monthly -d 2026-01 -u <userId>
  cashlenx statistic dashboard -p yearly -d 2026 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dashboardUserId == "" {
			var err error
			dashboardUserId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}

		dashboard, err := statistic_service.GetDashboardOverviewForUser(dashboardPeriod, dashboardDate, dashboardUserId)
		if err != nil {
			return fmt.Errorf("failed to get dashboard overview: %w", err)
		}

		fmt.Printf("\n=== Dashboard Overview (%s %s) ===\n", dashboard.PeriodType, dashboard.Period)
		if dashboard.Summary != nil {
			fmt.Printf("Income:              %.2f (%d transactions)\n", dashboard.Summary.Income, dashboard.Summary.IncomeCount)
			fmt.Printf("Expense:             %.2f (%d transactions)\n", dashboard.Summary.Expense, dashboard.Summary.ExpenseCount)
			fmt.Printf("Balance:             %.2f\n", dashboard.Summary.Balance)
			fmt.Printf("Total Transactions:  %d\n", dashboard.Summary.TransactionCount)
		}

		fmt.Println("\nQuick Stats:")
		fmt.Printf("  Total Transactions: %d\n", dashboard.QuickStats.TotalTransactions)
		fmt.Printf("  Average Daily:      %.2f\n", dashboard.QuickStats.AverageDaily)
		fmt.Printf("  Highest Expense:    %.2f\n", dashboard.QuickStats.HighestExpense)
		fmt.Printf("  Lowest Expense:     %.2f\n", dashboard.QuickStats.LowestExpense)
		fmt.Printf("  Recent Trend:       %s\n", dashboard.RecentTrend)

		if len(dashboard.TopCategories) > 0 {
			fmt.Println("\nTop Expense Categories:")
			for _, cat := range dashboard.TopCategories {
				fmt.Printf("  %-25s %.2f (%.1f%%) - %d transactions\n",
					cat.Category, cat.Amount, cat.Percentage, cat.Count)
			}
		}

		return nil
	},
}

func init() {
	dashboardCmd.Flags().StringVarP(&dashboardPeriod, "period", "p", "monthly", "period type: daily, monthly, yearly")
	dashboardCmd.Flags().StringVarP(&dashboardDate, "date", "d", "", "date for dashboard (required)")
	dashboardCmd.Flags().StringVarP(&dashboardUserId, "user", "u", "", "user ID (required)")
	dashboardCmd.MarkFlagRequired("date")
}
