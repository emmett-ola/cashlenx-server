package model

import "time"

// RefreshToken represents a refresh token for users
type RefreshToken struct {
	BaseEntity `bson:",inline"`
	Id         string     `bson:"_id,omitempty" json:"id"`
	UserId     string     `bson:"user_id" json:"user_id"`
	Token      string     `bson:"token" json:"token"`
	ExpiresAt  time.Time  `bson:"expires_at" json:"expires_at"`
	RevokedAt  *time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
	RevokedBy  string     `bson:"revoked_by,omitempty" json:"revoked_by,omitempty"`
	// Device information
	DeviceId   string `bson:"device_id,omitempty" json:"device_id,omitempty"`
	DeviceName string `bson:"device_name,omitempty" json:"device_name,omitempty"`
	IPAddress  string `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	UserAgent  string `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
}

// RefreshTokenRequest represents a request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents a response with new access and refresh tokens
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         UserEntity `json:"user"`
}
