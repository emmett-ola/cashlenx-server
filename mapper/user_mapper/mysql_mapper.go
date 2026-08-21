package user_mapper

import (
	"bytes"
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
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, phone_number, location, birth_date, create_user_id, create_time, update_user_id, update_time FROM ")
	sqlString.WriteString(database.UserTableName)
	sqlString.WriteString(" WHERE id = ? ")
	sqlString.WriteString(" AND is_delete = FALSE") // Added explicit check just in case SqlExcludeDeleted is not enough or for clarity

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(sqlString.String(), plainId)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createTime, updateTime time.Time
	var id, createUserId, updateUserId string
	var nickname, avatarUrl, emailAddress, gender, phoneNumber, location, birthDate sql.NullString

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &phoneNumber, &location, &birthDate, &createUserId, &createTime, &updateUserId, &updateTime)
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
	user.CreateUserId = util.Convert2ObjectId(createUserId)
	user.CreateTime = createTime
	user.UpdateUserId = util.Convert2ObjectId(updateUserId)
	user.UpdateTime = updateTime

	// Set profile fields if not null
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if avatarUrl.Valid {
		user.AvatarUrl = avatarUrl.String
	}
	if emailAddress.Valid {
		user.EmailAddress = emailAddress.String
	}
	if gender.Valid {
		user.Gender = gender.String
	}
	if phoneNumber.Valid {
		user.PhoneNumber = phoneNumber.String
	}
	if location.Valid {
		user.Location = location.String
	}
	if birthDate.Valid {
		user.BirthDate = birthDate.String
	}

	return user
}

// GetUserByUsername retrieves a user by their username from MySQL
func (m UserMySqlMapper) GetUserByUsername(username string) model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserTableName + ` WHERE username = ? AND is_delete = FALSE`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, username)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createTime, updateTime time.Time
	var id, createUserId, updateUserId string
	var nickname, avatarUrl, emailAddress, gender sql.NullString

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
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
	user.CreateUserId = util.Convert2ObjectId(createUserId)
	user.CreateTime = createTime
	user.UpdateUserId = util.Convert2ObjectId(updateUserId)
	user.UpdateTime = updateTime

	// Set profile fields if not null
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if avatarUrl.Valid {
		user.AvatarUrl = avatarUrl.String
	}
	if emailAddress.Valid {
		user.EmailAddress = emailAddress.String
	}
	if gender.Valid {
		user.Gender = gender.String
	}

	return user
}

// GetUserByUsernameIncludeDeleted retrieves a user by their username including deleted ones from MySQL
func (m UserMySqlMapper) GetUserByUsernameIncludeDeleted(username string) model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserTableName + ` WHERE username = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, username)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createTime, updateTime time.Time
	var id, createUserId, updateUserId string
	var nickname, avatarUrl, emailAddress, gender sql.NullString

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
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
	user.CreateUserId = util.Convert2ObjectId(createUserId)
	user.CreateTime = createTime
	user.UpdateUserId = util.Convert2ObjectId(updateUserId)
	user.UpdateTime = updateTime

	// Set profile fields if not null
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if avatarUrl.Valid {
		user.AvatarUrl = avatarUrl.String
	}
	if emailAddress.Valid {
		user.EmailAddress = emailAddress.String
	}
	if gender.Valid {
		user.Gender = gender.String
	}

	return user
}

// GetUserByEmail retrieves a user by their email address from MySQL
func (m UserMySqlMapper) GetUserByEmail(email string) model.UserEntity {
	// Create the SQL query
	query := `SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserTableName + ` WHERE email_address = ? AND is_delete = FALSE`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, email)

	// Scan the result into a UserEntity
	var user model.UserEntity
	var createTime, updateTime time.Time
	var id, createUserId, updateUserId string
	var nickname, avatarUrl, emailAddress, gender sql.NullString

	err := row.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("User not found", "email", email)
			return model.UserEntity{}
		}
		util.Logger.Errorw("Failed to get user", "error", err, "email", email)
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
	user.CreateUserId = util.Convert2ObjectId(createUserId)
	user.CreateTime = createTime
	user.UpdateUserId = util.Convert2ObjectId(updateUserId)
	user.UpdateTime = updateTime

	// Set profile fields if not null
	if nickname.Valid {
		user.Nickname = nickname.String
	}
	if avatarUrl.Valid {
		user.AvatarUrl = avatarUrl.String
	}
	if emailAddress.Valid {
		user.EmailAddress = emailAddress.String
	}
	if gender.Valid {
		user.Gender = gender.String
	}

	return user
}

// InsertUserByEntity inserts a new user entity into MySQL
func (m UserMySqlMapper) InsertUserByEntity(newEntity model.UserEntity) string {
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

	// Generate a new ObjectID if not provided
	if newEntity.Id.IsZero() {
		newEntity.Id = primitive.NewObjectID()
	}

	// Create the SQL query
	query := `INSERT INTO ` + database.UserTableName + ` (id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time, is_delete) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		newEntity.Nickname,
		newEntity.AvatarUrl,
		newEntity.EmailAddress,
		newEntity.Gender,
		newEntity.CreateUserId.Hex(),
		newEntity.CreateTime,
		newEntity.UpdateUserId.Hex(),
		newEntity.UpdateTime,
		false, // is_delete
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
	updatedEntity.UpdateTime = time.Now()

	// Create the SQL query
	query := `UPDATE ` + database.UserTableName + ` SET username = ?, password_hash = ?, is_active = ?, role = ?, nickname = ?, avatar_url = ?, email_address = ?, gender = ?, phone_number = COALESCE(NULLIF(?, ''), phone_number), location = COALESCE(NULLIF(?, ''), location), birth_date = COALESCE(NULLIF(?, ''), birth_date), update_user_id = ?, update_time = ? WHERE id = ? AND is_delete = FALSE`

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
		updatedEntity.Nickname,
		updatedEntity.AvatarUrl,
		updatedEntity.EmailAddress,
		updatedEntity.Gender,
		updatedEntity.PhoneNumber,
		updatedEntity.Location,
		updatedEntity.BirthDate,
		updatedEntity.UpdateUserId.Hex(),
		updatedEntity.UpdateTime,
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
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ")
	sqlString.WriteString(database.UserTableName)
	sqlString.WriteString(" WHERE is_delete = FALSE")
	sqlString.WriteString(" LIMIT ? OFFSET ?")

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(sqlString.String(), limit, offset)
	if err != nil {
		util.Logger.Errorw("Failed to get all users", "error", err)
		return []model.UserEntity{}
	}
	defer rows.Close()

	// Scan the results into a slice of UserEntity
	var users []model.UserEntity

	for rows.Next() {
		var user model.UserEntity
		var createTime, updateTime time.Time
		var id, createUserId, updateUserId string
		var nickname, avatarUrl, emailAddress, gender sql.NullString

		err := rows.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
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
		user.CreateUserId = util.Convert2ObjectId(createUserId)
		user.CreateTime = createTime
		user.UpdateUserId = util.Convert2ObjectId(updateUserId)
		user.UpdateTime = updateTime

		// Set profile fields if not null
		if nickname.Valid {
			user.Nickname = nickname.String
		}
		if avatarUrl.Valid {
			user.AvatarUrl = avatarUrl.String
		}
		if emailAddress.Valid {
			user.EmailAddress = emailAddress.String
		}
		if gender.Valid {
			user.Gender = gender.String
		}

		// Add the user to the slice
		users = append(users, user)
	}

	return users
}

// GetAllUsersIncludeDeleted retrieves all users including deleted ones with pagination from MySQL
func (m UserMySqlMapper) GetAllUsersIncludeDeleted(limit, offset int) []model.UserEntity {
	// Create the SQL query with pagination
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ")
	sqlString.WriteString(database.UserTableName)
	// No SqlExcludeDeleted
	sqlString.WriteString(" LIMIT ? OFFSET ?")

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(sqlString.String(), limit, offset)
	if err != nil {
		util.Logger.Errorw("Failed to get all users", "error", err)
		return []model.UserEntity{}
	}
	defer rows.Close()

	// Scan the results into a slice of UserEntity
	var users []model.UserEntity

	for rows.Next() {
		var user model.UserEntity
		var createTime, updateTime time.Time
		var id, createUserId, updateUserId string
		var nickname, avatarUrl, emailAddress, gender sql.NullString

		err := rows.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
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
		user.CreateUserId = util.Convert2ObjectId(createUserId)
		user.CreateTime = createTime
		user.UpdateUserId = util.Convert2ObjectId(updateUserId)
		user.UpdateTime = updateTime

		// Set profile fields if not null
		if nickname.Valid {
			user.Nickname = nickname.String
		}
		if avatarUrl.Valid {
			user.AvatarUrl = avatarUrl.String
		}
		if emailAddress.Valid {
			user.EmailAddress = emailAddress.String
		}
		if gender.Valid {
			user.Gender = gender.String
		}

		// Add the user to the slice
		users = append(users, user)
	}

	return users
}

// CountAllUsers returns the total number of users from MySQL
func (m UserMySqlMapper) CountAllUsers() int64 {
	// Create the SQL query
	var sqlString bytes.Buffer
	sqlString.WriteString("SELECT COUNT(*) FROM ")
	sqlString.WriteString(database.UserTableName)
	sqlString.WriteString(" WHERE is_delete = FALSE")

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	var count int64
	err := connection.QueryRow(sqlString.String()).Scan(&count)
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

	// Create the SQL query (Soft Delete)
	// Note: We need the operator ID here, but the interface signature doesn't provide it.
	// For now, we'll set delete_user_id to the user's own ID or empty if unknown.
	// Ideally, the interface should be updated, but for quick fix we use soft delete with current time.
	// We no longer rename username to allow unique constraint check on registration to fail if username exists (even if deleted).

	now := time.Now()

	query := `UPDATE ` + database.UserTableName + ` SET is_delete = TRUE, delete_time = NOW() WHERE id = ? AND is_delete = FALSE`

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

	user.IsDelete = true
	user.DeleteTime = &now
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
	query := `SELECT id, username, password_hash, is_active, role, nickname, avatar_url, email_address, gender, create_user_id, create_time, update_user_id, update_time FROM ` + database.UserTableName + ` WHERE role = ? AND is_delete = FALSE`

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
		var createTime, updateTime time.Time
		var id, createUserId, updateUserId string
		var nickname, avatarUrl, emailAddress, gender sql.NullString

		err := rows.Scan(&id, &user.Username, &user.PasswordHash, &user.IsActive, &user.Role, &nickname, &avatarUrl, &emailAddress, &gender, &createUserId, &createTime, &updateUserId, &updateTime)
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
		user.CreateUserId = util.Convert2ObjectId(createUserId)
		user.CreateTime = createTime
		user.UpdateUserId = util.Convert2ObjectId(updateUserId)
		user.UpdateTime = updateTime

		// Set profile fields if not null
		if nickname.Valid {
			user.Nickname = nickname.String
		}
		if avatarUrl.Valid {
			user.AvatarUrl = avatarUrl.String
		}
		if emailAddress.Valid {
			user.EmailAddress = emailAddress.String
		}
		if gender.Valid {
			user.Gender = gender.String
		}

		// Add the user to the slice
		users = append(users, user)
	}

	return users
}
