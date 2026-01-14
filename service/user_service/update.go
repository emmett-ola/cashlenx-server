package user_service

import (
	std_errors "errors"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"golang.org/x/crypto/bcrypt"
)

// UpdateService updates an existing user
func UpdateService(plainId string, requestBody model.UserDTO) (model.UserEntity, error) {
	// Get existing user
	existingUser := user_mapper.INSTANCE.GetUserByObjectId(plainId)
	if existingUser.Id.IsZero() {
		return model.UserEntity{}, errors.NewNotFoundError("user not found")
	}

	// Update fields if provided
	if requestBody.Username != "" {
		// Check if username is already taken by another user
		checkUser := user_mapper.INSTANCE.GetUserByUsername(requestBody.Username)
		if !checkUser.Id.IsZero() && checkUser.Id.Hex() != plainId {
			return model.UserEntity{}, errors.NewFieldAlreadyExistsError("username", "username is already taken")
		}
		existingUser.Username = requestBody.Username
	}

	if requestBody.Role != "" {
		existingUser.Role = requestBody.Role
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
		existingUser.PasswordSet = true
	}

	// Update updated_at timestamp
	existingUser.UpdatedAt = util.GetCurrentTime()

	// Update user in database
	updatedUser := user_mapper.INSTANCE.UpdateUserByEntity(plainId, existingUser)
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
	existingUser := user_mapper.INSTANCE.GetUserByObjectId(plainId)
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
	existingUser.PasswordSet = true
	existingUser.UpdatedAt = util.GetCurrentTime()

	// Update user in database
	updatedUser := user_mapper.INSTANCE.UpdateUserByEntity(plainId, existingUser)
	if updatedUser.Id.IsZero() {
		return model.UserEntity{}, std_errors.New("failed to update password")
	}

	return updatedUser, nil
}
