package refresh_token_service

import (
	"strconv"
	"time"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/refresh_token_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

// GetRefreshTokenByToken retrieves a refresh token by its token string
func GetRefreshTokenByToken(token string) (model.RefreshToken, error) {
	refreshToken := refresh_token_mapper.INSTANCE.GetTokenByToken(token)
	if refreshToken.Id == "" {
		return model.RefreshToken{}, errors.NewUnauthorizedError("invalid or expired refresh token")
	}
	return refreshToken, nil
}

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(userID string) (string, error) {
	// Get refresh token expiration seconds from configuration
	expSecondsStr := util.GetConfigByKey("auth.refresh_token.expiration_seconds")
	expSeconds := 43200 // Default to 12 hours (43200 seconds)
	if expSecondsStr != "" {
		if parsedSeconds, err := strconv.Atoi(expSecondsStr); err == nil {
			expSeconds = parsedSeconds
		}
	}

	// Generate refresh token
	refreshToken := model.RefreshToken{
		Id:        util.GenerateUUID(),
		UserId:    userID,
		Token:     util.GenerateUUID(),
		ExpiresAt: time.Now().Add(time.Duration(expSeconds) * time.Second),
		CreatedAt: time.Now(),
	}

	createdToken := refresh_token_mapper.INSTANCE.CreateToken(refreshToken)
	if createdToken == "" {
		return "", errors.NewInternalError("failed to create refresh token", nil)
	}

	return createdToken, nil
}

// RevokeRefreshToken revokes a refresh token
func RevokeRefreshToken(token string, revokedBy string) error {
	err := refresh_token_mapper.INSTANCE.RevokeToken(token, revokedBy)
	if err != nil {
		return errors.NewUnauthorizedError("invalid refresh token")
	}
	return nil
}

// RevokeAllRefreshTokens revokes all refresh tokens for a user
func RevokeAllRefreshTokens(userID string) error {
	err := refresh_token_mapper.INSTANCE.RevokeAllTokensByUserId(userID)
	if err != nil {
		return errors.NewInternalError("failed to revoke refresh tokens", nil)
	}
	return nil
}
