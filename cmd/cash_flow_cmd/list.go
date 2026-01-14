package cash_flow_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var (
	limit    int
	offset   int
	cashType string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all cash_flow records",
	Long: `List all cash flow records with optional filtering and pagination.
Use --type to filter by income/expense, --limit for pagination.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cashFlowEntityList, _, err := cash_flow_service.QueryAll(cashType, "", "", "", "", "", limit, offset)
		if err != nil {
			return err
		}

		if len(cashFlowEntityList) == 0 {
			fmt.Println("No cash flows found")
			return nil
		}

		var totalIncome, totalExpense float64
		for index, cashFlowEntity := range cashFlowEntityList {
			fmt.Println("cash_flow", index+offset, ":", cashFlowEntity.ToString())
			if cashFlowEntity.FlowType == model.FlowTypeIncome {
				totalIncome += cashFlowEntity.Amount
			} else {
				totalExpense += cashFlowEntity.Amount
			}
		}

		fmt.Printf("\n--- Summary (showing %d records) ---\n", len(cashFlowEntityList))
		fmt.Printf("Total Income: %.2f\n", totalIncome)
		fmt.Printf("Total Expense: %.2f\n", totalExpense)
		fmt.Printf("Balance: %.2f\n", totalIncome-totalExpense)

		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(
		&limit, "limit", "l", 50, "maximum number of records to return")
	listCmd.Flags().IntVarP(
		&offset, "offset", "o", 0, "number of records to skip")
	listCmd.Flags().StringVarP(
		&cashType, "type", "t", "", "filter by type (income/expense)")

	CashCmd.AddCommand(listCmd)
}
