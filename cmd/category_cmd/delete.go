package category_cmd

import (
	"errors"

	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete category data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if plainId == "" && categoryName == "" {
			return errors.New("must provide either id or name to delete")
		}

		if plainId != "" {
			_, err := category_service.DeleteByIdForUser(plainId, userId)
			return err
		}

		if categoryName != "" {
			// Query by name first to get ID
			category, err := category_service.QueryByNameForUser(categoryName, userId)
			if err != nil {
				return err
			}
			_, err = category_service.DeleteByIdForUser(category.Id.Hex(), userId)
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
	CategoryCmd.AddCommand(deleteCmd)
}
