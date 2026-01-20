package user_service

import (
	std_errors "errors"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// CreateService creates a new user with the provided details
func CreateService(requestBody model.UserDTO) (string, error) {
	// Validate password
	err := validation.ValidatePassword(requestBody.Password)
	if err != nil {
		return "", err
	}

	// Check if username is already taken
	existingUser := user_mapper.INSTANCE.GetUserByUsername(requestBody.Username)
	if !existingUser.Id.IsZero() {
		return "", errors.NewFieldAlreadyExistsError("username", "username is already taken")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestBody.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", std_errors.New("failed to hash password")
	}

	// Create the user entity
	userEntity := model.UserEntity{
		Id:           primitive.NewObjectID(),
		Username:     requestBody.Username,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		Role:         model.UserRoleUser, // Always create as normal user
		Nickname:     requestBody.Nickname,
		AvatarUrl:    requestBody.AvatarUrl,
		EmailAddress: requestBody.EmailAddress,
		Gender:       requestBody.Gender,
		BaseEntity: model.BaseEntity{
			CreateTime: util.GetCurrentTime(),
			UpdateTime: util.GetCurrentTime(),
		},
	}

	// Insert the user into the database
	userId := user_mapper.INSTANCE.InsertUserByEntity(userEntity)
	if userId == "" {
		return "", std_errors.New("failed to create user")
	}

	// Initialize default categories for the new user
	if err := category_service.InitializeDefaultCategoriesForUser(userId); err != nil {
		util.Logger.Warnw("Failed to initialize default categories for new user",
			"userId", userId,
			"error", err)
		// Don't fail user creation if category initialization fails
	}

	return userId, nil
}
