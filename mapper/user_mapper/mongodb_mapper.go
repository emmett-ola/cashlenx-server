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

	// Convert profile fields
	if nickname, ok := bsonData["nickname"].(string); ok {
		user.Nickname = nickname
	}
	if avatarUrl, ok := bsonData["avatar_url"].(string); ok {
		user.AvatarUrl = avatarUrl
	}
	if emailAddress, ok := bsonData["email_address"].(string); ok {
		user.EmailAddress = emailAddress
	}
	if gender, ok := bsonData["gender"].(string); ok {
		user.Gender = gender
	}

	// Convert time fields
	if createTime, ok := bsonData["create_time"].(time.Time); ok {
		user.CreateTime = createTime
	}
	if updateTime, ok := bsonData["update_time"].(time.Time); ok {
		user.UpdateTime = updateTime
	}

	// Convert BaseEntity fields
	if createUserId, ok := bsonData["create_user_id"].(primitive.ObjectID); ok {
		user.CreateUserId = createUserId
	}
	if updateUserId, ok := bsonData["update_user_id"].(primitive.ObjectID); ok {
		user.UpdateUserId = updateUserId
	}
	if deleteUserId, ok := bsonData["delete_user_id"].(primitive.ObjectID); ok {
		user.DeleteUserId = &deleteUserId
	}
	if deleteTime, ok := bsonData["delete_time"].(time.Time); ok {
		user.DeleteTime = &deleteTime
	}
	if isDelete, ok := bsonData["is_delete"].(bool); ok {
		user.IsDelete = isDelete
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
		primitive.E{Key: "nickname", Value: user.Nickname},
		primitive.E{Key: "avatar_url", Value: user.AvatarUrl},
		primitive.E{Key: "email_address", Value: user.EmailAddress},
		primitive.E{Key: "gender", Value: user.Gender},
		primitive.E{Key: "create_user_id", Value: user.CreateUserId},
		primitive.E{Key: "create_time", Value: user.CreateTime},
		primitive.E{Key: "update_user_id", Value: user.UpdateUserId},
		primitive.E{Key: "update_time", Value: user.UpdateTime},
		primitive.E{Key: "delete_user_id", Value: user.DeleteUserId},
		primitive.E{Key: "delete_time", Value: user.DeleteTime},
		primitive.E{Key: "is_delete", Value: user.IsDelete},
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
		primitive.E{Key: "is_delete", Value: false},
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
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2UserEntity(database.GetOneInMongoDB(filter))
}

// GetUserByEmail retrieves a user by their email address from MongoDB
func (m UserMongoDbMapper) GetUserByEmail(email string) model.UserEntity {
	// Create a filter to find the user by email
	filter := bson.D{
		primitive.E{Key: "email_address", Value: email},
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()
	return convertBsonM2UserEntity(database.GetOneInMongoDB(filter))
}

// InsertUserByEntity inserts a new user entity into MongoDB
func (m UserMongoDbMapper) InsertUserByEntity(newEntity model.UserEntity) string {
	// Set default values if not provided
	if newEntity.CreateTime.IsZero() {
		newEntity.CreateTime = time.Now()
	}
	if newEntity.UpdateTime.IsZero() {
		newEntity.UpdateTime = time.Now()
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
	updatedEntity.UpdateTime = time.Now()

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Update the user document in the database
	database.UpdateManyInMongoDB(filter, convertUserEntity2BsonD(updatedEntity))

	// Return the updated user
	return m.GetUserByObjectId(plainId)
}

// GetAllUsers retrieves all users with pagination from MongoDB
func (m UserMongoDbMapper) GetAllUsers(limit, offset int) []model.UserEntity {
	// Create a filter to get all active users
	filter := bson.D{
		primitive.E{Key: "is_delete", Value: false},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Get the user documents from the database
	var users []model.UserEntity
	queryResultList := database.GetManyInMongoDBWithPagination(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		users = append(users, convertBsonM2UserEntity(queryResult))
	}

	return users
}

// GetAllUsersIncludeDeleted retrieves all users including deleted ones with pagination from MongoDB
func (m UserMongoDbMapper) GetAllUsersIncludeDeleted(limit, offset int) []model.UserEntity {
	// Create an empty filter to get all users
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Get the user documents from the database
	var users []model.UserEntity
	queryResultList := database.GetManyInMongoDBWithPaginationIncludeDeleted(filter, int64(limit), int64(offset))
	for _, queryResult := range queryResultList {
		users = append(users, convertBsonM2UserEntity(queryResult))
	}

	return users
}

// CountAllUsers returns the total number of users from MongoDB
func (m UserMongoDbMapper) CountAllUsers() int64 {
	// Create a filter to count all active users
	filter := bson.D{
		primitive.E{Key: "is_delete", Value: false},
	}

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
	
	// Soft delete: Update is_delete to true
	now := time.Now()
	newUsername := user.Username + "_deleted_" + util.FormatDateToStringWithoutDash(now)
	update := bson.D{
		primitive.E{Key: "is_delete", Value: true},
		primitive.E{Key: "delete_time", Value: now},
		// Rename username to avoid conflict if reused
		primitive.E{Key: "username", Value: newUsername},
	}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Update the user document in the database
	database.UpdateManyInMongoDB(filter, update)

	user.IsDelete = true
	user.DeleteTime = &now
	user.Username = newUsername
	return user
}

// TruncateUsers deletes all users from MongoDB
func (m UserMongoDbMapper) TruncateUsers() error {
	// Create an empty filter to delete all users
	filter := bson.D{}

	database.OpenMongoDbConnection(database.UserTableName)
	defer database.CloseMongoDbConnection()

	// Delete all user documents from the database
	deletedCount := database.DeleteManyInMongoDBIncludeDeleted(filter)
	util.Logger.Infow("Users truncated successfully", "deleted_count", deletedCount)
	return nil
}

// GetUsersByRole retrieves all users with a specific role from MongoDB
func (m UserMongoDbMapper) GetUsersByRole(role string) []model.UserEntity {
	// Create a filter to find users by role
	filter := bson.D{
		primitive.E{Key: "role", Value: role},
		primitive.E{Key: "is_delete", Value: false},
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