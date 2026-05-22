package model

import (
	"time"
)

// OperationConfirmCode represents a verification code for various operations
type OperationConfirmCode struct {
	Id                string     `json:"id" bson:"_id,omitempty"`
	UserId            string     `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Code              string     `json:"code" bson:"code"`
	VerificationToken string     `json:"verification_token,omitempty" bson:"verification_token,omitempty"`
	OperationType     string     `json:"operation_type" bson:"operation_type"`
	Payload           string     `json:"payload" bson:"payload"`
	ExpiresTime       time.Time  `json:"expires_time" bson:"expires_time"`
	UsedTime          *time.Time `json:"used_time" bson:"used_time"`
	BaseEntity        `bson:",inline"`
}

// Collection name for Mongo
func (OperationConfirmCode) CollectionName() string {
	return "operation_confirm_codes"
}

const (
	VerificationPurposeSignup        = "signup"
	VerificationPurposePasswordReset = "password_reset"
	VerificationPurposeEmailChange   = "email_change"
)

type SendVerificationCodeRequest struct {
	Purpose string `json:"purpose"`
	Email   string `json:"email"`
}

type VerifyVerificationCodeRequest struct {
	Purpose string `json:"purpose"`
	Email   string `json:"email"`
	Code    string `json:"code"`
}

type VerifyVerificationCodeResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
