package user_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
	"github.com/macar-x/cashlenx-server/util"
)

// GetProfile handles getting the current user's profile
func GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	// Get user
	user := getProfileUser(userID)
	if user.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError("user not found"))
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, buildUserProfileResponse(user))
}

// UpdateProfile handles updating the current user's profile
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	var requestBody model.UserProfileUpdateRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	// Update profile
	updatedUser, err := updateProfileUser(userID, requestBody)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, buildUserProfileResponse(updatedUser))
}

func buildUserProfileResponse(user model.UserEntity) model.UserProfileResponse {
	return model.UserProfileResponse{
		Id:              user.Id.Hex(),
		Username:        user.Username,
		Nickname:        user.Nickname,
		AvatarUrl:       user.AvatarUrl,
		EmailAddress:    user.EmailAddress,
		IsEmailVerified: user.IsEmailVerified,
		Gender:          user.Gender,
		IsActive:        user.IsActive,
		Role:            user.Role,
		CreatedAt:       util.FormatDateToStringWithDash(user.CreateTime),
		UpdatedAt:       util.FormatDateToStringWithDash(user.UpdateTime),
	}
}

// ChangePassword handles changing the current user's password
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	var requestBody model.UserChangePasswordRequest
	if err := util.ParseJSONRequest(r, &requestBody); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("invalid request body"))
		return
	}

	if requestBody.OldPassword == "" || requestBody.NewPassword == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewValidationError("old_password and new_password are required"))
		return
	}

	// Change password
	err := user_service.ChangePasswordService(userID, requestBody.OldPassword, requestBody.NewPassword)
	if err != nil {
		if errors.IsUnauthorizedError(err) {
			util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
			return
		}
		if errors.IsValidationError(err) {
			util.ComposeJSONResponse(w, http.StatusBadRequest, err)
			return
		}
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully. Please login again.",
	})
}

// DeleteAccount handles deleting the current user's account
func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("unauthorized"))
		return
	}

	// Delete account
	err := user_service.DeleteService(userID)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}
