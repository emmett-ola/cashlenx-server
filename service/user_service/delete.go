package user_service

import (
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/refresh_token_service"
	"github.com/macar-x/cashlenx-server/util"
)

// DeleteService deletes a user by ID
func DeleteService(userId string) error {
	// Check if user exists
	existingUser := user_mapper.INSTANCE.GetUserByObjectId(userId)
	if existingUser.Id.IsZero() {
		return errors.NewNotFoundError("user not found")
	}

	// Prevent deletion of admin users
	if existingUser.Role == model.UserRoleAdmin {
		return errors.NewForbiddenError("admin users cannot be deleted")
	}

	// Delete the user (Soft Delete)
	deletedUser := user_mapper.INSTANCE.DeleteUserByObjectId(userId)
	if deletedUser.Id.IsZero() {
		return errors.NewInternalError("failed to delete user", nil)
	}

	// Invalidate tokens
	if err := refresh_token_service.RevokeAllRefreshTokens(userId); err != nil {
		util.Logger.Warnw("Failed to revoke tokens after account deletion", "userId", userId, "error", err)
	}

	return nil
}
