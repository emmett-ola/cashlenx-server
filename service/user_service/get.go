package user_service

import (
	"errors"
	"strings"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

// GetDefaultAdminUserId retrieves the default admin user ID
func GetDefaultAdminUserId() (string, error) {
	adminUsername := util.GetConfigByKey("admin.username")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	user := GetUserByUsername(adminUsername)
	if user.Id.IsZero() {
		return "", errors.New("default admin user not found")
	}
	return user.Id.Hex(), nil
}

// GetService retrieves a user by their ID (Alias for GetUserByObjectId)
func GetService(userId string) model.UserEntity {
	return GetUserByObjectId(userId)
}

// GetUser retrieves a user by their ID (Alias for GetUserByObjectId)
func GetUser(userId string) model.UserEntity {
	return GetUserByObjectId(userId)
}

// GetUserByObjectId retrieves a user by their ID
func GetUserByObjectId(userId string) model.UserEntity {
	return userRepo.GetUserByObjectId(userId)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) model.UserEntity {
	return userRepo.GetUserByUsername(username)
}

// GetUserByLoginIdentifier retrieves a user by username or normalized email.
// The login request retains its historical "username" field for wire
// compatibility while accepting either identifier.
func GetUserByLoginIdentifier(identifier string) model.UserEntity {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		return userRepo.GetUserByEmail(strings.ToLower(identifier))
	}
	return userRepo.GetUserByUsername(identifier)
}

// GetAllUsers retrieves all users with pagination
func GetAllUsers(limit, offset int) []model.UserEntity {
	return userRepo.GetAllUsers(limit, offset)
}

// CountAllUsers returns the total number of users
func CountAllUsers() int64 {
	return userRepo.CountAllUsers()
}
