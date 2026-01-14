package user_mapper

import (
	"database/sql"
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserMySqlMapper MySQL implementation of UserMapper
type UserMySqlMapper struct{}

// GetUserByObjectId retrieves a user by their ID from MySQL
func (m UserMySqlMapper) GetUserByObjectId(plainId string) model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, created_at, updated_at FROM ` + database.UserTableName + ` WHERE id = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, plainId)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createdAt, updatedAt time.Time
	var id string

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("User not found", "userId", plainId)
			return model.UserEntity{}
		}
		util.Logger.Errorw("Failed to get user", "error", err, "userId", plainId)
		return model.UserEntity{}
	}

	// Parse the ID string to ObjectID
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		util.Logger.Errorw("Invalid ObjectID", "error", err, "id", id)
		return model.UserEntity{}
	}

	// Set the parsed ID and timestamps
	user.Id = objectId
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user
}

// GetUserByUsername retrieves a user by their username from MySQL
func (m UserMySqlMapper) GetUserByUsername(username string) model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, created_at, updated_at FROM ` + database.UserTableName + ` WHERE username = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, username)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createdAt, updatedAt time.Time
	var id string

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("User not found", "username", username)
			return model.UserEntity{}
		}
		util.Logger.Errorw("Failed to get user", "error", err, "username", username)
		return model.UserEntity{}
	}

	// Parse the ID string to ObjectID
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		util.Logger.Errorw("Invalid ObjectID", "error", err, "id", id)
		return model.UserEntity{}
	}

	// Set the parsed ID and timestamps
	user.Id = objectId
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return user
}

// InsertUserByEntity inserts a new user entity into MySQL
func (m UserMySqlMapper) InsertUserByEntity(newEntity model.UserEntity) string {
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

	// Generate a new ObjectID if not provided
	if newEntity.Id.IsZero() {
		newEntity.Id = primitive.NewObjectID()
	}

	// Create the SQL query
	query := `INSERT INTO ` + database.UserTableName + ` (id, username, password_hash, is_active, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	result, err := connection.Exec(
		query,
		newEntity.Id.Hex(),
		newEntity.Username,
		newEntity.PasswordHash,
		newEntity.IsActive,
		newEntity.Role,
		newEntity.CreatedAt,
		newEntity.UpdatedAt,
	)
	if err != nil {
		util.Logger.Errorw("Failed to insert user", "error", err, "username", newEntity.Username)
		return ""
	}

	// Check if the insert was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to insert user", "error", err, "username", newEntity.Username)
		return ""
	}

	// Return the inserted ID as a string
	return newEntity.Id.Hex()
}

// UpdateUserByEntity updates an existing user entity in MySQL
func (m UserMySqlMapper) UpdateUserByEntity(plainId string, updatedEntity model.UserEntity) model.UserEntity {
	// Set the updated timestamp
	updatedEntity.UpdatedAt = time.Now()

	// Create the SQL query
	query := `UPDATE ` + database.UserTableName + ` SET username = ?, password_hash = ?, is_active = ?, role = ?, updated_at = ? WHERE id = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	result, err := connection.Exec(
		query,
		updatedEntity.Username,
		updatedEntity.PasswordHash,
		updatedEntity.IsActive,
		updatedEntity.Role,
		updatedEntity.UpdatedAt,
		plainId,
	)
	if err != nil {
		util.Logger.Errorw("Failed to update user", "error", err, "userId", plainId)
		return model.UserEntity{}
	}

	// Check if the update was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to update user", "error", err, "userId", plainId)
		return model.UserEntity{}
	}

	// Return the updated user by fetching it from the database
	return m.GetUserByObjectId(plainId)
}

// GetAllUsers retrieves all users with pagination from MySQL
func (m UserMySqlMapper) GetAllUsers(limit, offset int) []model.UserEntity {
	// Create the SQL query with pagination
	query := `SELECT id, username, password_hash, is_active, role, created_at, updated_at FROM ` + database.UserTableName + ` LIMIT ? OFFSET ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(query, limit, offset)
	if err != nil {
		util.Logger.Errorw("Failed to get all users", "error", err)
		return []model.UserEntity{}
	}
	defer rows.Close()

	// Scan the results into a slice of UserEntity
	var users []model.UserEntity

	for rows.Next() {
		var user model.UserEntity
		var createdAt, updatedAt time.Time
		var id string

		err := rows.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &createdAt, &updatedAt)
		if err != nil {
			util.Logger.Errorw("Failed to scan user", "error", err)
			continue
		}

		// Parse the ID string to ObjectID
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			util.Logger.Errorw("Invalid ObjectID", "error", err, "id", id)
			continue
		}

		// Set the parsed ID and timestamps
		user.Id = objectId
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt

		// Add the user to the slice
		users = append(users, user)
	}

	return users
}

// CountAllUsers returns the total number of users from MySQL
func (m UserMySqlMapper) CountAllUsers() int64 {
	// Create the SQL query
	query := `SELECT COUNT(*) FROM ` + database.UserTableName

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	var count int64
	err := connection.QueryRow(query).Scan(&count)
	if err != nil {
		util.Logger.Errorw("Failed to count users", "error", err)
		return 0
	}

	return count
}

// DeleteUserByObjectId deletes a user by their ID from MySQL
func (m UserMySqlMapper) DeleteUserByObjectId(plainId string) model.UserEntity {
	// First, get the user to return it
	user := m.GetUserByObjectId(plainId)
	if user.IsEmpty() {
		return model.UserEntity{}
	}

	// Create the SQL query
	query := `DELETE FROM ` + database.UserTableName + ` WHERE id = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	result, err := connection.Exec(query, plainId)
	if err != nil {
		util.Logger.Errorw("Failed to delete user", "error", err, "userId", plainId)
		return model.UserEntity{}
	}

	// Check if the delete was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to delete user", "error", err, "userId", plainId)
		return model.UserEntity{}
	}

	return user
}

// TruncateUsers deletes all users from MySQL
func (m UserMySqlMapper) TruncateUsers() error {
	// Create the SQL query
	query := `TRUNCATE TABLE ` + database.UserTableName

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	_, err := connection.Exec(query)
	if err != nil {
		util.Logger.Errorw("Failed to truncate users", "error", err)
		return err
	}

	util.Logger.Infow("Users truncated successfully")
	return nil
}

// GetUsersByRole retrieves all users with a specific role from MySQL
func (m UserMySqlMapper) GetUsersByRole(role string) []model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, created_at, updated_at FROM ` + database.UserTableName + ` WHERE role = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(query, role)
	if err != nil {
		util.Logger.Errorw("Failed to get users by role", "error", err, "role", role)
		return []model.UserEntity{}
	}
	defer rows.Close()

	// Scan the results into a slice of UserEntity
	var users []model.UserEntity

	for rows.Next() {
		var user model.UserEntity
		var createdAt, updatedAt time.Time
		var id string

		err := rows.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &createdAt, &updatedAt)
		if err != nil {
			util.Logger.Errorw("Failed to scan user", "error", err)
			continue
		}

		// Parse the ID string to ObjectID
		objectId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			util.Logger.Errorw("Invalid ObjectID", "error", err, "id", id)
			continue
		}

		// Set the parsed ID and timestamps
		user.Id = objectId
		user.CreatedAt = createdAt
		user.UpdatedAt = updatedAt

		// Add the user to the slice
		users = append(users, user)
	}

	return users
}