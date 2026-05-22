package user_cmd

import (
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var userId string

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "User profile, configuration, account, and backup commands",
	Long:  `User-scoped commands that mirror /api/user/* endpoints.`,
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
}

func init() {
	UserCmd.PersistentFlags().StringVarP(&userId, "user", "u", "", "user ID (required)")

	UserCmd.AddCommand(profileCmd)
	UserCmd.AddCommand(configurationCmd)
	UserCmd.AddCommand(passwordCmd)
	UserCmd.AddCommand(emailCmd)
	UserCmd.AddCommand(accountCmd)
	UserCmd.AddCommand(databaseCmd)
}
