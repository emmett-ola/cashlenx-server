package open_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/cmd/cli_auth"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication and password reset flows",
}

var (
	loginUsername     string
	loginPassword     string
	loginRefreshToken string

	registerUsername          string
	registerPassword          string
	registerEmail             string
	registerVerificationToken string

	logoutRefreshToken string
	logoutAccessToken  string

	resetEmailOrUsername string
	resetToken           string
	resetPassword        string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with username/password or refresh token",
	RunE: func(cmd *cobra.Command, args []string) error {
		var accessToken, refreshToken string
		var user model.UserEntity
		var err error
		if loginRefreshToken != "" {
			accessToken, refreshToken, user, err = refreshOpenToken(loginRefreshToken, cli_auth.DeviceID, cli_auth.DeviceName, cli_auth.IPAddress, cli_auth.UserAgent)
		} else {
			if loginUsername == "" || loginPassword == "" {
				return errors.New("username and password are required unless --refresh-token is provided")
			}
			accessToken, refreshToken, user, err = authenticateOpenUser(loginUsername, loginPassword, cli_auth.DeviceID, cli_auth.DeviceName, cli_auth.IPAddress, cli_auth.UserAgent)
		}
		if err != nil {
			return err
		}
		if err := saveOpenSession(accessToken, refreshToken, user); err != nil {
			return err
		}

		fmt.Printf("User: %s (%s)\n", user.Username, user.Id.Hex())
		fmt.Printf("Role: %s\n", user.Role)
		fmt.Printf("Access Token: %s\n", accessToken)
		fmt.Printf("Refresh Token: %s\n", refreshToken)
		fmt.Println("CLI session saved")
		return nil
	},
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if registerUsername == "" || registerPassword == "" || registerEmail == "" || registerVerificationToken == "" {
			return errors.New("username, password, email, and verification token are required")
		}
		userId, err := registerOpenUser(registerUsername, registerPassword, registerEmail, registerVerificationToken)
		if err != nil {
			return err
		}
		fmt.Printf("User registered successfully: %s (%s)\n", registerUsername, userId)
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout by revoking refresh token or all tokens for an access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if logoutRefreshToken == "" && logoutAccessToken == "" {
			session, err := currentOpenSession()
			if err == nil {
				logoutRefreshToken = session.RefreshToken
			}
		}
		if logoutRefreshToken != "" {
			refreshToken, err := getOpenRefreshToken(logoutRefreshToken, cli_auth.DeviceID, cli_auth.DeviceName, cli_auth.IPAddress, cli_auth.UserAgent)
			if err != nil {
				return err
			}
			if err := revokeOpenRefreshToken(logoutRefreshToken, refreshToken.UserId); err != nil {
				return err
			}
			_ = clearOpenSession()
			fmt.Println("Successfully logged out from this device")
			return nil
		}
		if logoutAccessToken != "" {
			claims, err := validateOpenAccessToken(logoutAccessToken)
			if err != nil {
				return err
			}
			if err := revokeAllOpenRefreshTokens(claims.UserID); err != nil {
				return err
			}
			_ = clearOpenSession()
			fmt.Println("Successfully logged out from all devices")
			return nil
		}
		_ = clearOpenSession()
		fmt.Println("Logout accepted")
		return nil
	},
}

var resetPasswordCmd = &cobra.Command{
	Use:   "reset-password",
	Short: "Request a password reset verification code",
	RunE: func(cmd *cobra.Command, args []string) error {
		if resetEmailOrUsername == "" {
			return errors.New("email or username is required")
		}
		if err := requestOpenPasswordReset(resetEmailOrUsername, cli_auth.IPAddress); err != nil {
			return err
		}
		fmt.Println("Password reset request accepted")
		return nil
	},
}

var resetPasswordConfirmCmd = &cobra.Command{
	Use:   "reset-password-confirm",
	Short: "Confirm password reset with a verification token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if resetToken == "" || resetPassword == "" {
			return errors.New("token and password are required")
		}
		if err := confirmOpenPasswordReset(resetToken, resetPassword); err != nil {
			return err
		}
		fmt.Println("Password reset successfully")
		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(registerCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(resetPasswordCmd)
	authCmd.AddCommand(resetPasswordConfirmCmd)

	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "password")
	loginCmd.Flags().StringVar(&loginRefreshToken, "refresh-token", "", "refresh token")

	registerCmd.Flags().StringVarP(&registerUsername, "username", "u", "", "username (required)")
	registerCmd.Flags().StringVarP(&registerPassword, "password", "p", "", "password (required)")
	registerCmd.Flags().StringVarP(&registerEmail, "email", "e", "", "email address (required)")
	registerCmd.Flags().StringVarP(&registerVerificationToken, "verification-token", "t", "", "verification token from open verification verify (required)")

	logoutCmd.Flags().StringVar(&logoutRefreshToken, "refresh-token", "", "refresh token to revoke")
	logoutCmd.Flags().StringVar(&logoutAccessToken, "access-token", "", "access token whose user's refresh tokens should be revoked")

	resetPasswordCmd.Flags().StringVarP(&resetEmailOrUsername, "email-or-username", "e", "", "email address or username (required)")
	resetPasswordCmd.MarkFlagRequired("email-or-username")

	resetPasswordConfirmCmd.Flags().StringVarP(&resetToken, "token", "t", "", "verification token (required)")
	resetPasswordConfirmCmd.Flags().StringVarP(&resetPassword, "password", "p", "", "new password (required)")
	resetPasswordConfirmCmd.MarkFlagRequired("token")
	resetPasswordConfirmCmd.MarkFlagRequired("password")
}
