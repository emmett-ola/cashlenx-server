package user_service

import (
	"strings"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	maxDisplayLanguageLength = 20
	currencyCodeLength       = 3
	maxThemeColorLength      = 50
)

func GetConfigurationService(plainUserId string) (model.UserConfigurationEntity, error) {
	user := userRepo.GetUserByObjectId(plainUserId)
	if user.Id.IsZero() {
		return model.UserConfigurationEntity{}, errors.NewNotFoundError("user not found")
	}

	config := userConfigurationRepo.GetByUserId(plainUserId)
	if config.Id.IsZero() {
		return model.DefaultUserConfiguration(user.Id), nil
	}

	return config, nil
}

func UpsertConfigurationService(plainUserId string, requestBody model.UserConfigurationRequest) (model.UserConfigurationEntity, error) {
	user := userRepo.GetUserByObjectId(plainUserId)
	if user.Id.IsZero() {
		return model.UserConfigurationEntity{}, errors.NewNotFoundError("user not found")
	}

	existingConfig := userConfigurationRepo.GetByUserId(plainUserId)
	config := existingConfig
	now := util.GetCurrentTime()
	if config.Id.IsZero() {
		config = model.DefaultUserConfiguration(user.Id)
		config.Id = primitive.NewObjectID()
		config.CreateUserId = user.Id
		config.CreateTime = now
	}
	config.UpdateUserId = user.Id
	config.UpdateTime = now

	if err := applyUserConfigurationRequest(&config, requestBody); err != nil {
		return model.UserConfigurationEntity{}, err
	}

	if existingConfig.Id.IsZero() {
		insertedId := userConfigurationRepo.InsertByEntity(config)
		if insertedId == "" {
			return model.UserConfigurationEntity{}, errors.NewInternalError("failed to create user configuration", nil)
		}
		return config, nil
	}

	updatedConfig := userConfigurationRepo.UpdateByUserId(plainUserId, config)
	if updatedConfig.Id.IsZero() {
		return model.UserConfigurationEntity{}, errors.NewInternalError("failed to update user configuration", nil)
	}

	return updatedConfig, nil
}

func applyUserConfigurationRequest(config *model.UserConfigurationEntity, requestBody model.UserConfigurationRequest) error {
	if requestBody.DisplayLanguage != nil {
		value := strings.TrimSpace(*requestBody.DisplayLanguage)
		if value == "" || len(value) > maxDisplayLanguageLength {
			return errors.NewFieldValidationError("display_language", "display_language must be between 1 and 20 characters")
		}
		config.DisplayLanguage = value
	}

	if requestBody.CurrencyCode != nil {
		value := strings.ToUpper(strings.TrimSpace(*requestBody.CurrencyCode))
		if !isCurrencyCode(value) {
			return errors.NewFieldValidationError("currency_code", "currency_code must be a 3-letter ISO 4217 currency code")
		}
		config.CurrencyCode = value
	}

	if requestBody.ActiveThemeColor != nil {
		value := strings.TrimSpace(*requestBody.ActiveThemeColor)
		if value == "" || len(value) > maxThemeColorLength {
			return errors.NewFieldValidationError("active_theme_color", "active_theme_color must be between 1 and 50 characters")
		}
		config.ActiveThemeColor = value
	}

	return nil
}

func isCurrencyCode(value string) bool {
	if len(value) != currencyCodeLength {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
