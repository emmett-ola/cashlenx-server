package user_service

import (
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
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

	// Delete the user (Soft Delete)
	deletedUser := userRepo.DeleteUserByObjectId(userId)
	if deletedUser.Id.IsZero() {
		return errors.NewInternalError("failed to delete user", nil)
	}

	// Invalidate tokens
	if err := revokeAllRefreshTokens(userId); err != nil {
		util.Logger.Warnw("Failed to revoke tokens after account deletion", "userId", userId, "error", err)
	}

	return nil
}
