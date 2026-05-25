package user_cmd

import (
	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/spf13/cobra"
)

var userId string

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "User profile, configuration, account, and backup commands",
	Long:  `User-scoped commands that mirror /api/user/* endpoints.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cli_auth.RequireUserID(&userId)
	},
}

func init() {
	UserCmd.PersistentFlags().StringVarP(&userId, "user", "u", "", "user ID; must match the logged-in user")

	UserCmd.AddCommand(profileCmd)
	UserCmd.AddCommand(configurationCmd)
	UserCmd.AddCommand(passwordCmd)
	UserCmd.AddCommand(emailCmd)
	UserCmd.AddCommand(accountCmd)
	UserCmd.AddCommand(databaseCmd)
}
