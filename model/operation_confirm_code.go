package model

import (
	"time"
)

// OperationConfirmCode represents a verification code for various operations
type OperationConfirmCode struct {
	Id            string             `json:"id" bson:"_id,omitempty"`
	UserId        string             `json:"user_id" bson:"user_id"`
	Code          string             `json:"code" bson:"code"`
	OperationType string             `json:"operation_type" bson:"operation_type"`
	Payload       string             `json:"payload" bson:"payload"`
	ExpiresTime   time.Time          `json:"expires_time" bson:"expires_time"`
	UsedTime      *time.Time         `json:"used_time" bson:"used_time"`
	BaseEntity    `bson:",inline"`
}

// Collection name for Mongo
func (OperationConfirmCode) CollectionName() string {
	return "operation_confirm_codes"
}
