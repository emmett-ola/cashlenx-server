package model

import (
	"reflect"
	"strconv"

	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserEntity represents a user in the database
type UserEntity struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	IsActive     bool               `json:"is_active" bson:"is_active"`
	Role         string             `json:"role" bson:"role"`
	Nickname     string             `json:"nickname,omitempty" bson:"nickname,omitempty"`
	AvatarUrl    string             `json:"avatar_url,omitempty" bson:"avatar_url,omitempty"`
	EmailAddress string             `json:"email_address,omitempty" bson:"email_address,omitempty"`
	Gender       string             `json:"gender,omitempty" bson:"gender,omitempty"`
	BaseEntity   `bson:",inline"`
}

func (entity UserEntity) IsEmpty() bool {
	return reflect.DeepEqual(entity, UserEntity{})
}

func (entity UserEntity) ToString() string {
	return "[ " +
		"Id: " + entity.Id.Hex() +
		", Username: " + entity.Username +
		", Role: " + entity.Role +
		", IsActive: " + strconv.FormatBool(entity.IsActive) +
		", CreateTime: " + util.FormatDateToStringWithDash(entity.CreateTime) +
		" ]"
}

func (entity UserEntity) Build(fieldMap map[string]string) UserEntity {
	newEntity := entity
	for key, value := range fieldMap {
		switch key {
		case "Id":
			objectId, err := primitive.ObjectIDFromHex(value)
			if err != nil {
				util.Logger.Warnln("build user failed with err: " + err.Error())
			}
			newEntity.Id = objectId
		case "Username":
			newEntity.Username = value
		case "Role":
			newEntity.Role = value
		case "IsActive":
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				util.Logger.Warnln("build user failed with err: " + err.Error())
			}
			newEntity.IsActive = boolValue
		}
	}
	return newEntity
}

// UserDTO represents a user for API requests/responses
type UserDTO struct {
	Id           string `json:"id,omitempty"`
	Username     string `json:"username"`
	Password     string `json:"password,omitempty"` // Only used for password creation/updates
	IsActive     bool   `json:"is_active,omitempty"`
	Role         string `json:"role,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	AvatarUrl    string `json:"avatar_url,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
	Gender       string `json:"gender,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"` // ISO formatted string for API
	UpdatedAt    string `json:"updated_at,omitempty"` // ISO formatted string for API
}

// UserLoginRequest represents a login request
type UserLoginRequest struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
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