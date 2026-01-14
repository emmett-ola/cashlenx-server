package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	summaryPeriod string // daily, monthly, yearly
	summaryDate   string // YYYY-MM-DD, YYYY-MM, or YYYY
	summaryUserId string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show financial summary",
	Long: `Display financial summary for the specified period.
Only includes your own transactions.

Periods:
  daily   - Summary for a specific day (date format: YYYY-MM-DD)
  monthly - Summary for a specific month (date format: YYYY-MM)
  yearly  - Summary for a specific year (date format: YYYY)`,
	Example: `  cashlenx statistic summary -p daily -d 2024-01-15 -u <userId>
  cashlenx statistic summary -p monthly -d 2024-01 -u <userId>
  cashlenx statistic summary -p yearly -d 2024 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if summaryUserId == "" {
			return fmt.Errorf("user ID is required (use --user flag)")
		}

		summary, err := statistic_service.GetSummaryForUser(summaryPeriod, summaryDate, summaryUserId)
		if err != nil {
			return fmt.Errorf("failed to get summary: %w", err)
		}

		// Display summary
		fmt.Printf("\n=== Financial Summary (%s %s) ===\n", summary.PeriodType, summary.Period)
		fmt.Printf("Income:              %.2f (%d transactions)\n", summary.Income, summary.IncomeCount)
		fmt.Printf("Expense:             %.2f (%d transactions)\n", summary.Expense, summary.ExpenseCount)
		fmt.Printf("Balance:             %.2f\n", summary.Balance)
		fmt.Printf("Total Transactions:  %d\n", summary.TransactionCount)

		if summary.TransactionCount > 0 {
			fmt.Printf("Average Transaction: %.2f\n", summary.AverageTransaction)
		}

		if len(summary.Categories) > 0 {
			fmt.Println("\nTop Categories:")
			count := 0
			for cat, amount := range summary.Categories {
				fmt.Printf("  - %s: %.2f\n", cat, amount)
				count++
				if count >= 5 { // Show top 5
					break
				}
			}
		}

		return nil
	},
}

func init() {
	summaryCmd.Flags().StringVarP(&summaryPeriod, "period", "p", "monthly", "period type: daily, monthly, yearly")
	summaryCmd.Flags().StringVarP(&summaryDate, "date", "d", "", "date for summary (required)")
	summaryCmd.Flags().StringVarP(&summaryUserId, "user", "u", "", "user ID (required)")
	summaryCmd.MarkFlagRequired("date")
	summaryCmd.MarkFlagRequired("user")
}
