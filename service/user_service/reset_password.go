package user_service

import (
	"strings"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/macar-x/cashlenx-server/util"
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

	return verification_service.SendVerificationCode(string(verification_service.OperationPasswordReset), user.EmailAddress, ipAddress)
}

// ConfirmPasswordReset confirms a password reset using a token
func ConfirmPasswordReset(token string, newPassword string) error {
	// Validate password
	err := validation.ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	verification, err := verification_service.ConsumeVerifiedToken(token, verification_service.OperationPasswordReset)
	if err != nil {
		return err
	}

	// Get user
	user := userRepo.GetUserByEmail(strings.ToLower(strings.TrimSpace(verification.Payload)))
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

	return nil
}
