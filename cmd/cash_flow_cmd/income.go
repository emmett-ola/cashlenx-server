package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var incomeCmd = &cobra.Command{
	Use:   "income",
	Short: "add new income cash_flow",
	RunE: func(cmd *cobra.Command, args []string) error {
	belongsDate, _ := cmd.Flags().GetString("date")
  categoryName, _ := cmd.Flags().GetString("category")
  amount, _ := cmd.Flags().GetFloat64("amount")
  descriptionExact, _ := cmd.Flags().GetString("description")
incomeUserId, err := cmd.Flags().GetString("user")
		

		if !cash_flow_service.IsIncomeRequiredFiledSatisfied(categoryName, amount) {
			return errors.New("some required fields are empty")
		}
		cashFlowEntity, err := cash_flow_service.SaveIncome(belongsDate, categoryName, amount, descriptionExact, incomeUserId)
		if err != nil {
			return err
		}
		fmt.Println("cash_flow ", 0, ": ", cashFlowEntity.ToString())
		return nil
	},
}

func init() {
	incomeCmd.Flags().StringP("date", "b", "", "flow's belongs-date (optional, blank for today)")
	incomeCmd.Flags().StringP("category", "c", "", "flow's category name (required)")
	incomeCmd.Flags().Float64P("amount", "a", 0.00, "flow's amount (required)")
	incomeCmd.Flags().StringP("description", "d", "", "flow's description (optional, could be blank)")
	incomeCmd.Flags().StringP("user", "u", "", "user ID (required)")

	// Mark required flags
	_ = incomeCmd.MarkFlagRequired("category")
	_ = incomeCmd.MarkFlagRequired("amount")
	_ = incomeCmd.MarkFlagRequired("user")
	CashCmd.AddCommand(incomeCmd)
}
