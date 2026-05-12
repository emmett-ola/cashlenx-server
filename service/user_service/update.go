package user_service

import (
	std_errors "errors"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"golang.org/x/crypto/bcrypt"
)

// UpdateService updates an existing user
func UpdateService(plainId string, requestBody model.UserDTO) (model.UserEntity, error) {
	// Get existing user
	existingUser := userRepo.GetUserByObjectId(plainId)
	if existingUser.Id.IsZero() {
		return model.UserEntity{}, errors.NewNotFoundError("user not found")
	}

	// Update fields if provided
	if requestBody.Username != "" {
		// Check if username is already taken by another user
		checkUser := userRepo.GetUserByUsername(requestBody.Username)
		if !checkUser.Id.IsZero() && checkUser.Id.Hex() != plainId {
			return model.UserEntity{}, errors.NewFieldAlreadyExistsError("username", "username is already taken")
		}
		existingUser.Username = requestBody.Username
	}

	if requestBody.Role != "" {
		existingUser.Role = requestBody.Role
	}

	// Update profile fields if provided
	if requestBody.Nickname != "" {
		existingUser.Nickname = requestBody.Nickname
	}
	if requestBody.AvatarUrl != "" {
		existingUser.AvatarUrl = requestBody.AvatarUrl
	}
	if requestBody.EmailAddress != "" {
		existingUser.EmailAddress = requestBody.EmailAddress
	}
	if requestBody.Gender != "" {
		if err := validation.ValidateGender(requestBody.Gender); err != nil {
			return model.UserEntity{}, err
		}
		existingUser.Gender = requestBody.Gender
	}

	// Update password if provided
	if requestBody.Password != "" {
		err := validation.ValidatePassword(requestBody.Password)
		if err != nil {
			return model.UserEntity{}, err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
		if err != nil {
			return model.UserEntity{}, std_errors.New("failed to hash password")
		}

		existingUser.PasswordHash = string(hashedPassword)

		// Revoke all refresh tokens when password changes
		if err := revokeAllRefreshTokens(plainId); err != nil {
			util.Logger.Warnw("Failed to revoke refresh tokens after password change", "userId", plainId, "error", err)
		}
	}

	// Update updated_at timestamp
	existingUser.UpdateTime = util.GetCurrentTime()

	// Update user in database
	updatedUser := userRepo.UpdateUserByEntity(plainId, existingUser)
	if updatedUser.Id.IsZero() {
		return model.UserEntity{}, std_errors.New("failed to update user")
	}

	return updatedUser, nil
}

// SetPasswordService allows users to set or reset their password
func SetPasswordService(plainId string, password string) (model.UserEntity, error) {
	// Validate password
	err := validation.ValidatePassword(password)
	if err != nil {
		return model.UserEntity{}, err
	}

	// Get existing user
	existingUser := userRepo.GetUserByObjectId(plainId)
	if existingUser.Id.IsZero() {
		return model.UserEntity{}, errors.NewNotFoundError("user not found")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserEntity{}, std_errors.New("failed to hash password")
	}

	// Update user with new password
	existingUser.PasswordHash = string(hashedPassword)
	existingUser.UpdateTime = util.GetCurrentTime()

	// Revoke all refresh tokens when password changes
	if err := revokeAllRefreshTokens(plainId); err != nil {
		util.Logger.Warnw("Failed to revoke refresh tokens after password change", "userId", plainId, "error", err)
	}

	// Update user in database
	updatedUser := userRepo.UpdateUserByEntity(plainId, existingUser)
	if updatedUser.Id.IsZero() {
		return model.UserEntity{}, std_errors.New("failed to update password")
	}

	return updatedUser, nil
}

// UpdateProfileService updates user profile information (nickname, avatar, gender)
// Note: Email updates are handled separately via email verification
func UpdateProfileService(plainId string, requestBody model.UserProfileUpdateRequest) (model.UserEntity, error) {
	// Get existing user
	existingUser := userRepo.GetUserByObjectId(plainId)
	if existingUser.Id.IsZero() {
		return model.UserEntity{}, errors.NewNotFoundError("user not found")
	}

	// Update profile fields if provided
	// We allow updating to empty strings if that's what user wants?
	// Or only if not empty? The DTO structure usually implies optional fields.
	// Let's assume non-empty updates for now, or use pointers in struct to distinguish nil vs empty.
	// Given the struct has string types, empty string is the zero value.
	// But nickname/avatar/gender can be updated.

	if requestBody.Nickname != "" {
		existingUser.Nickname = requestBody.Nickname
	}
	if requestBody.AvatarUrl != "" {
		existingUser.AvatarUrl = requestBody.AvatarUrl
	}
	if requestBody.Gender != "" {
		if err := validation.ValidateGender(requestBody.Gender); err != nil {
			return model.UserEntity{}, err
		}
		existingUser.Gender = requestBody.Gender
	}

	// Update updated_at timestamp
	existingUser.UpdateTime = util.GetCurrentTime()
	existingUser.UpdateUserId = existingUser.Id

	// Update user in database
	updatedUser := userRepo.UpdateUserByEntity(plainId, existingUser)
	if updatedUser.Id.IsZero() {
		return model.UserEntity{}, std_errors.New("failed to update profile")
	}

	return updatedUser, nil
}

// ChangePasswordService changes the user's password after verifying the old one
func ChangePasswordService(plainId string, oldPassword, newPassword string) error {
	// Validate new password format
	if err := validation.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Get existing user
	user := userRepo.GetUserByObjectId(plainId)
	if user.Id.IsZero() {
		return errors.NewNotFoundError("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.NewUnauthorizedError("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError("failed to hash password", err)
	}

	// Update user
	user.PasswordHash = string(hashedPassword)
	user.UpdateTime = util.GetCurrentTime()
	user.UpdateUserId = user.Id

	updatedUser := userRepo.UpdateUserByEntity(plainId, user)
	if updatedUser.Id.IsZero() {
		return errors.NewInternalError("failed to update password", nil)
	}

	// Revoke all refresh tokens for security
	if err := revokeAllRefreshTokens(plainId); err != nil {
		util.Logger.Warnw("Failed to revoke tokens after password change", "userId", plainId, "error", err)
	}

	return nil
}
