package password_reset_mapper

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var INSTANCE PasswordResetMapper

// PasswordResetMapper defines operations for managing password reset tokens
type PasswordResetMapper interface {
	// CreateToken creates a new password reset token
	CreateToken(token model.PasswordResetToken) string
	
	// GetTokenByToken retrieves a password reset token by its token string
	GetTokenByToken(token string) model.PasswordResetToken
	
	// GetTokensByUserId retrieves all password reset tokens for a user
	GetTokensByUserId(userId string) []model.PasswordResetToken
	
	// MarkTokenAsUsed marks a password reset token as used
	MarkTokenAsUsed(tokenId string) error
	
	// DeleteToken deletes a password reset token
	DeleteToken(tokenId string) error
	
	// DeleteExpiredTokens deletes all expired password reset tokens
	DeleteExpiredTokens() error
	
	// DeleteTokensByUserId deletes all password reset tokens for a user
	DeleteTokensByUserId(userId string) error
}

func init() {
	switch util.GetConfigByKey("db.type") {
	case "mongodb":
		INSTANCE = PasswordResetMongoDbMapper{}
	case "mysql":
		INSTANCE = PasswordResetMySqlMapper{}
	default:
		panic("database type not supported")
	}
}
