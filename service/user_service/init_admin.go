package user_service

import (
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// InitAdminUser initializes the admin user if no admin users exist
func InitAdminUser() {
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

	// Check if any admin users exist
	adminUsers := user_mapper.INSTANCE.GetUsersByRole(model.UserRoleAdmin)
	if len(adminUsers) > 0 {
		// Admin user(s) exist, check if we need to update the password
		for _, adminUser := range adminUsers {
			// Check if the admin user's username matches the configured admin username
			if adminUser.Username == adminUsername {
				// Check if the password is still the default "admin" or doesn't match the configured password
				// We'll update it if it's the default or if the configured password is different
				err := bcrypt.CompareHashAndPassword([]byte(adminUser.PasswordHash), []byte("admin"))
				if err == nil || adminPassword != "admin" {
					// Password is either default or has been configured, update it
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
					if err != nil {
						util.Logger.Errorw("Failed to hash admin password", "error", err)
						continue
					}

					// Update the password
					adminUser.PasswordHash = string(hashedPassword)
					adminUser.UpdateTime = util.GetCurrentTime()

					// Save the updated user
					updatedUser := user_mapper.INSTANCE.UpdateUserByEntity(adminUser.Id.Hex(), adminUser)
					if !updatedUser.Id.IsZero() {
						util.Logger.Infof("Updated admin user %s password", adminUsername)
					} else {
						util.Logger.Errorw("Failed to update admin user password", "username", adminUsername)
					}
				}
			}
		}
		util.Logger.Info("Admin user already exists, skipping initialization")
		return
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
		BaseEntity: model.BaseEntity{
			CreateTime: util.GetCurrentTime(),
			UpdateTime: util.GetCurrentTime(),
		},
	}
	// Set self as creator
	adminUser.CreateUserId = adminUser.Id
	adminUser.UpdateUserId = adminUser.Id

	// Insert the admin user into the database
	userId := user_mapper.INSTANCE.InsertUserByEntity(adminUser)
	if userId == "" {
		util.Logger.Error("Failed to create admin user")
		return
	}

	util.Logger.Infof("Admin user %s created successfully", adminUsername)

	// Initialize default categories for the admin user
	if err := category_service.InitializeDefaultCategoriesForUser(userId); err != nil {
		util.Logger.Warnw("Failed to initialize default categories for admin user",
			"userId", userId,
			"error", err)
		// Don't fail admin user creation if category initialization fails
	}
}
