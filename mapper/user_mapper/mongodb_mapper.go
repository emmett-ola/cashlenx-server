package user_mapper

import (
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// conversion functions
func convertBsonM2UserEntity(bsonData bson.M) model.UserEntity {
	if bsonData == nil {
		return model.UserEntity{}
	}

	user := model.UserEntity{}

	// Convert ID
	if id, ok := bsonData["_id"].(primitive.ObjectID); ok {
		user.Id = id
	}

	// Convert string fields
	if username, ok := bsonData["username"].(string); ok {
		user.Username = username
	}
	if passwordHash, ok := bsonData["password_hash"].(string); ok {
		user.PasswordHash = passwordHash
	}
	if role, ok := bsonData["role"].(string); ok {
		user.Role = role
	}

	// Convert boolean fields
	if isActive, ok := bsonData["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	// Convert time fields
	if createdAt, ok := bsonData["created_at"].(time.Time); ok {
		user.CreatedAt = createdAt
	}
	if updatedAt, ok := bsonData["updated_at"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}

	return user
}

func convertUserEntity2BsonD(user model.UserEntity) bson.D {
	return bson.D{
		primitive.E{Key: "_id", Value: user.Id},
		primitive.E{Key: "username", Value: user.Username},
		primitive.E{Key: "password_hash", Value: user.PasswordHash},
		primitive.E{Key: "is_active", Value: user.IsActive},
		primitive.E{Key: "role", Value: user.Role},
		primitive.E{Key: "created_at", Value: user.CreatedAt},
		primitive.E{Key: "updated_at", Value: user.UpdatedAt},
	}
}

// UserMongoDbMapper MongoDB implementation of UserMapper

type UserMongoDbMapper struct{}

// GetUserByObjectId retrieves a user by their ID from MongoDB
func (m UserMongoDbMapper) GetUserByObjectId(plainId string) model.UserEntity {
	// Parse the plain ID to ObjectID
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("user's id is not acceptable")
		return model.UserEntity{}
	}

	// Create a filter to find the user by ID
	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2UserEntity(database.GetOneInMongoDB(filter))
}

// GetUserByUsername retrieves a user by their username from MongoDB
func (m UserMongoDbMapper) GetUserByUsername(username string) model.UserEntity {
	// Create a filter to find the user by username
	filter := bson.D{
		primitive.E{Key: "username", Value: username},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2UserEntity(database.GetOneInMongoDB(filter))
}

// InsertUserByEntity inserts a new user entity into MongoDB
func (m UserMongoDbMapper) InsertUserByEntity(newEntity model.UserEntity) string {
	// Set default values if not provided
	if newEntity.CreatedAt.IsZero() {
		newEntity.CreatedAt = time.Now()
	}
	if newEntity.UpdatedAt.IsZero() {
		newEntity.UpdatedAt = time.Now()
	}
	if newEntity.IsActive == false {
		newEntity.IsActive = true
	}
	if newEntity.Role == "" {
		newEntity.Role = model.UserRoleUser
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()
	newUserId := database.InsertOneInMongoDB(convertUserEntity2BsonD(newEntity))
	return newUserId.Hex()
}

// UpdateUserByEntity updates an existing user entity in MongoDB
func (m UserMongoDbMapper) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	// Parse the plain ID to ObjectID
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("user's id is not acceptable")
		return model.UserEntity{}
	}

	// Create a filter to find the user by ID
	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	// Set the updated timestamp
	updatedEntity.UpdatedAt = time.Now()

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Update the user document in the database
	database.UpdateManyInMongoDB(filter, convertUserEntity2BsonD(updatedEntity))

	// Return the updated user
	return m.GetUserByObjectId(plainId)
}

// GetAllUsers retrieves all users with pagination from MongoDB
func (m UserMongoDbMapper) GetAllUsers(limit, offset int) []model.UserEntity {
	// Create an empty filter to get all users
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Get the user documents from the database
	var users []model.UserEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		users = append(users, convertBsonM2UserEntity(queryResult))
	}

	return users
}

// CountAllUsers returns the total number of users from MongoDB
func (m UserMongoDbMapper) CountAllUsers() int64 {
	// Create an empty filter to count all users
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	return database.CountInMongoDB(filter)
}

// DeleteUserByObjectId deletes a user by their ID from MongoDB
func (m UserMongoDbMapper) DeleteUserByObjectId(plainId string) model.UserEntity {
	// Parse the plain ID to ObjectID
	objectId := util.Convert2ObjectId(plainId)
	if plainId == "" || objectId == primitive.NilObjectID {
		util.Logger.Warnln("user's id is not acceptable")
		return model.UserEntity{}
	}

	// Get the user first to return it
	user := m.GetUserByObjectId(plainId)
	if user.IsEmpty() {
		return model.UserEntity{}
	}

	// Create a filter to find the user by ID
	filter := bson.D{
		primitive.E{Key: "_id", Value: objectId},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Delete the user document from the database
	database.DeleteManyInMongoDB(filter)

	return user
}

// TruncateUsers deletes all users from MongoDB
func (m UserMongoDbMapper) TruncateUsers() error {
	// Create an empty filter to delete all users
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Delete all user documents from the database
	deletedCount := database.DeleteManyInMongoDB(filter)
	util.Logger.Infow("Users truncated successfully", "deleted_count", deletedCount)
	return nil
}

// GetUsersByRole retrieves all users with a specific role from MongoDB
func (m UserMongoDbMapper) GetUsersByRole(role string) []model.UserEntity {
	// Create a filter to find users by role
	filter := bson.D{
		primitive.E{Key: "role", Value: role},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Get the user documents from the database
	var users []model.UserEntity
	queryResultList := database.GetManyInMongoDB(filter)
	for _, queryResult := range queryResultList {
		users = append(users, convertBsonM2UserEntity(queryResult))
	}

	return users
}