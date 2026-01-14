package statistic_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/spf13/cobra"
)

var (
	topLimit   int
	topPeriod  string
	topDate    string
	topUserId  string
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top N expenses",
	Long: `Display the top N expenses for the specified period.
Only includes your own transactions.`,
	Example: `  cashlenx statistic top -n 10 -p month -d 2024-01 -u <userId>
  cashlenx statistic top -n 20 -p year -d 2024 -u <userId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if topUserId == "" {
			return fmt.Errorf("user ID is required (use --user flag)")
		}

		topExpenses, err := statistic_service.GetTopExpensesForUser(topLimit, topPeriod, topDate, topUserId)
		if err != nil {
			return fmt.Errorf("failed to get top expenses: %w", err)
		}

		// Display top expenses
		fmt.Printf("\n=== Top %d Expenses (%s %s) ===\n", topLimit, topPeriod, topExpenses.Period)
		fmt.Printf("Total Expense: %.2f\n\n", topExpenses.TotalExpense)

		if len(topExpenses.Expenses) > 0 {
			fmt.Println("Rank | Date       | Category         | Amount    | %     | Description")
			fmt.Println("-----|------------|------------------|-----------|-------|------------------")
			for i, exp := range topExpenses.Expenses {
				desc := exp.Description
				if len(desc) > 18 {
					desc = desc[:15] + "..."
				}
				fmt.Printf("%4d | %s | %-16s | %9.2f | %5.1f%% | %s\n",
					i+1, exp.Date, exp.Category, exp.Amount, exp.Percentage, desc)
			}
		} else {
			fmt.Println("No expenses found for the specified period")
		}

		return nil
	},
}

func init() {
	topCmd.Flags().IntVarP(&topLimit, "number", "n", 10, "number of top expenses to show")
	topCmd.Flags().StringVarP(&topPeriod, "period", "p", "month", "period type: month, year")
	topCmd.Flags().StringVarP(&topDate, "date", "d", "", "date (YYYY-MM for month, YYYY for year) (required)")
	topCmd.Flags().StringVarP(&topUserId, "user", "u", "", "user ID (required)")
	topCmd.MarkFlagRequired("date")
	topCmd.MarkFlagRequired("user")
}
