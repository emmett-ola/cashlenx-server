package user_service

import (
	"fmt"

	appErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/email"
	"golang.org/x/crypto/bcrypt"
)

// RequestEmailChange initiates the email change process
func RequestEmailChange(userId string, newEmail string, ipAddress string) error {
	// Validate email format
	if newEmail == "" {
		return appErrors.NewInvalidInputError("new email is required")
	}

	// Check if email is already in use
	// Note: Ideally check against all users, but simplified here

	// Get current user
	user := GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return appErrors.NewNotFoundError("user not found")
	}

	// Check if SMTP is configured
	if !email.GetService().IsConfigured() {
		return appErrors.NewInternalError("SMTP service is not configured", nil)
	}

	if err := email.CheckAndRecordPurposeEmailAllowance(
		string(verification_service.OperationEmailChange),
		ipAddress,
		[]string{newEmail},
	); err != nil {
		return err
	}

	// Generate verification token
	// Store the new email as payload in the token
	token, err := verification_service.GenerateVerificationToken(userId, verification_service.OperationEmailChange, newEmail)
	if err != nil {
		util.Logger.Errorw("Failed to generate verification token", "error", err)
		return appErrors.NewInternalError("failed to generate verification token", err)
	}

	// Send verification email first - if this fails, we shouldn't update the user
	subject := "Verify your new email address"
	body := fmt.Sprintf("Please use the following code to verify your new email address: %s\n\nThis code will expire in %s.", token, verificationCodeExpiryText())
	err = email.GetService().SendEmail([]string{newEmail}, subject, body, false)
	if err != nil {
		util.Logger.Errorw("Failed to send verification email", "error", err)
		// Return specific error so client knows it was an SMTP failure
		return err
	}

	// Update user email immediately and set verified to false
	user.EmailAddress = newEmail
	user.IsEmailVerified = false
	_, err = UpdateService(user.Id.Hex(), model.UserDTO{
		EmailAddress:    newEmail,
		IsEmailVerified: false,
	})
	if err != nil {
		return appErrors.NewInternalError("failed to update user profile", err)
	}

	return nil
}

// ConfirmEmailChange finalizes the email change process
func ConfirmEmailChange(userId string, token string, password string) error {
	// Verify token
	// Retrieve new email from token payload
	newEmail, tokenUserId, valid := verification_service.VerifyToken(token, verification_service.OperationEmailChange)
	if !valid {
		return appErrors.NewInvalidInputError("invalid or expired token")
	}

	// Ensure token belongs to the requesting user
	if tokenUserId != userId {
		return appErrors.NewInvalidInputError("invalid token for this user")
	}

	// Verify password for security
	user := GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return appErrors.NewNotFoundError("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return appErrors.NewUnauthorizedError("invalid password")
	}

	// Check if the email in token matches user's current pending email
	if user.EmailAddress != newEmail {
		return appErrors.NewInvalidInputError("email address mismatch, please request a new verification code")
	}

	// Set verified to true
	user.IsEmailVerified = true
	_, err = UpdateService(user.Id.Hex(), model.UserDTO{
		IsEmailVerified: true,
	})
	if err != nil {
		return err
	}

	// Invalidate token
	verification_service.InvalidateToken(token)

	return nil
}
