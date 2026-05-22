package verification_service

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/email"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OperationType defines the type of verification operation
type OperationType string

const (
	OperationSignup        OperationType = "signup"
	OperationEmailChange   OperationType = "email_change"
	OperationPasswordReset OperationType = "password_reset"
)

const verificationCodeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// SendVerificationCode sends a purpose-scoped email code and stores it in the database.
func SendVerificationCode(purpose string, recipientEmail string, ipAddress string) error {
	operation, err := parseOperation(purpose)
	if err != nil {
		return err
	}

	recipientEmail = strings.ToLower(strings.TrimSpace(recipientEmail))
	if err := validation.ValidateEmail(recipientEmail); err != nil {
		return err
	}

	existingEmailUser := user_mapper.INSTANCE.GetUserByEmail(recipientEmail)
	switch operation {
	case OperationPasswordReset:
		if existingEmailUser.Id.IsZero() {
			return nil
		}
	case OperationSignup, OperationEmailChange:
		if !existingEmailUser.Id.IsZero() {
			return errors.NewFieldAlreadyExistsError("email", "email address already exists")
		}
	}

	if !email.GetService().IsConfigured() {
		return errors.NewInternalError("SMTP service is not configured", nil)
	}

	if err := email.CheckAndRecordPurposeEmailAllowance(string(operation), ipAddress, []string{recipientEmail}); err != nil {
		return err
	}

	if err := enforceSendInterval(operation, recipientEmail); err != nil {
		return err
	}

	if err := codeRepo.InvalidateActiveCodesByPurposeAndPayload(string(operation), recipientEmail); err != nil {
		util.Logger.Errorw("Failed to invalidate prior verification codes", "error", err, "purpose", operation, "email", recipientEmail)
	}

	code, err := generateReadableCode(6)
	if err != nil {
		return errors.NewInternalError("failed to generate verification code", err)
	}

	currentTime := util.GetCurrentTime()
	expiresAt := currentTime.Add(time.Duration(verificationExpireMinutes()) * time.Minute)
	entity := model.OperationConfirmCode{
		BaseEntity: model.BaseEntity{
			CreateTime: currentTime,
			UpdateTime: currentTime,
			IsDelete:   false,
		},
		Id:            primitive.NewObjectID().Hex(),
		Code:          code,
		OperationType: string(operation),
		Payload:       recipientEmail,
		ExpiresTime:   expiresAt,
	}

	if err := codeRepo.CreateCode(entity); err != nil {
		return errors.NewInternalError("failed to store verification code", err)
	}

	subject := "CashLenX verification code"
	body := fmt.Sprintf("Your CashLenX verification code is: %s\n\nThis code will expire in %s.", code, verificationExpiryText())
	if err := email.GetService().SendEmail([]string{recipientEmail}, subject, body, false); err != nil {
		codeRepo.DeleteCode(entity.Id)
		return errors.NewInternalError("failed to send verification email", err)
	}

	return nil
}

// VerifyCode checks a code and returns a one-time token valid for the same expiry window.
func VerifyCode(purpose string, recipientEmail string, code string) (model.VerifyVerificationCodeResponse, error) {
	operation, err := parseOperation(purpose)
	if err != nil {
		return model.VerifyVerificationCodeResponse{}, err
	}

	recipientEmail = strings.ToLower(strings.TrimSpace(recipientEmail))
	if err := validation.ValidateEmail(recipientEmail); err != nil {
		return model.VerifyVerificationCodeResponse{}, err
	}

	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) != 6 {
		return model.VerifyVerificationCodeResponse{}, errors.NewInvalidInputError("verification code must be 6 characters")
	}

	stored := codeRepo.GetActiveCodeByPurposeAndPayload(string(operation), recipientEmail)
	if stored.Id == "" || stored.Code != code {
		return model.VerifyVerificationCodeResponse{}, errors.NewInvalidInputError("invalid verification code")
	}
	if !isCodeUsable(stored) {
		return model.VerifyVerificationCodeResponse{}, errors.NewInvalidInputError("invalid or expired verification code")
	}

	token, err := generateHexToken(32)
	if err != nil {
		return model.VerifyVerificationCodeResponse{}, errors.NewInternalError("failed to generate verification token", err)
	}
	if err := codeRepo.SetVerificationToken(stored.Id, token); err != nil {
		return model.VerifyVerificationCodeResponse{}, errors.NewInternalError("failed to store verification token", err)
	}

	return model.VerifyVerificationCodeResponse{
		Token:     token,
		ExpiresAt: stored.ExpiresTime,
	}, nil
}

// ConsumeVerifiedToken consumes a one-time token returned from VerifyCode.
func ConsumeVerifiedToken(token string, operation OperationType) (model.OperationConfirmCode, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return model.OperationConfirmCode{}, errors.NewInvalidInputError("verification token is required")
	}

	code := codeRepo.GetCodeByVerificationToken(token)
	if code.Id == "" || code.OperationType != string(operation) || code.VerificationToken != token {
		return model.OperationConfirmCode{}, errors.NewInvalidInputError("invalid verification token")
	}
	if !isCodeUsable(code) {
		return model.OperationConfirmCode{}, errors.NewInvalidInputError("invalid or expired verification token")
	}
	if err := codeRepo.MarkCodeAsUsed(code.Id); err != nil {
		return model.OperationConfirmCode{}, errors.NewInternalError("failed to consume verification token", err)
	}

	return code, nil
}

// GenerateVerificationToken generates a random token for a specific operation
func GenerateVerificationToken(userId string, operation OperationType, payload string) (string, error) {
	// Revoke previous tokens for this user and operation
	// This ensures only one valid token exists at a time for a specific flow
	err := codeRepo.InvalidateActiveCodes(userId, string(operation))
	if err != nil {
		util.Logger.Errorw("Failed to invalidate active tokens", "error", err)
		// Continue even if invalidation fails? Ideally yes, but let's log it.
	}

	token, err := generateHexToken(32)
	if err != nil {
		return "", err
	}

	// Create entity
	userObjectId := util.Convert2ObjectId(userId)
	currentTime := util.GetCurrentTime()

	code := model.OperationConfirmCode{
		BaseEntity: model.BaseEntity{
			CreateTime:   currentTime,
			CreateUserId: userObjectId,
			UpdateTime:   currentTime,
			UpdateUserId: userObjectId,
			IsDelete:     false,
		},
		Id:                primitive.NewObjectID().Hex(),
		UserId:            userId,
		Code:              token,
		VerificationToken: token,
		OperationType:     string(operation),
		Payload:           payload,
		ExpiresTime:       time.Now().Add(time.Duration(verificationExpireMinutes()) * time.Minute),
		UsedTime:          nil,
	}

	// Store in database
	err = codeRepo.CreateCode(code)
	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyToken checks if a token is valid for a specific operation and returns the payload and userId
func VerifyToken(token string, operation OperationType) (string, string, bool) {
	code := codeRepo.GetCodeByVerificationToken(token)
	if code.Id == "" {
		code = codeRepo.GetCodeByToken(token)
	}

	if code.Id == "" {
		return "", "", false
	}

	if code.OperationType != string(operation) {
		return "", "", false
	}

	if time.Now().After(code.ExpiresTime) {
		// Logically delete expired token for cleanup
		codeRepo.DeleteCode(code.Id)
		return "", "", false
	}

	if code.UsedTime != nil {
		return "", "", false
	}

	return code.Payload, code.UserId, true
}

// InvalidateToken marks a token as used
func InvalidateToken(token string) {
	code := codeRepo.GetCodeByVerificationToken(token)
	if code.Id == "" {
		code = codeRepo.GetCodeByToken(token)
	}
	if code.Id != "" {
		codeRepo.MarkCodeAsUsed(code.Id)
	}
}

// InvalidateTokensByUserAndOperation removes all tokens for a specific user and operation
func InvalidateTokensByUserAndOperation(userId string, operation OperationType) {
	codeRepo.InvalidateActiveCodes(userId, string(operation))
}

func parseOperation(purpose string) (OperationType, error) {
	switch OperationType(strings.TrimSpace(purpose)) {
	case OperationSignup:
		return OperationSignup, nil
	case OperationPasswordReset:
		return OperationPasswordReset, nil
	case OperationEmailChange:
		return OperationEmailChange, nil
	default:
		return "", errors.NewInvalidInputError("invalid verification purpose")
	}
}

func enforceSendInterval(operation OperationType, recipientEmail string) error {
	intervalSeconds := int(util.GetConfigInt("verification.code.send_interval_seconds", 60))
	if intervalSeconds <= 0 {
		return nil
	}

	active := codeRepo.GetActiveCodeByPurposeAndPayload(string(operation), recipientEmail)
	if active.Id == "" || active.UsedTime != nil {
		return nil
	}
	if time.Now().After(active.ExpiresTime) {
		codeRepo.DeleteCode(active.Id)
		return nil
	}
	if time.Since(active.CreateTime) < time.Duration(intervalSeconds)*time.Second {
		return errors.NewRateLimitedError("verification code was sent recently; try again later")
	}
	return nil
}

func generateReadableCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := randomRead(bytes); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.Grow(length)
	for _, value := range bytes {
		builder.WriteByte(verificationCodeAlphabet[int(value)%len(verificationCodeAlphabet)])
	}
	return builder.String(), nil
}

func generateHexToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := randomRead(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func verificationExpireMinutes() int {
	minutes := int(util.GetConfigInt("verification.code.expire_minutes", 30))
	if minutes <= 0 {
		return 30
	}
	return minutes
}

func verificationExpiryText() string {
	minutes := verificationExpireMinutes()
	if minutes == 1 {
		return "1 minute"
	}
	if minutes%60 == 0 {
		hours := minutes / 60
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func isCodeUsable(code model.OperationConfirmCode) bool {
	if code.Id == "" || code.UsedTime != nil {
		return false
	}
	if time.Now().After(code.ExpiresTime) {
		codeRepo.DeleteCode(code.Id)
		return false
	}
	return true
}
