package category_cmd

import (
	"errors"

	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
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
		return cli_auth.RequireUserID(&userId)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must provide a valid sub command")
	},
}

func init() {
	CategoryCmd.PersistentFlags().StringVarP(&userId, "user", "u", "", "user ID; must match the logged-in user")
}
