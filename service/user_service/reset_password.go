package user_service

import (
	"fmt"
	"strconv"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/email"
	"github.com/macar-x/cashlenx-server/validation"
	"golang.org/x/crypto/bcrypt"
)

// RequestPasswordReset initiates a password reset request
func RequestPasswordReset(emailOrUsername string, ipAddress string) error {
	// 1. Get user by username or email
	var user model.UserEntity
	// Try username first
	user = GetUserByUsername(emailOrUsername)
	if user.Id.IsZero() {
		// Try email
		user = userRepo.GetUserByEmail(emailOrUsername)
	}

	if user.Id.IsZero() {
		// If user not found, return nil (security best practice: do not reveal user existence)
		return nil
	}

	userId := user.Id.Hex()

	if user.EmailAddress == "" {
		util.Logger.Warnw("User has no email address, cannot send reset token", "userId", userId)
		return errors.NewInvalidInputError("user has no email address configured")
	}

	if !email.GetService().IsConfigured() {
		util.Logger.Errorw("SMTP service is not configured, cannot send reset token", "userId", userId)
		return errors.NewInternalError("SMTP service is not configured", nil)
	}

	if err := email.CheckAndRecordPurposeEmailAllowance(
		string(verification_service.OperationPasswordReset),
		ipAddress,
		[]string{user.EmailAddress},
	); err != nil {
		return err
	}

	// 2. Invalidate existing active tokens for this operation.
	verification_service.InvalidateTokensByUserAndOperation(userId, verification_service.OperationPasswordReset)

	// 3. Generate verification token.
	token, err := verification_service.GenerateVerificationToken(userId, verification_service.OperationPasswordReset, "")
	if err != nil {
		util.Logger.Errorw("Failed to generate verification token", "error", err)
		return errors.NewInternalError("failed to generate verification token", err)
	}

	// 4. Send email.
	subject := "Password Reset Request - CashLenX"
	body := fmt.Sprintf(`Hello %s,

We received a request to reset your password for your CashLenX account.
Your password reset token is: %s

This token will expire in %s.

If you did not request a password reset, please ignore this email.

Best regards,
The CashLenX Team`, user.Username, token, verificationCodeExpiryText())

	err = email.GetService().SendEmail([]string{user.EmailAddress}, subject, body, false)
	if err != nil {
		util.Logger.Errorw("Failed to send password reset email", "error", err, "userId", userId, "email", user.EmailAddress)
		verification_service.InvalidateToken(token)
		return errors.NewInternalError("failed to send password reset email", err)
	}

	return nil
}

func verificationCodeExpiryText() string {
	minutes := int(util.GetConfigInt("verification.code.expire_minutes", 30))
	if minutes <= 0 {
		minutes = 30
	}
	if minutes == 1 {
		return "1 minute"
	}
	if minutes%60 == 0 {
		hours := minutes / 60
		if hours == 1 {
			return "1 hour"
		}
		return strconv.Itoa(hours) + " hours"
	}
	return strconv.Itoa(minutes) + " minutes"
}

// ConfirmPasswordReset confirms a password reset using a token
func ConfirmPasswordReset(token string, newPassword string) error {
	// Validate password
	err := validation.ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	// Verify token using verification service
	// This replaces the database lookup
	_, userId, valid := verification_service.VerifyToken(token, verification_service.OperationPasswordReset)
	if !valid {
		return errors.NewInvalidInputError("invalid or expired password reset token")
	}

	// Get user
	user := GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return errors.NewNotFoundError("user not found")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.NewInternalError("failed to hash password", err)
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	user.UpdateTime = util.GetCurrentTime()

	// Save updated user
	updatedUser := userRepo.UpdateUserByEntity(user.Id.Hex(), user)
	if updatedUser.Id.IsZero() {
		return errors.NewInternalError("failed to update password", nil)
	}

	// Mark token as used/invalidate it
	verification_service.InvalidateToken(token)

	// Note: We are no longer deleting tokens from DB since we use in-memory store

	return nil
}
