package user_service

import (
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// InitAdminUser initializes the admin user if no admin users exist
func InitAdminUser() {
	// Check if any admin users exist
	adminUsers := user_mapper.INSTANCE.GetUsersByRole(model.UserRoleAdmin)
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
	existingUser := user_mapper.INSTANCE.GetUserByUsername(adminUsername)
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
		CreatedAt:    util.GetCurrentTime(),
		UpdatedAt:    util.GetCurrentTime(),
	}

	// Insert the admin user into the database
	userId := user_mapper.INSTANCE.InsertUserByEntity(adminUser)
	if userId == "" {
		util.Logger.Error("Failed to create admin user")
		return
	}

	util.Logger.Infof("Admin user %s created successfully", adminUsername)
}
