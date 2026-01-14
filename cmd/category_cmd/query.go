package category_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query for category data",
	RunE: func(cmd *cobra.Command, args []string) error {
		if plainId != "" {
			category, err := category_service.QueryByIdForUser(plainId, userId)
			if err != nil {
				return err
			}
			fmt.Printf("category 0: %s (ID: %s)\n", category.Name, category.Id.Hex())
			return nil
		}

		if categoryName != "" {
			category, err := category_service.QueryByNameForUser(categoryName, userId)
			if err != nil {
				return err
			}
			fmt.Printf("category 0: %s (ID: %s)\n", category.Name, category.Id.Hex())
			return nil
		}

		if parentPlainId != "" {
			categories, err := category_service.GetChildCategoriesForUser(parentPlainId, userId, "")
			if err != nil {
				return err
			}
			for index, categoryEntity := range categories {
				fmt.Printf("category %d: %s (ID: %s)\n", index, categoryEntity.Name, categoryEntity.Id.Hex())
			}
			return nil
		}

		return fmt.Errorf("must provide query criteria (id, name, or parent)")
	},
}

func init() {
	queryCmd.Flags().StringVarP(
		&plainId, "id", "i", "", "query by id")
	queryCmd.Flags().StringVarP(
		&parentPlainId, "parent", "p", "", "query by parent id")
	queryCmd.Flags().StringVarP(
		&categoryName, "name", "n", "", "query by name")
	CategoryCmd.AddCommand(queryCmd)
}
