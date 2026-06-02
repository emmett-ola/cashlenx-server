package user_cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	emailNewEmail          string
	emailVerificationToken string
	emailConfirmToken      string
	emailConfirmPassword   string
)

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Change current user's email address",
}

var emailChangeCmd = &cobra.Command{
	Use:   "change",
	Short: "Change email using an email_change verification token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailNewEmail == "" || emailVerificationToken == "" {
			return errors.New("new-email and verification-token are required")
		}
		if err := requestUserEmailChange(userId, emailNewEmail, emailVerificationToken); err != nil {
			return err
		}
		fmt.Println("Email changed successfully")
		return nil
	},
}

var emailConfirmCmd = &cobra.Command{
	Use:   "confirm",
	Short: "Confirm email change with token and password",
	RunE: func(cmd *cobra.Command, args []string) error {
		if emailConfirmToken == "" || emailConfirmPassword == "" {
			return errors.New("token and password are required")
		}
		if err := confirmUserEmailChange(userId, emailConfirmToken, emailConfirmPassword); err != nil {
			return err
		}
		fmt.Println("Email changed successfully")
		return nil
	},
}

func init() {
	emailCmd.AddCommand(emailChangeCmd)
	emailCmd.AddCommand(emailConfirmCmd)

	emailChangeCmd.Flags().StringVar(&emailNewEmail, "new-email", "", "new email address (required)")
	emailChangeCmd.Flags().StringVar(&emailVerificationToken, "verification-token", "", "verification token (required)")
	emailChangeCmd.MarkFlagRequired("new-email")
	emailChangeCmd.MarkFlagRequired("verification-token")

	emailConfirmCmd.Flags().StringVar(&emailConfirmToken, "token", "", "verification token (required)")
	emailConfirmCmd.Flags().StringVar(&emailConfirmPassword, "password", "", "current password (required)")
	emailConfirmCmd.MarkFlagRequired("token")
	emailConfirmCmd.MarkFlagRequired("password")
}
