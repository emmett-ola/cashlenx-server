package model

// UserProfileUpdateRequest represents a request to update user profile
type UserProfileUpdateRequest struct {
	Nickname  string `json:"nickname"`
	AvatarUrl string `json:"avatar_url"`
	Gender    string `json:"gender"`
}

// UserChangePasswordRequest represents a request to change password
type UserChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UserChangeEmailRequest represents a request to initiate email change
type UserChangeEmailRequest struct {
	NewEmail string `json:"new_email"`
}

// UserConfirmEmailChangeRequest represents a request to confirm email change
type UserConfirmEmailChangeRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"` // Requires password for extra security
}
