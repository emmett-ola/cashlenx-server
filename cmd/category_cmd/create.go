package category_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create new category",
	RunE: func(cmd *cobra.Command, args []string) error {
		createdCategory, err := createCategoryForUser(categoryName, catType, categoryRemark, parentPlainId, userId)
		if err != nil {
			return err
		}
		fmt.Printf("category 0: %s (ID: %s)\n", createdCategory.Name, createdCategory.Id.Hex())
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(
		&parentPlainId, "parent", "p", "", "category's parent's id (optional)")
	createCmd.Flags().StringVarP(
		&categoryName, "name", "n", "", "category's name (required)")
	createCmd.Flags().StringVarP(
		&catType, "type", "t", "", "category's type (required, must be 'income' or 'expense')")
	createCmd.Flags().StringVarP(
		&categoryRemark, "remark", "r", "", "category remark")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("type")
	CategoryCmd.AddCommand(createCmd)
}
