package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserEntity represents a user in the database
type UserEntity struct {
	Id           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password_hash" json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	Role         string             `bson:"role" json:"role"` // Default: "user", can be "admin"
}

// IsEmpty checks if the user entity is empty
func (u UserEntity) IsEmpty() bool {
	return u.Id == primitive.NilObjectID && u.Username == ""
}

// UserDTO represents a user for API requests/responses
type UserDTO struct {
	Id        string `json:"id,omitempty"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"` // Only used for password creation/updates
	IsActive  bool   `json:"is_active,omitempty"`
	Role      string `json:"role,omitempty"`
	CreatedAt string `json:"created_at,omitempty"` // ISO formatted string for API
	UpdatedAt string `json:"updated_at,omitempty"` // ISO formatted string for API
}

// UserLoginRequest represents a login request
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLoginResponse represents a login response with JWT token
type UserLoginResponse struct {
	Token string    `json:"token"`
	User  UserEntity `json:"user"`
}

// UserRegistrationRequest represents a user registration request
type UserRegistrationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}