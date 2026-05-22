package open_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/spf13/cobra"
)

var verificationCmd = &cobra.Command{
	Use:   "verification",
	Short: "Email verification code flows",
}

var (
	verificationPurpose string
	verificationEmail   string
	verificationCode    string
)

var sendVerificationCodeCmd = &cobra.Command{
	Use:   "code",
	Short: "Send a purpose-scoped verification code",
	RunE: func(cmd *cobra.Command, args []string) error {
		if verificationPurpose == "" || verificationEmail == "" {
			return errors.New("purpose and email are required")
		}
		if err := verification_service.SendVerificationCode(verificationPurpose, verificationEmail, cliIPAddress); err != nil {
			return err
		}
		fmt.Println("Verification code sent")
		return nil
	},
}

var verifyVerificationCodeCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a code and print a one-time verification token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if verificationPurpose == "" || verificationEmail == "" || verificationCode == "" {
			return errors.New("purpose, email, and code are required")
		}
		result, err := verification_service.VerifyCode(verificationPurpose, verificationEmail, verificationCode)
		if err != nil {
			return err
		}
		fmt.Printf("Verification Token: %s\n", result.Token)
		fmt.Printf("Expires At: %s\n", result.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"))
		return nil
	},
}

func init() {
	verificationCmd.AddCommand(sendVerificationCodeCmd)
	verificationCmd.AddCommand(verifyVerificationCodeCmd)

	sendVerificationCodeCmd.Flags().StringVarP(&verificationPurpose, "purpose", "p", "", "purpose: signup, password_reset, email_change (required)")
	sendVerificationCodeCmd.Flags().StringVarP(&verificationEmail, "email", "e", "", "recipient email address (required)")
	sendVerificationCodeCmd.MarkFlagRequired("purpose")
	sendVerificationCodeCmd.MarkFlagRequired("email")

	verifyVerificationCodeCmd.Flags().StringVarP(&verificationPurpose, "purpose", "p", "", "purpose: signup, password_reset, email_change (required)")
	verifyVerificationCodeCmd.Flags().StringVarP(&verificationEmail, "email", "e", "", "recipient email address (required)")
	verifyVerificationCodeCmd.Flags().StringVarP(&verificationCode, "code", "c", "", "verification code (required)")
	verifyVerificationCodeCmd.MarkFlagRequired("purpose")
	verifyVerificationCodeCmd.MarkFlagRequired("email")
	verifyVerificationCodeCmd.MarkFlagRequired("code")
}
