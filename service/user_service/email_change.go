package user_service

import (
	"strings"

	appErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/macar-x/cashlenx-server/validation"
	"golang.org/x/crypto/bcrypt"
)

// RequestEmailChange initiates the email change process
func RequestEmailChange(userId string, newEmail string, verificationToken string) error {
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	if err := validation.ValidateEmail(newEmail); err != nil {
		return err
	}

	if verificationToken == "" {
		return appErrors.NewInvalidInputError("verification token is required")
	}

	existingEmailUser := userRepo.GetUserByEmail(newEmail)
	if !existingEmailUser.Id.IsZero() && existingEmailUser.Id.Hex() != userId {
		return appErrors.NewFieldAlreadyExistsError("new_email", "email address already in use")
	}

	// Get current user
	user := GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return appErrors.NewNotFoundError("user not found")
	}

	verification, err := verification_service.ConsumeVerifiedToken(verificationToken, verification_service.OperationEmailChange)
	if err != nil {
		return err
	}
	if !strings.EqualFold(verification.Payload, newEmail) {
		return appErrors.NewInvalidInputError("verification token email does not match new email")
	}

	// Update user email after the target address has been verified.
	user.EmailAddress = newEmail
	user.IsEmailVerified = true
	_, err = UpdateService(user.Id.Hex(), model.UserDTO{
		EmailAddress:    newEmail,
		IsEmailVerified: true,
	})
	if err != nil {
		return appErrors.NewInternalError("failed to update user profile", err)
	}

	return nil
}

// ConfirmEmailChange finalizes the email change process
func ConfirmEmailChange(userId string, token string, password string) error {
	// Verify password for security
	user := GetUserByObjectId(userId)
	if user.Id.IsZero() {
		return appErrors.NewNotFoundError("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return appErrors.NewUnauthorizedError("invalid password")
	}

	verification, err := verification_service.ConsumeVerifiedToken(token, verification_service.OperationEmailChange)
	if err != nil {
		return err
	}
	newEmail := strings.ToLower(strings.TrimSpace(verification.Payload))

	existingEmailUser := userRepo.GetUserByEmail(newEmail)
	if !existingEmailUser.Id.IsZero() && existingEmailUser.Id.Hex() != userId {
		return appErrors.NewFieldAlreadyExistsError("new_email", "email address already in use")
	}

	user.EmailAddress = newEmail
	user.IsEmailVerified = true
	_, err = UpdateService(user.Id.Hex(), model.UserDTO{
		EmailAddress:    newEmail,
		IsEmailVerified: true,
	})
	if err != nil {
		return err
	}

	return nil
}
