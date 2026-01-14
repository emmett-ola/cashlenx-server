package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var (
	summaryPeriod string
	summaryDate   string
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "show cash_flow summary",
	Long: `Show summary of cash flows for different periods.
Periods: daily, monthly, yearly
Examples:
  cashlenx cash summary --period daily --date 2024-01-15
  cashlenx cash summary --period monthly --date 2024-01
  cashlenx cash summary --period yearly --date 2024`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if summaryPeriod == "" {
			return errors.New("period is required (daily, monthly, yearly)")
		}

		if summaryDate == "" {
			return errors.New("date is required (format depends on period)")
		}

		summary, err := cash_flow_service.GetSummary(summaryPeriod, summaryDate)
		if err != nil {
			return err
		}

		fmt.Printf("\n=== %s Summary for %s ===\n", summaryPeriod, summaryDate)
		fmt.Printf("Total Income:  %.2f\n", summary.TotalIncome)
		fmt.Printf("Total Expense: %.2f\n", summary.TotalExpense)
		fmt.Printf("Balance:       %.2f\n", summary.Balance)
		fmt.Printf("Transactions:  %d\n", summary.TransactionCount)

		if len(summary.CategoryBreakdown) > 0 {
			fmt.Printf("\n--- Category Breakdown ---\n")
			for category, amount := range summary.CategoryBreakdown {
				fmt.Printf("  %-20s: %.2f\n", category, amount)
			}
		}

		return nil
	},
}

func init() {
	summaryCmd.Flags().StringVarP(
		&summaryPeriod, "period", "p", "", "summary period (daily/monthly/yearly) (required)")
	summaryCmd.Flags().StringVarP(
		&summaryDate, "date", "d", "", "date for summary (format: YYYY-MM-DD for daily, YYYY-MM for monthly, YYYY for yearly) (required)")

	summaryCmd.MarkFlagRequired("period")
	summaryCmd.MarkFlagRequired("date")
	CashCmd.AddCommand(summaryCmd)
}
