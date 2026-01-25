package verification_service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/verification_code_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OperationType defines the type of verification operation
type OperationType string

const (
	OperationEmailChange   OperationType = "email_change"
	OperationPasswordReset OperationType = "password_reset"
)

// GenerateVerificationToken generates a random token for a specific operation
func GenerateVerificationToken(userId string, operation OperationType, payload string) (string, error) {
	// Revoke previous tokens for this user and operation
	// This ensures only one valid token exists at a time for a specific flow
	err := verification_code_mapper.INSTANCE.InvalidateActiveCodes(userId, string(operation))
	if err != nil {
		util.Logger.Errorw("Failed to invalidate active tokens", "error", err)
		// Continue even if invalidation fails? Ideally yes, but let's log it.
	}

	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	// Create entity
	userObjectId := util.Convert2ObjectId(userId)
	currentTime := util.GetCurrentTime()
	
	expireMinutes := util.GetConfigInt("verification.code.expire_minutes", 30)

	code := model.OperationConfirmCode{
		BaseEntity: model.BaseEntity{
			CreateTime:   currentTime,
			CreateUserId: userObjectId,
			UpdateTime:   currentTime,
			UpdateUserId: userObjectId,
			IsDelete:     false,
		},
		Id:            primitive.NewObjectID().Hex(),
		UserId:        userId,
		Code:          token,
		OperationType: string(operation),
		Payload:       payload,
		ExpiresTime:   time.Now().Add(time.Duration(expireMinutes) * time.Minute),
		UsedTime:      nil,
	}

	// Store in database
	err = verification_code_mapper.INSTANCE.CreateCode(code)
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyToken checks if a token is valid for a specific operation and returns the payload and userId
func VerifyToken(token string, operation OperationType) (string, string, bool) {
	code := verification_code_mapper.INSTANCE.GetCodeByToken(token)
	
	if code.Id == "" {
		return "", "", false
	}

	if code.OperationType != string(operation) {
		return "", "", false
	}

	if time.Now().After(code.ExpiresTime) {
		// Logically delete expired token for cleanup
		verification_code_mapper.INSTANCE.DeleteCode(code.Id)
		return "", "", false
	}

	if code.UsedTime != nil {
		return "", "", false
	}

	return code.Payload, code.UserId, true
}

// InvalidateToken marks a token as used
func InvalidateToken(token string) {
	code := verification_code_mapper.INSTANCE.GetCodeByToken(token)
	if code.Id != "" {
		verification_code_mapper.INSTANCE.MarkCodeAsUsed(code.Id)
	}
}

// InvalidateTokensByUserAndOperation removes all tokens for a specific user and operation
func InvalidateTokensByUserAndOperation(userId string, operation OperationType) {
	verification_code_mapper.INSTANCE.InvalidateActiveCodes(userId, string(operation))
}
