package user_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
)

func printUser(user model.UserEntity) {
	fmt.Printf("ID:              %s\n", user.Id.Hex())
	fmt.Printf("Username:        %s\n", user.Username)
	fmt.Printf("Role:            %s\n", user.Role)
	fmt.Printf("Active:          %t\n", user.IsActive)
	fmt.Printf("Nickname:        %s\n", user.Nickname)
	fmt.Printf("Email:           %s\n", user.EmailAddress)
	fmt.Printf("Email Verified:  %t\n", user.IsEmailVerified)
	fmt.Printf("Gender:          %s\n", user.Gender)
	fmt.Printf("Phone:           %s\n", user.PhoneNumber)
	fmt.Printf("Location:        %s\n", user.Location)
	fmt.Printf("Birth Date:      %s\n", user.BirthDate)
	fmt.Printf("Created At:      %s\n", user.CreateTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated At:      %s\n", user.UpdateTime.Format("2006-01-02 15:04:05"))
}
