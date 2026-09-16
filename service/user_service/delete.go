package user_service

import (
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
)

// DeleteService deletes a user by ID
func DeleteService(userId string) error {
	// Check if user exists
	existingUser := userRepo.GetUserByObjectId(userId)
	if existingUser.Id.IsZero() {
		return errors.NewNotFoundError("user not found")
	}

	// Prevent deletion of admin users
	if existingUser.Role == model.UserRoleAdmin {
		return errors.NewForbiddenError("admin users cannot be deleted")
	}

	// Revoke sessions before deleting the account. A revocation failure leaves
	// the account unchanged rather than reporting deletion while sessions live.
	if err := revokeAllRefreshTokens(userId); err != nil {
		return errors.NewInternalError("failed to revoke sessions before account deletion", err)
	}

	// Delete the user (Soft Delete)
	deletedUser := userRepo.DeleteUserByObjectId(userId)
	if deletedUser.Id.IsZero() {
		return errors.NewInternalError("failed to delete user", nil)
	}

	return nil
}
