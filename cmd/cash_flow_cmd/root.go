package cash_flow_cmd

import (
	"errors"

	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var (
	plainId          string
	descriptionFuzzy string
	cashUserId       string
)

var CashCmd = &cobra.Command{
	Use:   "cash",
	Short: "manage cash flow transactions",
	Long: `Manage cash flow transactions (income and expenses).

Available sub-commands:
  income   - Add new income transaction
  expense  - Add new expense transaction
  update   - Update existing transaction
  delete   - Delete transaction(s)
  query    - Query transactions by filters
  list     - List all transactions with pagination
  range    - Query transactions by date range
  summary  - Show financial summary`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must provide a valid sub command")
	},
}

func ensureCashUser() error {
	if cashUserId != "" {
		return nil
	}
	var err error
	cashUserId, err = user_service.GetDefaultAdminUserId()
	return err
}

func init() {
	CashCmd.PersistentFlags().StringVarP(&cashUserId, "user", "u", "", "user ID (required)")
}
