package password_reset_mapper

import (
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

// PasswordResetMySqlMapper implements PasswordResetMapper for MySQL
type PasswordResetMySqlMapper struct{}

// CreateToken creates a new password reset token in MySQL
func (m PasswordResetMySqlMapper) CreateToken(token model.PasswordResetToken) string {
	// Create the SQL query
	query := `INSERT INTO auth_token_password_reset (id, user_id, token, expires_at, 
		create_time, create_user_id, update_time, update_user_id, is_delete, ip_address) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	result, err := connection.Exec(
		query,
		token.Id,
		token.UserId,
		token.Token,
		token.ExpiresAt,
		token.CreateTime,
		token.CreateUserId.Hex(),
		token.UpdateTime,
		token.UpdateUserId.Hex(),
		token.IsDelete,
		token.IPAddress,
	)
	if err != nil {
		util.Logger.Errorw("Failed to create password reset token", "error", err, "user_id", token.UserId)
		return ""
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to create password reset token (no rows affected)", "error", err, "user_id", token.UserId)
		return ""
	}

	return token.Token
}

// GetTokenByToken retrieves a password reset token by its token string from MySQL
func (m PasswordResetMySqlMapper) GetTokenByToken(tokenStr string) model.PasswordResetToken {
	// Create the SQL query
	query := `SELECT id, user_id, token, expires_at, create_time, create_user_id, used_at,
		update_time, update_user_id, ip_address
		FROM auth_token_password_reset WHERE token = ? AND is_delete = 0`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, tokenStr)

	// Scan the result
	var token model.PasswordResetToken
	var usedAt sql.NullTime
	var createUserIdStr, updateUserIdStr string
	var updateTime sql.NullTime
	var ipAddress sql.NullString

	err := row.Scan(&token.Id, &token.UserId, &token.Token, &token.ExpiresAt, &token.CreateTime, &createUserIdStr,
		&usedAt, &updateTime, &updateUserIdStr, &ipAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("Password reset token not found", "token", tokenStr)
			return model.PasswordResetToken{}
		}
		util.Logger.Errorw("Failed to get password reset token", "error", err, "token", tokenStr)
		return model.PasswordResetToken{}
	}

	// Parse ObjectIDs
	if createUserIdStr != "" {
		token.CreateUserId, _ = primitive.ObjectIDFromHex(createUserIdStr)
	}
	if updateUserIdStr != "" {
		token.UpdateUserId, _ = primitive.ObjectIDFromHex(updateUserIdStr)
	}
	if updateTime.Valid {
		token.UpdateTime = updateTime.Time
	}

	// Handle nullable fields
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	if ipAddress.Valid {
		token.IPAddress = ipAddress.String
	}

	return token
}

// GetTokensByUserId retrieves all password reset tokens for a user from MySQL
func (m PasswordResetMySqlMapper) GetTokensByUserId(userId string) []model.PasswordResetToken {
	var tokens []model.PasswordResetToken

	// Create the SQL query
	query := `SELECT id, user_id, token, expires_at, create_time, create_user_id, used_at,
		update_time, update_user_id, ip_address
		FROM auth_token_password_reset 
		WHERE user_id = ? AND is_delete = 0
		ORDER BY create_time DESC`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(query, userId)
	if err != nil {
		util.Logger.Errorw("Failed to get password reset tokens by user ID", "error", err, "userId", userId)
		return tokens
	}
	defer rows.Close()

	for rows.Next() {
		var token model.PasswordResetToken
		var usedAt sql.NullTime
		var createUserIdStr, updateUserIdStr string
		var updateTime sql.NullTime
		var ipAddress sql.NullString

		err := rows.Scan(&token.Id, &token.UserId, &token.Token, &token.ExpiresAt, &token.CreateTime, &createUserIdStr,
			&usedAt, &updateTime, &updateUserIdStr, &ipAddress)
		if err != nil {
			util.Logger.Errorw("Failed to scan password reset token", "error", err)
			continue
		}

		// Parse ObjectIDs
		if createUserIdStr != "" {
			token.CreateUserId, _ = primitive.ObjectIDFromHex(createUserIdStr)
		}
		if updateUserIdStr != "" {
			token.UpdateUserId, _ = primitive.ObjectIDFromHex(updateUserIdStr)
		}
		if updateTime.Valid {
			token.UpdateTime = updateTime.Time
		}

		// Handle nullable fields
		if usedAt.Valid {
			token.UsedAt = &usedAt.Time
		}
		if ipAddress.Valid {
			token.IPAddress = ipAddress.String
		}

		tokens = append(tokens, token)
	}

	return tokens
}

// MarkTokenAsUsed marks a password reset token as used in MySQL
func (m PasswordResetMySqlMapper) MarkTokenAsUsed(tokenId string) error {
	// Create the SQL query
	query := `UPDATE auth_token_password_reset SET used_at = ?, update_time = ?, update_user_id = ? WHERE id = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	updateUserId := util.Convert2ObjectId(tokenId) // Using token ID as placeholder

	// Execute the query
	result, err := connection.Exec(query, time.Now(), time.Now(), updateUserId.Hex(), tokenId)
	if err != nil {
		util.Logger.Errorw("Failed to mark password reset token as used", "error", err, "token_id", tokenId)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteToken deletes a password reset token from MySQL
func (m PasswordResetMySqlMapper) DeleteToken(tokenId string) error {
	// Create the SQL query
	query := `UPDATE auth_token_password_reset SET is_delete = 1, delete_time = ? WHERE id = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	result, err := connection.Exec(query, time.Now(), tokenId)
	if err != nil {
		util.Logger.Errorw("Failed to delete password reset token", "error", err, "token_id", tokenId)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteExpiredTokens deletes all expired password reset tokens from MySQL
func (m PasswordResetMySqlMapper) DeleteExpiredTokens() error {
	// Create the SQL query
	query := `UPDATE auth_token_password_reset SET is_delete = 1, delete_time = ? WHERE expires_at < ? AND is_delete = 0`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	_, err := connection.Exec(query, time.Now(), time.Now())
	if err != nil {
		util.Logger.Errorw("Failed to delete expired password reset tokens", "error", err)
		return err
	}

	return nil
}

// DeleteTokensByUserId deletes all password reset tokens for a user from MySQL
func (m PasswordResetMySqlMapper) DeleteTokensByUserId(userId string) error {
	// Create the SQL query
	query := `UPDATE auth_token_password_reset SET is_delete = 1, delete_time = ?, delete_user_id = ? WHERE user_id = ? AND is_delete = 0`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	deleteUserId := util.Convert2ObjectId(userId)

	// Execute the query
	_, err := connection.Exec(query, time.Now(), deleteUserId.Hex(), userId)
	if err != nil {
		util.Logger.Errorw("Failed to delete password reset tokens for user", "error", err, "user_id", userId)
		return err
	}

	return nil
}

// InvalidateActiveTokensByUserId invalidates all active tokens for a user
func (m PasswordResetMySqlMapper) InvalidateActiveTokensByUserId(userId string) error {
	query := `UPDATE auth_token_password_reset 
		SET is_delete = 1, delete_time = ?, delete_user_id = ? 
		WHERE user_id = ? AND used_at IS NULL AND expires_at > ? AND is_delete = 0`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	deleteUserId := util.Convert2ObjectId(userId)

	_, err := connection.Exec(query, time.Now(), deleteUserId.Hex(), userId, time.Now())
	return err
}

// CountTokensByUserIdAndDateRange counts tokens for a user in a range
func (m PasswordResetMySqlMapper) CountTokensByUserIdAndDateRange(userId string, from, to int64) (int64, error) {
	query := `SELECT COUNT(*) FROM auth_token_password_reset 
		WHERE user_id = ? AND create_time BETWEEN ? AND ?`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var count int64
	err := connection.QueryRow(query, userId, time.Unix(from, 0), time.Unix(to, 0)).Scan(&count)
	return count, err
}

// CountTokensByIPAndDateRange counts tokens for an IP in a range
func (m PasswordResetMySqlMapper) CountTokensByIPAndDateRange(ipAddress string, from, to int64) (int64, error) {
	query := `SELECT COUNT(DISTINCT user_id) FROM auth_token_password_reset 
		WHERE ip_address = ? AND create_time BETWEEN ? AND ?`

	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	var count int64
	err := connection.QueryRow(query, ipAddress, time.Unix(from, 0), time.Unix(to, 0)).Scan(&count)
	return count, err
}
