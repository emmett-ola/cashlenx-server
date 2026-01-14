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
func UpdateService(userId string, requestBody model.UserDTO) error {
	// Check if user exists
	existingUser := user_mapper.INSTANCE.GetUserByObjectId(userId)
	if existingUser.Id.IsZero() {
		return std_errors.New("user not found")
	}

	// Check if username is already taken by another user
	if requestBody.Username != "" && requestBody.Username != existingUser.Username {
		userWithSameName := user_mapper.INSTANCE.GetUserByUsername(requestBody.Username)
		if !userWithSameName.Id.IsZero() && userWithSameName.Id.Hex() != userId {
			return errors.NewFieldAlreadyExistsError("username", "username is already taken")
		}
		existingUser.Username = requestBody.Username
	}

	// Update password if provided
	if requestBody.Password != "" {
		// Validate password
		err := validation.ValidatePassword(requestBody.Password)
		if err != nil {
			return err
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
		if err != nil {
			return std_errors.New("failed to hash password")
		}
		existingUser.PasswordHash = string(hashedPassword)
	}

	// Update other fields if provided
	if requestBody.Role != "" {
		existingUser.Role = requestBody.Role
	}

	// Update active status if provided
	existingUser.IsActive = requestBody.IsActive

	// Update timestamp
	existingUser.UpdatedAt = util.GetCurrentTime()

	// Update the user in the database
	updatedUser := user_mapper.INSTANCE.UpdateUserByEntity(userId, existingUser)
	if updatedUser.Id.IsZero() {
		return std_errors.New("failed to update user")
	}

	return nil
}
