package user_service

import (
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
)

// GetUserByObjectId retrieves a user by their ID
func GetUserByObjectId(userId string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByObjectId(userId)
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(username string) model.UserEntity {
	return user_mapper.INSTANCE.GetUserByUsername(username)
}

// GetAllUsers retrieves all users with pagination
func GetAllUsers(limit, offset int) []model.UserEntity {
	return user_mapper.INSTANCE.GetAllUsers(limit, offset)
}

// CountAllUsers returns the total number of users
func CountAllUsers() int64 {
	return user_mapper.INSTANCE.CountAllUsers()
}