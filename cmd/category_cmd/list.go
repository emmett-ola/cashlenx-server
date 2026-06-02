package category_cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	categoryLimit    int
	categoryOffset   int
	listCategoryType string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all categories",
	Long:  `List all categories in the system with optional pagination.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if userId == "" {
			return fmt.Errorf("user ID is required")
		}
		categoryEntityList, _, err := queryCategoriesForUser(userId, listCategoryType, categoryLimit, categoryOffset)
		if err != nil {
			return err
		}

		for index, categoryEntity := range categoryEntityList {
			fmt.Printf("category %d: %s (ID: %s)\n", index+categoryOffset, categoryEntity.Name, categoryEntity.Id.Hex())
		}

		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(
		&categoryLimit, "limit", "l", 50, "maximum number of records to return")
	listCmd.Flags().IntVarP(
		&categoryOffset, "offset", "o", 0, "number of records to skip")
	listCmd.Flags().StringVarP(
		&listCategoryType, "type", "t", "", "category type filter (optional)")
	CategoryCmd.AddCommand(listCmd)
}
