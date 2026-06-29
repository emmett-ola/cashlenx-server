package user_service

import (
	"strings"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/verification_service"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/validation"
)

// RegisterPublicUser creates a self-registered user after email verification.
func RegisterPublicUser(username, password, emailAddress, verificationToken string) (string, error) {
	registerEnabled := util.GetConfigByKey("auth.registration.enabled")
	if registerEnabled != "true" {
		return "", errors.NewForbiddenError("user registration is disabled")
	}
	if username == "" {
		return "", errors.NewValidationError("username is required")
	}
	if password == "" {
		return "", errors.NewValidationError("password is required")
	}
	if emailAddress == "" {
		return "", errors.NewValidationError("email is required")
	}
	if verificationToken == "" {
		return "", errors.NewValidationError("verification_token is required")
	}
	if err := validation.ValidatePassword(password); err != nil {
		return "", err
	}
	emailAddress = strings.ToLower(strings.TrimSpace(emailAddress))
	if err := validation.ValidateEmail(emailAddress); err != nil {
		return "", err
	}

	existingUser := userRepo.GetUserByUsernameIncludeDeleted(username)
	if !existingUser.Id.IsZero() {
		return "", errors.NewFieldAlreadyExistsError("username", "username already exists")
	}
	existingEmailUser := userRepo.GetUserByEmail(emailAddress)
	if !existingEmailUser.Id.IsZero() {
		return "", errors.NewFieldAlreadyExistsError("email", "email address already exists")
	}

	verification, err := consumeVerifiedToken(verificationToken, verification_service.OperationSignup)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(verification.Payload, emailAddress) {
		return "", errors.NewValidationError("verification token email does not match registration email")
	}

	return CreateService(model.UserDTO{
		Username:        username,
		Password:        password,
		EmailAddress:    emailAddress,
		IsEmailVerified: true,
	}, nil)
}
