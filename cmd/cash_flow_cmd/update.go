package cash_flow_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/cash_flow_service"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update existing cash_flow by id",
	Long: `Update an existing cash flow record by its ID.
You can update amount, category, date, and description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		plainId, _ := cmd.Flags().GetString("id")
		belongsDate, _ := cmd.Flags().GetString("date")
		categoryName, _ := cmd.Flags().GetString("category")
		amount, _ := cmd.Flags().GetFloat64("amount")
		descriptionExact, _ := cmd.Flags().GetString("description")

		if plainId == "" {
			return errors.New("id is required for update operation")
		}

		// Check if at least one field to update is provided
		if amount == 0 && categoryName == "" && belongsDate == "" && descriptionExact == "" {
			return errors.New("at least one field to update must be provided (amount, category, date, or description)")
		}

		cashFlowEntity, err := cash_flow_service.UpdateById(plainId, belongsDate, categoryName, amount, descriptionExact)
		if err != nil {
			return err
		}

		fmt.Println("Updated cash_flow:", cashFlowEntity.ToString())
		return nil
	},
}

func init() {
	updateCmd.Flags().StringP(
		"id", "i", "", "cash_flow id (required)")
	updateCmd.Flags().StringP(
		"date", "b", "", "new belongs-date (optional)")
	updateCmd.Flags().StringP(
		"category", "c", "", "new category name (optional)")
	updateCmd.Flags().Float64P(
		"amount", "a", 0.00, "new amount (optional)")
	updateCmd.Flags().StringP(
		"description", "d", "", "new description (optional)")

	updateCmd.MarkFlagRequired("id")
	CashCmd.AddCommand(updateCmd)
}
