package model

import "time"

// RefreshToken represents a refresh token for users
type RefreshToken struct {
	Id            string     `bson:"_id,omitempty" json:"id"`
	BelongsUserId string     `bson:"belongs_user_id" json:"belongs_user_id"`
	Token         string     `bson:"token" json:"token"`
	ExpiresAt     time.Time  `bson:"expires_at" json:"expires_at"`
	CreatedAt     time.Time  `bson:"created_at" json:"created_at"`
	RevokedAt     *time.Time `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
	RevokedBy     string     `bson:"revoked_by,omitempty" json:"revoked_by,omitempty"`
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
