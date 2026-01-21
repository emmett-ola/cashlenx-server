package password_reset_mapper

import (
	"time"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

// PasswordResetMongoDbMapper implements PasswordResetMapper for MongoDB
type PasswordResetMongoDbMapper struct{}

// CreateToken creates a new password reset token in MongoDB
func (m PasswordResetMongoDbMapper) CreateToken(token model.PasswordResetToken) string {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Insert the token into the collection
	_, err := collection.InsertOne(nil, token)
	if err != nil {
		util.Logger.Errorw("Failed to create password reset token", "error", err, "user_id", token.UserId)
		return ""
	}

	return token.Token
}

// GetTokenByToken retrieves a password reset token by its token string from MongoDB
func (m PasswordResetMongoDbMapper) GetTokenByToken(tokenStr string) model.PasswordResetToken {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter
	filter := map[string]interface{}{
		"token": tokenStr,
		"$or": []map[string]interface{}{
			{"used_at": nil},
			{"used_at": map[string]interface{}{"$exists": false}},
		},
	}

	// Find the token
	var token model.PasswordResetToken
	err := collection.FindOne(nil, filter).Decode(&token)
	if err != nil {
		util.Logger.Debugw("Password reset token not found", "error", err, "token", tokenStr)
		return model.PasswordResetToken{}
	}

	return token
}

// GetTokensByUserId retrieves all password reset tokens for a user from MongoDB
func (m PasswordResetMongoDbMapper) GetTokensByUserId(userId string) []model.PasswordResetToken {
	var tokens []model.PasswordResetToken

	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter
	filter := map[string]interface{}{
		"user_id": userId,
	}

	// Find all tokens for the user
	cursor, err := collection.Find(nil, filter)
	if err != nil {
		util.Logger.Errorw("Failed to get password reset tokens by user ID", "error", err, "userId", userId)
		return tokens
	}
	defer cursor.Close(nil)

	// Decode the results
	if err := cursor.All(nil, &tokens); err != nil {
		util.Logger.Errorw("Failed to decode password reset tokens", "error", err)
		return tokens
	}

	return tokens
}

// MarkTokenAsUsed marks a password reset token as used in MongoDB
func (m PasswordResetMongoDbMapper) MarkTokenAsUsed(tokenId string) error {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter and update
	filter := map[string]interface{}{"_id": tokenId}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"used_at":        util.GetCurrentTime(),
			"update_time":    util.GetCurrentTime(),
			"update_user_id": util.Convert2ObjectId(tokenId), // Using token ID as user placeholder if needed, or better passing userId
		},
	}

	// Update the token
	_, err := collection.UpdateOne(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to mark password reset token as used", "error", err, "token_id", tokenId)
		return err
	}

	return nil
}

// DeleteToken deletes a password reset token from MongoDB
func (m PasswordResetMongoDbMapper) DeleteToken(tokenId string) error {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter
	filter := map[string]interface{}{"_id": tokenId}

	// Soft delete
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_delete":   true,
			"delete_time": util.GetCurrentTime(),
		},
	}

	_, err := collection.UpdateOne(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to delete password reset token", "error", err, "token_id", tokenId)
		return err
	}

	return nil
}

// DeleteExpiredTokens deletes all expired password reset tokens from MongoDB
func (m PasswordResetMongoDbMapper) DeleteExpiredTokens() error {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter for expired tokens
	filter := map[string]interface{}{
		"expires_at": map[string]interface{}{"$lt": util.GetCurrentTime()},
		"is_delete":  false,
	}

	// Soft delete
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_delete":   true,
			"delete_time": util.GetCurrentTime(),
		},
	}

	_, err := collection.UpdateMany(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to delete expired password reset tokens", "error", err)
		return err
	}

	return nil
}

// DeleteTokensByUserId deletes all password reset tokens for a user from MongoDB
func (m PasswordResetMongoDbMapper) DeleteTokensByUserId(userId string) error {
	// Get database connection
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	// Define filter
	filter := map[string]interface{}{
		"user_id":   userId,
		"is_delete": false,
	}

	// Soft delete
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_delete":      true,
			"delete_time":    util.GetCurrentTime(),
			"delete_user_id": util.Convert2ObjectId(userId),
		},
	}

	_, err := collection.UpdateMany(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to delete password reset tokens for user", "error", err, "user_id", userId)
		return err
	}

	return nil
}

// InvalidateActiveTokensByUserId invalidates all active tokens for a user
func (m PasswordResetMongoDbMapper) InvalidateActiveTokensByUserId(userId string) error {
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	filter := map[string]interface{}{
		"user_id":    userId,
		"used_at":    nil,
		"is_delete":  false,
		"expires_at": map[string]interface{}{"$gt": util.GetCurrentTime()},
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_delete":      true,
			"delete_time":    util.GetCurrentTime(),
			"delete_user_id": util.Convert2ObjectId(userId),
		},
	}

	_, err := collection.UpdateMany(nil, filter, update)
	return err
}

// CountTokensByUserIdAndDateRange counts tokens for a user in a range
func (m PasswordResetMongoDbMapper) CountTokensByUserIdAndDateRange(userId string, from, to int64) (int64, error) {
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	filter := map[string]interface{}{
		"user_id": userId,
		"create_time": map[string]interface{}{
			"$gte": time.Unix(from, 0),
			"$lte": time.Unix(to, 0),
		},
	}

	return collection.CountDocuments(nil, filter)
}

// CountTokensByIPAndDateRange counts tokens for an IP in a range
func (m PasswordResetMongoDbMapper) CountTokensByIPAndDateRange(ipAddress string, from, to int64) (int64, error) {
	collection := database.GetMongoCollection(database.PasswordResetCollectionName)

	filter := map[string]interface{}{
		"ip_address": ipAddress,
		"create_time": map[string]interface{}{
			"$gte": time.Unix(from, 0),
			"$lte": time.Unix(to, 0),
		},
	}

	return collection.CountDocuments(nil, filter)
}
