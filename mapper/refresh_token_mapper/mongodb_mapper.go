package refresh_token_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
)

// RefreshTokenMongoDbMapper implements RefreshTokenMapper for MongoDB
type RefreshTokenMongoDbMapper struct{}

// CreateToken creates a new refresh token in MongoDB
func (m RefreshTokenMongoDbMapper) CreateToken(token model.RefreshToken) string {
	// Get database connection
	collection := database.GetMongoCollection(database.RefreshTokenCollectionName)

	// Insert the token into the collection
	_, err := collection.InsertOne(nil, token)
	if err != nil {
		util.Logger.Errorw("Failed to create refresh token", "error", err, "user_id", token.UserId)
		return ""
	}

	return token.Token
}

// GetTokenByToken retrieves a refresh token by its token string from MongoDB
func (m RefreshTokenMongoDbMapper) GetTokenByToken(tokenStr string) model.RefreshToken {
	// Get database connection
	collection := database.GetMongoCollection(database.RefreshTokenCollectionName)

	// Define filter
	filter := map[string]interface{}{
		"token": tokenStr,
		"$or": []map[string]interface{}{
			{"revoked_at": nil},
			{"revoked_at": map[string]interface{}{"$gt": util.GetCurrentTime()}},
		},
	}

	// Find the token
	var token model.RefreshToken
	err := collection.FindOne(nil, filter).Decode(&token)
	if err != nil {
		util.Logger.Debugw("Refresh token not found", "error", err, "token", tokenStr)
		return model.RefreshToken{}
	}

	return token
}

// RevokeToken revokes a refresh token by its token string in MongoDB
func (m RefreshTokenMongoDbMapper) RevokeToken(tokenStr string, revokedBy string) error {
	// Get database connection
	collection := database.GetMongoCollection(database.RefreshTokenCollectionName)

	// Define filter and update
	filter := map[string]interface{}{"token": tokenStr}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"revoked_at": util.GetCurrentTime(),
			"revoked_by": revokedBy,
		},
	}

	// Update the token
	_, err := collection.UpdateOne(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to revoke refresh token", "error", err, "token", tokenStr)
		return err
	}

	return nil
}

// RevokeAllTokensByUserId revokes all refresh tokens for a user in MongoDB
func (m RefreshTokenMongoDbMapper) RevokeAllTokensByUserId(userId string) error {
	// Get database connection
	collection := database.GetMongoCollection(database.RefreshTokenCollectionName)

	// Define filter and update
	filter := map[string]interface{}{
		"user_id":    userId,
		"revoked_at": nil,
	}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"revoked_at": util.GetCurrentTime(),
			"revoked_by": userId,
		},
	}

	// Update all matching tokens
	_, err := collection.UpdateMany(nil, filter, update)
	if err != nil {
		util.Logger.Errorw("Failed to revoke all refresh tokens for user", "error", err, "user_id", userId)
		return err
	}

	return nil
}

	// GetTokensByUserId retrieves all refresh tokens for a user
func (m RefreshTokenMongoDbMapper) GetTokensByUserId(userId string) []model.RefreshToken {
	var tokens []model.RefreshToken
	
	// Get database connection
	collection := database.GetMongoCollection(database.RefreshTokenCollectionName)
	
	// Define filter
	filter := map[string]interface{}{
		"user_id": userId,
	}
	
	// Find all tokens for the user
	cursor, err := collection.Find(nil, filter)
	if err != nil {
		util.Logger.Errorw("Failed to get refresh tokens by user ID", "error", err, "userId", userId)
		return tokens
	}
	defer cursor.Close(nil)
	
	// Decode the results
	if err := cursor.All(nil, &tokens); err != nil {
		util.Logger.Errorw("Failed to decode refresh tokens", "error", err)
		return tokens
	}
	
	return tokens
}