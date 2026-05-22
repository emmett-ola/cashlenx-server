package model

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	DefaultDisplayLanguage = "en"
	DefaultCurrencyCode    = "USD"
	DefaultThemeColor      = "#2563eb"
)

// UserConfigurationEntity stores user-specific application preferences.
type UserConfigurationEntity struct {
	Id               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	BelongsUserId    primitive.ObjectID `json:"belongs_user_id" bson:"belongs_user_id"`
	DisplayLanguage  string             `json:"display_language" bson:"display_language"`
	CurrencyCode     string             `json:"currency_code" bson:"currency_code"`
	ActiveThemeColor string             `json:"active_theme_color" bson:"active_theme_color"`
	BaseEntity       `bson:",inline"`
}

// UserConfigurationRequest represents a create/update request for user preferences.
type UserConfigurationRequest struct {
	DisplayLanguage  *string `json:"display_language,omitempty"`
	CurrencyCode     *string `json:"currency_code,omitempty"`
	ActiveThemeColor *string `json:"active_theme_color,omitempty"`
}

func DefaultUserConfiguration(userId primitive.ObjectID) UserConfigurationEntity {
	return UserConfigurationEntity{
		BelongsUserId:    userId,
		DisplayLanguage:  DefaultDisplayLanguage,
		CurrencyCode:     DefaultCurrencyCode,
		ActiveThemeColor: DefaultThemeColor,
	}
}
