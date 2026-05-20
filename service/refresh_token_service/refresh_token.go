package refresh_token_service

import (
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

	// Validate device information
	// We primarily use User-Agent for validation as requested
	if userAgent != "" && refreshToken.UserAgent != "" && refreshToken.UserAgent != userAgent {
		return model.RefreshToken{}, errors.NewUnauthorizedError("device information does not match")
	}

	// We skip strict IP validation to allow network changes (e.g. wifi to cellular)
	// We skip DeviceID validation as it's not reliably provided by all clients

	return refreshToken, nil
}

// CreateRefreshToken creates a new refresh token for a user
func CreateRefreshToken(userID string, deviceID, deviceName, ipAddress, userAgent string) (string, error) {
	expDays := refreshTokenExpirationDays()

	userObjectId := util.Convert2ObjectId(userID)
	currentTime := util.GetCurrentTime()

	// Generate refresh token
	refreshToken := model.RefreshToken{
		BaseEntity: model.BaseEntity{
			CreateTime:   currentTime,
			CreateUserId: userObjectId,
			UpdateTime:   currentTime,
			UpdateUserId: userObjectId,
			IsDelete:     false,
		},
		Id:        util.GenerateUUID(),
		UserId:    userID,
		Token:     util.GenerateUUID(),
		ExpiresAt: time.Now().AddDate(0, 0, expDays),
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

func refreshTokenExpirationDays() int {
	expDays := util.GetConfigInt("auth.refresh_token.expiration_days", 14)
	if expDays <= 0 {
		util.Logger.Warnw("Invalid refresh token expiration days, using default", "value", expDays, "default", 14)
		return 14
	}
	return int(expDays)
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
