package category_cmd

import (
	"errors"

	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var (
	plainId       string
	parentPlainId string
	categoryName  string
	catType       string
	userId        string
)

var CategoryCmd = &cobra.Command{
	Use:   "category",
	Short: "manage transaction categories",
	Long: `Manage transaction categories for organizing cash flows.

Available sub-commands:
  create - Create new category
  update - Update existing category
  delete - Delete category
  query  - Query categories by filters
  list   - List all categories`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if userId == "" {
			var err error
			userId, err = user_service.GetDefaultAdminUserId()
			if err != nil {
				return err
			}
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must provide a valid sub command")
	},
}

func init() {
	CategoryCmd.PersistentFlags().StringVarP(&userId, "user", "u", "", "user ID (required)")
}