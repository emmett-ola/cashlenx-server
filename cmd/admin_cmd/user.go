package admin_cmd

import (
	"errors"
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Admin user management commands that mirror /api/admin/user endpoints.`,
}

var (
	adminUserId              string
	adminUserUsername        string
	adminUserPassword        string
	adminUserNickname        string
	adminUserAvatarURL       string
	adminUserEmail           string
	adminUserGender          string
	adminUserEmailVerified   bool
	adminUserEmailVerifiedOn bool
	adminUserLimit           int
	adminUserOffset          int
)

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a normal user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adminUserUsername == "" || adminUserPassword == "" {
			return errors.New("username and password are required")
		}
		request := model.UserDTO{
			Username:        adminUserUsername,
			Password:        adminUserPassword,
			Nickname:        adminUserNickname,
			AvatarUrl:       adminUserAvatarURL,
			EmailAddress:    adminUserEmail,
			Gender:          adminUserGender,
			IsEmailVerified: adminUserEmailVerified,
		}
		createdId, err := createAdminUser(request, &adminSessionUserId)
		if err != nil {
			return err
		}
		createdUser := getAdminUser(createdId)
		fmt.Println("User created successfully")
		printAdminUser(createdUser)
		return nil
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	RunE: func(cmd *cobra.Command, args []string) error {
		users := listAdminUsers(adminUserLimit, adminUserOffset)
		total := countAdminUsers()
		fmt.Printf("Total Users: %d\n\n", total)
		if len(users) == 0 {
			fmt.Println("No users found")
			return nil
		}
		fmt.Println("ID                       | Username            | Role  | Active | Email")
		fmt.Println("-------------------------|---------------------|-------|--------|--------------------------")
		for _, user := range users {
			fmt.Printf("%-24s | %-19s | %-5s | %-6t | %s\n",
				user.Id.Hex(), truncateUserField(user.Username, 19), user.Role, user.IsActive, user.EmailAddress)
		}
		return nil
	},
}

var userGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a user by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adminUserId == "" {
			return errors.New("id is required")
		}
		user := getAdminUser(adminUserId)
		if user.Id.IsZero() {
			return errors.New("user not found")
		}
		printAdminUser(user)
		return nil
	},
}

var userUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adminUserId == "" {
			return errors.New("id is required")
		}
		request := model.UserDTO{
			Username:     adminUserUsername,
			Password:     adminUserPassword,
			Nickname:     adminUserNickname,
			AvatarUrl:    adminUserAvatarURL,
			EmailAddress: adminUserEmail,
			Gender:       adminUserGender,
		}
		if adminUserEmailVerifiedOn {
			request.IsEmailVerified = adminUserEmailVerified
		}
		updatedUser, err := updateAdminUser(adminUserId, request)
		if err != nil {
			return err
		}
		fmt.Println("User updated successfully")
		printAdminUser(updatedUser)
		return nil
	},
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if adminUserId == "" {
			return errors.New("id is required")
		}
		if err := deleteAdminUser(adminUserId); err != nil {
			return err
		}
		fmt.Println("User deleted successfully")
		return nil
	},
}

func printAdminUser(user model.UserEntity) {
	fmt.Printf("ID:              %s\n", user.Id.Hex())
	fmt.Printf("Username:        %s\n", user.Username)
	fmt.Printf("Role:            %s\n", user.Role)
	fmt.Printf("Active:          %t\n", user.IsActive)
	fmt.Printf("Nickname:        %s\n", user.Nickname)
	fmt.Printf("Email:           %s\n", user.EmailAddress)
	fmt.Printf("Email Verified:  %t\n", user.IsEmailVerified)
	fmt.Printf("Gender:          %s\n", user.Gender)
}

func truncateUserField(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func init() {
	AdminCmd.AddCommand(userCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userGetCmd)
	userCmd.AddCommand(userUpdateCmd)
	userCmd.AddCommand(userDeleteCmd)

	userCreateCmd.Flags().StringVar(&adminUserUsername, "username", "", "username (required)")
	userCreateCmd.Flags().StringVar(&adminUserPassword, "password", "", "password (required)")
	userCreateCmd.Flags().StringVar(&adminUserNickname, "nickname", "", "nickname")
	userCreateCmd.Flags().StringVar(&adminUserAvatarURL, "avatar-url", "", "avatar URL")
	userCreateCmd.Flags().StringVar(&adminUserEmail, "email", "", "email address")
	userCreateCmd.Flags().StringVar(&adminUserGender, "gender", "", "gender: male, female, others")
	userCreateCmd.Flags().BoolVar(&adminUserEmailVerified, "email-verified", false, "mark email as verified")
	userCreateCmd.MarkFlagRequired("username")
	userCreateCmd.MarkFlagRequired("password")

	userListCmd.Flags().IntVar(&adminUserLimit, "limit", 20, "result limit")
	userListCmd.Flags().IntVar(&adminUserOffset, "offset", 0, "result offset")

	userGetCmd.Flags().StringVar(&adminUserId, "id", "", "user ID (required)")
	userGetCmd.MarkFlagRequired("id")

	userUpdateCmd.Flags().StringVar(&adminUserId, "id", "", "user ID (required)")
	userUpdateCmd.Flags().StringVar(&adminUserUsername, "username", "", "username")
	userUpdateCmd.Flags().StringVar(&adminUserPassword, "password", "", "password")
	userUpdateCmd.Flags().StringVar(&adminUserNickname, "nickname", "", "nickname")
	userUpdateCmd.Flags().StringVar(&adminUserAvatarURL, "avatar-url", "", "avatar URL")
	userUpdateCmd.Flags().StringVar(&adminUserEmail, "email", "", "email address")
	userUpdateCmd.Flags().StringVar(&adminUserGender, "gender", "", "gender: male, female, others")
	userUpdateCmd.Flags().BoolVar(&adminUserEmailVerified, "email-verified", false, "mark email as verified")
	userUpdateCmd.Flags().BoolVar(&adminUserEmailVerifiedOn, "set-email-verified", false, "apply email-verified flag")
	userUpdateCmd.MarkFlagRequired("id")

	userDeleteCmd.Flags().StringVar(&adminUserId, "id", "", "user ID (required)")
	userDeleteCmd.MarkFlagRequired("id")
}
