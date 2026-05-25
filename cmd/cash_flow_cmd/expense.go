package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var expenseCmd = &cobra.Command{
	Use:   "expense",
	Short: "add new expense cash_flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureCashUser(); err != nil {
			return err
		}
		belongsDate, _ := cmd.Flags().GetString("date")
		categoryName, _ := cmd.Flags().GetString("category")
		amount, _ := cmd.Flags().GetFloat64("amount")
		descriptionExact, _ := cmd.Flags().GetString("description")

		if !cash_flow_service.IsExpenseRequiredFiledSatisfied(categoryName, amount) {
			return errors.New("some required fields are empty")
		}
		cashFlowEntity, err := cash_flow_service.SaveExpense(belongsDate, categoryName, amount, descriptionExact, cashUserId)
		if err != nil {
			return err
		}
		fmt.Println("cash_flow ", 0, ": ", cashFlowEntity.ToString())
		return nil
	},
}

func init() {
	expenseCmd.Flags().StringP("date", "b", "", "flow's belongs-date (optional, blank for today)")
	expenseCmd.Flags().StringP("category", "c", "", "flow's category name (required)")
	expenseCmd.Flags().Float64P("amount", "a", 0.00, "flow's amount (required)")
	expenseCmd.Flags().StringP("description", "d", "", "flow's description (optional, could be blank)")

	// Mark required flags
	_ = expenseCmd.MarkFlagRequired("category")
	_ = expenseCmd.MarkFlagRequired("amount")
	CashCmd.AddCommand(expenseCmd)
}
