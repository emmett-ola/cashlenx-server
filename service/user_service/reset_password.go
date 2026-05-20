package user_service

import (
	"fmt"

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

	// 2. Check rate limits (Optional: Implement in-memory rate limiting if needed)
	// For now, relying on basic verification service which is simpler than DB-based tracking
	// To fully replicate the previous logic, we'd need to add tracking to verification_service or keep using DB for logs
	// Given the move to in-memory verification codes, we'll skip complex historical rate limiting for this iteration
	// or assume the verification service's ephemeral nature is acceptable.

	// 3. Invalidate existing active tokens for this operation
	// We use the new verification service to invalidate old tokens in memory
	verification_service.InvalidateTokensByUserAndOperation(userId, verification_service.OperationPasswordReset)

	// 4. Generate verification token
	// We use the verification service instead of database storage for the token
	// Payload is empty as we don't need to store extra data for password reset
	token, err := verification_service.GenerateVerificationToken(userId, verification_service.OperationPasswordReset, "")
	if err != nil {
		util.Logger.Errorw("Failed to generate verification token", "error", err)
		return errors.NewInternalError("failed to generate verification token", err)
	}

	// 5. Send email
	if user.EmailAddress != "" {
		// Check if SMTP is configured
		if !email.GetService().IsConfigured() {
			util.Logger.Errorw("SMTP service is not configured, cannot send reset token", "userId", userId)
			return errors.NewInternalError("SMTP service is not configured", nil)
		}

		// Construct email body
		subject := "Password Reset Request - CashLenX"
		body := fmt.Sprintf(`Hello %s,

We received a request to reset your password for your CashLenX account.
Your password reset token is: %s

This token will expire in 1 hour.

If you did not request a password reset, please ignore this email.

Best regards,
The CashLenX Team`, user.Username, token)

		// Send email synchronously to catch errors
		err := email.GetService().SendEmail([]string{user.EmailAddress}, subject, body, false)
		if err != nil {
			util.Logger.Errorw("Failed to send password reset email", "error", err, "userId", userId, "email", user.EmailAddress)
			// Invalidate the token we just generated since email failed
			verification_service.InvalidateToken(token)
			return errors.NewInternalError("failed to send password reset email", err)
		}
	} else {
		util.Logger.Warnw("User has no email address, cannot send reset token", "userId", userId)
		// If user has no email, they can't reset password this way.
		return errors.NewInvalidInputError("user has no email address configured")
	}

	return nil
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
