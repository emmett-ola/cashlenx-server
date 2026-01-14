package user_service

import (
	"time"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/mapper/password_reset_mapper"
	"github.com/macar-x/cashlenx-server/mapper/user_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// RequestPasswordReset initiates a password reset request
func RequestPasswordReset(emailOrUsername string) error {
	// Get user by username
	user := GetUserByUsername(emailOrUsername)
	if user.Id.IsZero() {
		// For security reasons, don't reveal if user exists
		return nil
	}

	// Generate a unique token
	token := util.GenerateUUID()

	// Set token expiration time (1 hour from now)
	expirationTime := time.Now().Add(1 * time.Hour)

	// Create password reset token
	resetToken := model.PasswordResetToken{
		Id:        primitive.NewObjectID().Hex(),
		UserId:    user.Id.Hex(),
		Token:     token,
		ExpiresAt: expirationTime,
		CreatedAt: util.GetCurrentTime(),
		UsedAt:    nil,
	}

	// Save token to database
	password_reset_mapper.INSTANCE.CreateToken(resetToken)

	// TODO: Send email with password reset link
	// This would typically involve sending an email with a link like:
	// https://your-app.com/reset-password?token={token}

	return nil
}

// ConfirmPasswordReset confirms a password reset using a token
func ConfirmPasswordReset(token string, newPassword string) error {
	// Validate password
	err := validation.ValidatePassword(newPassword)
	if err != nil {
		return err
	}

	// Get token from database
	resetToken := password_reset_mapper.INSTANCE.GetTokenByToken(token)
	if resetToken.Id == "" {
		return errors.NewInvalidInputError("invalid or expired password reset token")
	}

	// Check if token has expired
	if time.Now().After(resetToken.ExpiresAt) {
		// Delete expired token
		password_reset_mapper.INSTANCE.DeleteToken(resetToken.Id)
		return errors.NewInvalidInputError("invalid or expired password reset token")
	}

	// Check if token has already been used
	if resetToken.UsedAt != nil {
		return errors.NewInvalidInputError("password reset token has already been used")
	}

	// Get user
	user := GetUserByObjectId(resetToken.UserId)
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
	user.UpdatedAt = util.GetCurrentTime()

	// Save updated user
	// Note: We're using the mapper directly here since UpdateService has additional validation
	// that we don't want for password reset (like username uniqueness)
	updatedUser := user_mapper.INSTANCE.UpdateUserByEntity(user.Id.Hex(), user)
	if updatedUser.Id.IsZero() {
		return errors.NewInternalError("failed to update password", nil)
	}

	// Mark token as used
	password_reset_mapper.INSTANCE.MarkTokenAsUsed(resetToken.Id)

	// Delete all other tokens for this user
	password_reset_mapper.INSTANCE.DeleteTokensByUserId(user.Id.Hex())

	return nil
}
