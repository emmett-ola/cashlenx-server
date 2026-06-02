package category_cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	force bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete category data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if plainId == "" && categoryName == "" {
			return errors.New("must provide either id or name to delete")
		}

		if plainId != "" {
			_, err := deleteCategoryForUser(plainId, userId, force)
			return err
		}

		if categoryName != "" {
			// Query by name first to get ID
			category, err := queryCategoryByNameForUser(categoryName, userId)
			if err != nil {
				return err
			}
			_, err = deleteCategoryForUser(category.Id.Hex(), userId, force)
			return err
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(
		&plainId, "id", "i", "", "delete by id")
	deleteCmd.Flags().StringVarP(
		&categoryName, "name", "n", "", "delete by name")
	deleteCmd.Flags().BoolVarP(
		&force, "force", "f", false, "force delete even if category has associated cash flows")
	CategoryCmd.AddCommand(deleteCmd)
}
