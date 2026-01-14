package refresh_token_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var INSTANCE RefreshTokenMapper

// RefreshTokenMapper defines the interface for refresh token data access operations
type RefreshTokenMapper interface {
	// CreateToken creates a new refresh token
	CreateToken(token model.RefreshToken) string
	
	// GetTokenByToken retrieves a refresh token by its token string
	GetTokenByToken(token string) model.RefreshToken
	
	// RevokeToken revokes a refresh token by its token string
	RevokeToken(token string, revokedBy string) error
	
	// RevokeAllTokensByUserId revokes all refresh tokens for a user
	RevokeAllTokensByUserId(userId string) error
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = RefreshTokenMongoDbMapper{}
	case "mysql":
		INSTANCE = RefreshTokenMySqlMapper{}
	default:
		panic("database type not supported")
	}
}