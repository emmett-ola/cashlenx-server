package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var (
	fromDate string
	toDate   string
)

var rangeCmd = &cobra.Command{
	Use:   "range",
	Short: "query cash_flow by date range",
	Long: `Query cash flow records within a date range.
Displays all transactions between from-date and to-date (inclusive).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if fromDate == "" || toDate == "" {
			return errors.New("both from-date and to-date are required")
		}

		cashFlowEntityList, err := cash_flow_service.QueryByDateRange(fromDate, toDate)
		if err != nil {
			return err
		}

		if len(cashFlowEntityList) == 0 {
			fmt.Println("No cash flows found in the specified date range")
			return nil
		}

		var totalIncome, totalExpense float64
		for index, cashFlowEntity := range cashFlowEntityList {
			fmt.Println("cash_flow", index, ":", cashFlowEntity.ToString())
			if cashFlowEntity.FlowType == model.FlowTypeIncome {
				totalIncome += cashFlowEntity.Amount
			} else {
				totalExpense += cashFlowEntity.Amount
			}
		}

		fmt.Printf("\n--- Summary ---\n")
		fmt.Printf("Period: %s to %s\n", fromDate, toDate)
		fmt.Printf("Total Records: %d\n", len(cashFlowEntityList))
		fmt.Printf("Total Income: %.2f\n", totalIncome)
		fmt.Printf("Total Expense: %.2f\n", totalExpense)
		fmt.Printf("Balance: %.2f\n", totalIncome-totalExpense)

		return nil
	},
}

func init() {
	rangeCmd.Flags().StringVarP(
		&fromDate, "from", "f", "", "start date (YYYY-MM-DD) (required)")
	rangeCmd.Flags().StringVarP(
		&toDate, "to", "t", "", "end date (YYYY-MM-DD) (required)")

	rangeCmd.MarkFlagRequired("from")
	rangeCmd.MarkFlagRequired("to")
	CashCmd.AddCommand(rangeCmd)
}
