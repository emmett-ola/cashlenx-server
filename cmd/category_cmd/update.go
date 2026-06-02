package category_cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update existing category",
	Long: `Update an existing category by its ID.
You can update the category name, type, remark, and parent.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if plainId == "" {
			return errors.New("id is required for update operation")
		}

		if categoryName == "" && parentPlainId == "" && catType == "" && categoryRemark == "" {
			return errors.New("at least one field to update must be provided (name, type, remark, or parent)")
		}

		_, err := updateCategoryForUser(plainId, categoryName, catType, categoryRemark, parentPlainId, userId)
		if err != nil {
			return err
		}

		fmt.Printf("Category updated successfully (ID: %s)\n", plainId)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(
		&plainId, "id", "i", "", "category id (required)")
	updateCmd.Flags().StringVarP(
		&categoryName, "name", "n", "", "new category name (optional)")
	updateCmd.Flags().StringVarP(
		&catType, "type", "t", "", "new category type (optional, must be 'income' or 'expense')")
	updateCmd.Flags().StringVarP(
		&categoryRemark, "remark", "r", "", "new category remark")
	updateCmd.Flags().StringVarP(
		&parentPlainId, "parent", "p", "", "new parent category id (optional)")

	updateCmd.MarkFlagRequired("id")
	CategoryCmd.AddCommand(updateCmd)
}
