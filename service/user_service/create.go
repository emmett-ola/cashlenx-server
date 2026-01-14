package user_service

import (
	std_errors "errors"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// CreateService creates a new user with the provided details
func CreateService(requestBody model.UserDTO) (string, error) {
	// Validate password for non-external users
	if !requestBody.IsExternal {
		err := validation.ValidatePassword(requestBody.Password)
		if err != nil {
			return "", err
		}
	}

	// Check if username is already taken
	existingUser := user_mapper.INSTANCE.GetUserByUsername(requestBody.Username)
	if !existingUser.Id.IsZero() {
		return "", errors.NewFieldAlreadyExistsError("username", "username is already taken")
	}

	// Hash the password if provided and not external
	var hashedPassword string
	var passwordSet bool
	if requestBody.Password != "" {
		passwordBytes, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", std_errors.New("failed to hash password")
		}
		hashedPassword = string(passwordBytes)
		passwordSet = true
	}

	// Create the user entity
	userEntity := model.UserEntity{
		Id:           primitive.NewObjectID(),
		Username:     requestBody.Username,
		PasswordHash: hashedPassword,
		IsActive:     true,
		Role:         model.UserRoleUser, // Always create as normal user
		IsExternal:   requestBody.IsExternal,
		ExternalId:   requestBody.ExternalId,
		PasswordSet:  passwordSet,
		CreatedAt:    util.GetCurrentTime(),
		UpdatedAt:    util.GetCurrentTime(),
	}

	// Insert the user into the database
	userId := user_mapper.INSTANCE.InsertUserByEntity(userEntity)
	if userId == "" {
		return "", std_errors.New("failed to create user")
	}

	return userId, nil
}

// CreateExternalUser creates a new user from an external authentication system
func CreateExternalUser(username, externalId string) (model.UserEntity, error) {
	// Check if username is already taken
	existingUser := user_mapper.INSTANCE.GetUserByUsername(username)
	if !existingUser.Id.IsZero() {
		return existingUser, nil
	}

	// Create the user entity with external flags
	userEntity := model.UserEntity{
		Id:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: "", // No password initially for external users
		IsActive:     true,
		Role:         model.UserRoleUser,
		IsExternal:   true,
		ExternalId:   externalId,
		PasswordSet:  false,
		CreatedAt:    util.GetCurrentTime(),
		UpdatedAt:    util.GetCurrentTime(),
	}

	// Insert the user into the database
	userId := user_mapper.INSTANCE.InsertUserByEntity(userEntity)
	if userId == "" {
		return model.UserEntity{}, std_errors.New("failed to create external user")
	}

	// Get the created user
	createdUser := user_mapper.INSTANCE.GetUserByObjectId(userId)
	if createdUser.Id.IsZero() {
		return model.UserEntity{}, std_errors.New("failed to retrieve created external user")
	}

	return createdUser, nil
}
