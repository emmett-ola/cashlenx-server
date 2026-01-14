package user_service

import (
	"errors"

	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
)

// DeleteService deletes a user by ID
func DeleteService(userId string) error {
	// Check if user exists
	existingUser := user_mapper.INSTANCE.GetUserByObjectId(userId)
	if existingUser.Id.IsZero() {
		return errors.New("user not found")
	}

	// Delete the user
	deletedUser := user_mapper.INSTANCE.DeleteUserByObjectId(userId)
	if deletedUser.Id.IsZero() {
		return errors.New("failed to delete user")
	}

	return nil
}
