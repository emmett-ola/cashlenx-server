package user_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var INSTANCE UserMapper

// UserMapper interface defines operations for managing users
type UserMapper interface {
	// GetUserByObjectId retrieves a user by their ID
	GetUserByObjectId(plainId string) model.UserEntity
	
	// GetUserByUsername retrieves a user by their username
	GetUserByUsername(username string) model.UserEntity
	
	// InsertUserByEntity inserts a new user entity into the database
	InsertUserByEntity(newEntity model.UserEntity) string
	
	// UpdateUserByEntity updates an existing user entity
	UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity
	
	// GetAllUsers retrieves all users with pagination
	GetAllUsers(limit, offset int) []model.UserEntity
	
	// GetUsersByRole retrieves all users with a specific role
	GetUsersByRole(role string) []model.UserEntity
	
	// CountAllUsers returns the total number of users
	CountAllUsers() int64
	
	// DeleteUserByObjectId deletes a user by their ID
	DeleteUserByObjectId(plainId string) model.UserEntity
	
	// TruncateUsers deletes all users from the database
	TruncateUsers() error
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = UserMongoDbMapper{}
	case "mysql":
		INSTANCE = UserMySqlMapper{}
	default:
		panic("database type not supported")
	}
}