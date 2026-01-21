package refresh_token_mapper

import (
	"database/sql"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

// RefreshTokenMySqlMapper implements RefreshTokenMapper for MySQL
type RefreshTokenMySqlMapper struct{}

// CreateToken creates a new refresh token in MySQL
func (m RefreshTokenMySqlMapper) CreateToken(token model.RefreshToken) string {
	// Create the SQL query
	query := `INSERT INTO auth_token_refresh (id, user_id, token, expires_at, 
		device_id, device_name, ip_address, user_agent,
		create_time, create_user_id, update_time, update_user_id, is_delete) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		token.DeviceId,
		token.DeviceName,
		token.IPAddress,
		token.UserAgent,
		token.CreateTime,
		token.CreateUserId.Hex(),
		token.UpdateTime,
		token.UpdateUserId.Hex(),
		token.IsDelete,
	)
	if err != nil {
		util.Logger.Errorw("Failed to create refresh token", "error", err, "user_id", token.UserId)
		return ""
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		util.Logger.Errorw("Failed to create refresh token (no rows affected)", "error", err, "user_id", token.UserId)
		return ""
	}

	return token.Token
}

// GetTokenByToken retrieves a refresh token by its token string from MySQL
func (m RefreshTokenMySqlMapper) GetTokenByToken(tokenStr string) model.RefreshToken {
	// Create the SQL query
	query := `SELECT id, user_id, token, expires_at, create_time, create_user_id, revoked_at, revoked_by,
		device_id, device_name, ip_address, user_agent, update_time, update_user_id 
		FROM auth_token_refresh WHERE token = ? AND (revoked_at IS NULL OR revoked_at > ?)`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	row := connection.QueryRow(query, tokenStr, time.Now())

	// Scan the result into a RefreshToken
	var token model.RefreshToken
	var revokedAt sql.NullTime
	var revokedBy sql.NullString
	var deviceId, deviceName, ipAddress, userAgent sql.NullString
	var createUserIdStr, updateUserIdStr string
	var updateTime sql.NullTime

	err := row.Scan(&token.Id, &token.UserId, &token.Token, &token.ExpiresAt, &token.CreateTime, &createUserIdStr,
		&revokedAt, &revokedBy, &deviceId, &deviceName, &ipAddress, &userAgent, &updateTime, &updateUserIdStr)
	if err != nil {
		if err == sql.ErrNoRows {
			util.Logger.Debugw("Refresh token not found", "token", tokenStr)
			return model.RefreshToken{}
		}
		util.Logger.Errorw("Failed to get refresh token", "error", err, "token", tokenStr)
		return model.RefreshToken{}
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
	if revokedAt.Valid {
		token.RevokedAt = &revokedAt.Time
	}
	if revokedBy.Valid {
		token.RevokedBy = revokedBy.String
	}
	if deviceId.Valid {
		token.DeviceId = deviceId.String
	}
	if deviceName.Valid {
		token.DeviceName = deviceName.String
	}
	if ipAddress.Valid {
		token.IPAddress = ipAddress.String
	}
	if userAgent.Valid {
		token.UserAgent = userAgent.String
	}

	// Check if token is expired
	if token.ExpiresAt.Before(time.Now()) {
		util.Logger.Debugw("Refresh token expired", "token", tokenStr)
		return model.RefreshToken{}
	}

	return token
}

// RevokeToken revokes a refresh token by its token string
func (m RefreshTokenMySqlMapper) RevokeToken(tokenStr string, revokedBy string) error {
	// Create the SQL query
	query := `UPDATE auth_token_refresh SET revoked_at = ?, revoked_by = ?, update_time = ?, update_user_id = ? WHERE token = ?`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	updateUserId := util.Convert2ObjectId(revokedBy)

	// Execute the query
	result, err := connection.Exec(query, time.Now(), revokedBy, time.Now(), updateUserId.Hex(), tokenStr)
	if err != nil {
		util.Logger.Errorw("Failed to revoke refresh token", "error", err, "token", tokenStr)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		util.Logger.Errorw("Failed to check rows affected", "error", err, "token", tokenStr)
		return err
	}

	if rowsAffected == 0 {
		util.Logger.Debugw("Refresh token not found for revocation", "token", tokenStr)
		return sql.ErrNoRows
	}

	return nil
}

// RevokeAllTokensByUserId revokes all refresh tokens for a user
func (m RefreshTokenMySqlMapper) RevokeAllTokensByUserId(userId string) error {
	// Create the SQL query
	query := `UPDATE auth_token_refresh SET revoked_at = ?, revoked_by = ?, update_time = ?, update_user_id = ? WHERE user_id = ? AND revoked_at IS NULL`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	updateUserId := util.Convert2ObjectId(userId)

	// Execute the query
	_, err := connection.Exec(query, time.Now(), userId, time.Now(), updateUserId.Hex(), userId)
	if err != nil {
		util.Logger.Errorw("Failed to revoke all refresh tokens for user", "error", err, "user_id", userId)
		return err
	}

	return nil
}

// GetTokensByUserId retrieves all refresh tokens for a user
func (m RefreshTokenMySqlMapper) GetTokensByUserId(userId string) []model.RefreshToken {
	var tokens []model.RefreshToken

	// Create the SQL query
	query := `SELECT id, user_id, token, expires_at, create_time, create_user_id, revoked_at, revoked_by,
		device_id, device_name, ip_address, user_agent, update_time, update_user_id
		FROM auth_token_refresh 
		WHERE user_id = ? 
		ORDER BY create_time DESC`

	// Get database connection
	connection := database.GetMySqlConnection()
	defer database.CloseMySqlConnection()

	// Execute the query
	rows, err := connection.Query(query, userId)
	if err != nil {
		util.Logger.Errorw("Failed to get refresh tokens by user ID", "error", err, "userId", userId)
		return tokens
	}
	defer rows.Close()

	for rows.Next() {
		var token model.RefreshToken
		var revokedAt sql.NullTime
		var revokedBy sql.NullString
		var deviceId, deviceName, ipAddress, userAgent sql.NullString
		var createUserIdStr, updateUserIdStr string
		var updateTime sql.NullTime

		err := rows.Scan(&token.Id, &token.UserId, &token.Token, &token.ExpiresAt, &token.CreateTime, &createUserIdStr,
			&revokedAt, &revokedBy, &deviceId, &deviceName, &ipAddress, &userAgent, &updateTime, &updateUserIdStr)
		if err != nil {
			util.Logger.Errorw("Failed to scan refresh token", "error", err)
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
		if revokedAt.Valid {
			token.RevokedAt = &revokedAt.Time
		}
		if revokedBy.Valid {
			token.RevokedBy = revokedBy.String
		}
		if deviceId.Valid {
			token.DeviceId = deviceId.String
		}
		if deviceName.Valid {
			token.DeviceName = deviceName.String
		}
		if ipAddress.Valid {
			token.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			token.UserAgent = userAgent.String
		}

		tokens = append(tokens, token)
	}

	return tokens
}
