package model

import "time"

// PasswordResetToken represents a password reset token for users
type PasswordResetToken struct {
	Id            string     `bson:"_id,omitempty" json:"id"`
	BelongsUserId string     `bson:"belongs_user_id" json:"belongs_user_id"`
	Token         string     `bson:"token" json:"token"`
	ExpiresAt     time.Time  `bson:"expires_at" json:"expires_at"`
	CreatedAt     time.Time  `bson:"created_at" json:"created_at"`
	UsedAt        *time.Time `bson:"used_at,omitempty" json:"used_at,omitempty"`
}

// PasswordResetRequest represents a request to reset a password
type PasswordResetRequest struct {
	EmailOrUsername string `json:"email_or_username"`
}

// PasswordResetConfirmRequest represents a request to confirm a password reset
type PasswordResetConfirmRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
