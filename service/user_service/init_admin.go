package user_service

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// InitAdminUser initializes the admin user if no admin users exist
func InitAdminUser() {
	// Check if any admin users exist
	adminUsers := userRepo.GetUsersByRole(model.UserRoleAdmin)
	if len(adminUsers) > 0 {
		util.Logger.Info("Admin user already exists, skipping initialization")
		return
	}

	// Get admin credentials from environment variables
	adminUsername := util.GetConfigByKey("admin.username")
	adminPassword := util.GetConfigByKey("admin.password")

	// Set default admin credentials if not provided
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin"
	}

	// Check if the admin username is already taken by a non-admin user
	existingUser := userRepo.GetUserByUsername(adminUsername)
	if !existingUser.Id.IsZero() {
		util.Logger.Warnf("Username %s is already taken by a non-admin user, skipping admin initialization", adminUsername)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		util.Logger.Errorw("Failed to hash admin password", "error", err)
		return
	}

	// Create the admin user entity
	adminUser := model.UserEntity{
		Id:           primitive.NewObjectID(),
		Username:     adminUsername,
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		Role:         model.UserRoleAdmin,
		BaseEntity: model.BaseEntity{
			CreateTime: util.GetCurrentTime(),
			UpdateTime: util.GetCurrentTime(),
		},
	}
	// Set self as creator
	adminUser.CreateUserId = adminUser.Id
	adminUser.UpdateUserId = adminUser.Id

	// Insert the admin user into the database
	userId := userRepo.InsertUserByEntity(adminUser)
	if userId == "" {
		util.Logger.Error("Failed to create admin user")
		return
	}

	util.Logger.Infof("Admin user %s created successfully", adminUsername)

	// Initialize default categories for the admin user
	if err := initializeDefaultCategoriesForUser(userId); err != nil {
		util.Logger.Warnw("Failed to initialize default categories for admin user",
			"userId", userId,
			"error", err)
		// Don't fail admin user creation if category initialization fails
	}
}
