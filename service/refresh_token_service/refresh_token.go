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
func GetRefreshTokenByToken(token string, deviceID, deviceName, ipAddress, userAgent string) (model.RefreshToken, error) {
	refreshToken := refresh_token_mapper.INSTANCE.GetTokenByToken(token)
	if refreshToken.Id == "" {
		return model.RefreshToken{}, errors.NewUnauthorizedError("invalid or expired refresh token")
	}

	// Validate device information if provided
	if deviceID != "" && refreshToken.DeviceId != "" && refreshToken.DeviceId != deviceID {
		return model.RefreshToken{}, errors.NewUnauthorizedError("device information does not match")
	}
	if deviceName != "" && refreshToken.DeviceName != "" && refreshToken.DeviceName != deviceName {
		return model.RefreshToken{}, errors.NewUnauthorizedError("device information does not match")
	}
	if ipAddress != "" && refreshToken.IPAddress != "" && refreshToken.IPAddress != ipAddress {
		return model.RefreshToken{}, errors.NewUnauthorizedError("device information does not match")
	}
	if userAgent != "" && refreshToken.UserAgent != "" && refreshToken.UserAgent != userAgent {
		return model.RefreshToken{}, errors.NewUnauthorizedError("device information does not match")
	}

	return refreshToken, nil
}

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(userID string, deviceID, deviceName, ipAddress, userAgent string) (string, error) {
	// Get refresh token expiration days from configuration
	expDaysStr := util.GetConfigByKey("auth.refresh_token.expiration_days")
	expDays := 30 // Default to 30 days
	if expDaysStr != "" {
		if parsedDays, err := strconv.Atoi(expDaysStr); err == nil {
			expDays = parsedDays
		}
	}

	// Generate refresh token
	refreshToken := model.RefreshToken{
		Id:        util.GenerateUUID(),
		UserId:    userID,
		Token:     util.GenerateUUID(),
		ExpiresAt: time.Now().AddDate(0, 0, expDays),
		CreatedAt: time.Now(),
		// Device information
		DeviceId:   deviceID,
		DeviceName: deviceName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
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

// GetUserRefreshTokens retrieves all refresh tokens for a user
func GetUserRefreshTokens(userID string) []model.RefreshToken {
	return refresh_token_mapper.INSTANCE.GetTokensByUserId(userID)
}
