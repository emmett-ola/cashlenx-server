package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	breakdownPeriod string
	breakdownDate   string
	breakdownUserId string
)

var breakdownCmd = &cobra.Command{
	Use:   "breakdown",
	Short: "Show category breakdown analysis",
	Long: `Display spending breakdown by category for the specified period.
Only includes your own transactions.`,
	Example: `  cashlenx statistic breakdown -p month -d 2024-01 -u <userId>
  cashlenx statistic breakdown -p year -d 2024 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if breakdownUserId == "" {
			return fmt.Errorf("user ID is required (use --user flag)")
		}

		breakdown, err := statistic_service.GetBreakdownForUser(breakdownPeriod, breakdownDate, breakdownUserId)
		if err != nil {
			return fmt.Errorf("failed to get breakdown: %w", err)
		}

		// Display breakdown
		fmt.Printf("\n=== Category Breakdown (%s %s) ===\n", breakdownPeriod, breakdown.Period)
		fmt.Printf("Total Income:  %.2f\n", breakdown.TotalIncome)
		fmt.Printf("Total Expense: %.2f\n\n", breakdown.TotalExpense)

		if len(breakdown.ExpenseCategories) > 0 {
			fmt.Println("Expense Categories:")
			for _, cat := range breakdown.ExpenseCategories {
				fmt.Printf("  %-25s %.2f (%.1f%%) - %d transactions\n",
					cat.Category, cat.Amount, cat.Percentage, cat.Count)
			}
		}

		if len(breakdown.IncomeCategories) > 0 {
			fmt.Println("\nIncome Categories:")
			for _, cat := range breakdown.IncomeCategories {
				fmt.Printf("  %-25s %.2f (%.1f%%) - %d transactions\n",
					cat.Category, cat.Amount, cat.Percentage, cat.Count)
			}
		}

		return nil
	},
}

func init() {
	breakdownCmd.Flags().StringVarP(&breakdownPeriod, "period", "p", "month", "period type: month, year")
	breakdownCmd.Flags().StringVarP(&breakdownDate, "date", "d", "", "date (YYYY-MM for month, YYYY for year) (required)")
	breakdownCmd.Flags().StringVarP(&breakdownUserId, "user", "u", "", "user ID (required)")
	breakdownCmd.MarkFlagRequired("date")
	breakdownCmd.MarkFlagRequired("user")
}
