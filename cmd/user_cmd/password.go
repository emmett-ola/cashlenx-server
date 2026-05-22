package user_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/spf13/cobra"
)

var (
	oldPassword string
	newPassword string
)

var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change current user's password",
	RunE: func(cmd *cobra.Command, args []string) error {
		if oldPassword == "" || newPassword == "" {
			return errors.New("old-password and new-password are required")
		}
		if err := user_service.ChangePasswordService(userId, oldPassword, newPassword); err != nil {
			return err
		}
		fmt.Println("Password changed successfully. Please login again.")
		return nil
	},
}

func init() {
	passwordCmd.Flags().StringVar(&oldPassword, "old-password", "", "old password (required)")
	passwordCmd.Flags().StringVar(&newPassword, "new-password", "", "new password (required)")
	passwordCmd.MarkFlagRequired("old-password")
	passwordCmd.MarkFlagRequired("new-password")
}
