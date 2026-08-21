package user_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/spf13/cobra"
)

var (
	profileNickname string
	profileAvatar   string
	profileGender   string
	profilePhone    string
	profileLocation string
	profileBirth    string
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get or update the current user's profile",
}

var profileGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		user := getUserProfile(userId)
		if user.Id.IsZero() {
			return errors.NewNotFoundError("user not found")
		}
		printUser(user)
		return nil
	},
}

var profileUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		updatedUser, err := updateUserProfile(userId, model.UserProfileUpdateRequest{
			Nickname:    profileNickname,
			AvatarUrl:   profileAvatar,
			Gender:      profileGender,
			PhoneNumber: profilePhone,
			Location:    profileLocation,
			BirthDate:   profileBirth,
		})
		if err != nil {
			return err
		}
		fmt.Println("Profile updated successfully")
		printUser(updatedUser)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileGetCmd)
	profileCmd.AddCommand(profileUpdateCmd)

	profileUpdateCmd.Flags().StringVar(&profileNickname, "nickname", "", "nickname")
	profileUpdateCmd.Flags().StringVar(&profileAvatar, "avatar-url", "", "avatar URL")
	profileUpdateCmd.Flags().StringVar(&profileGender, "gender", "", "gender: male, female, others")
	profileUpdateCmd.Flags().StringVar(&profilePhone, "phone", "", "phone number")
	profileUpdateCmd.Flags().StringVar(&profileLocation, "location", "", "location")
	profileUpdateCmd.Flags().StringVar(&profileBirth, "birth-date", "", "birth date in YYYY-MM-DD")
}
