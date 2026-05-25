package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	trendsPeriod string
	trendsDate   string
	trendsUserId string
)

var trendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "Show spending trends over time",
	Long: `Display spending trends analysis for the specified period.
Only includes your own transactions.`,
	Example: `  cashlenx statistic trends -p yearly -d 2024 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureStatisticUser(&trendsUserId); err != nil {
			return err
		}

		trends, err := statistic_service.GetTrendsForUser(trendsPeriod, trendsDate, trendsUserId)
		if err != nil {
			return fmt.Errorf("failed to get trends: %w", err)
		}

		// Display trends
		fmt.Printf("\n=== Spending Trends (%s %s) ===\n", trends.PeriodType, trends.Period)

		if len(trends.DataPoints) > 0 {
			fmt.Println("\nMonthly Breakdown:")
			for _, dp := range trends.DataPoints {
				fmt.Printf("  %s: Income %.2f | Expense %.2f | Balance %.2f\n",
					dp.Date, dp.Income, dp.Expense, dp.Balance)
			}
		}

		fmt.Println("\nTrend Analysis:")
		fmt.Printf("  Income Trend:  %s\n", trends.Trends.IncomeTrend)
		fmt.Printf("  Expense Trend: %s\n", trends.Trends.ExpenseTrend)
		fmt.Printf("  Average Monthly Expense: %.2f\n", trends.Trends.AverageMonthlyExpense)

		return nil
	},
}

func init() {
	trendsCmd.Flags().StringVarP(&trendsPeriod, "period", "p", "yearly", "period type: daily, monthly, yearly")
	trendsCmd.Flags().StringVarP(&trendsDate, "date", "d", "", "date (required)")
	trendsCmd.Flags().StringVarP(&trendsUserId, "user", "u", "", "user ID; must match the logged-in user")
	trendsCmd.MarkFlagRequired("date")
}
