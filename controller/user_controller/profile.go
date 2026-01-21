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
	user := user_service.GetService(userID)
	if user.Id.IsZero() {
		util.ComposeJSONResponse(w, http.StatusNotFound, errors.NewNotFoundError("user not found"))
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, user)
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
	updatedUser, err := user_service.UpdateProfileService(userID, requestBody)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Convert to DTO
	userDTO := model.UserDTO{
		Id:           updatedUser.Id.Hex(),
		Username:     updatedUser.Username,
		IsActive:     updatedUser.IsActive,
		Role:         updatedUser.Role,
		Nickname:     updatedUser.Nickname,
		AvatarUrl:    updatedUser.AvatarUrl,
		EmailAddress: updatedUser.EmailAddress,
		Gender:       updatedUser.Gender,
		CreatedAt:    util.FormatDateToStringWithDash(updatedUser.CreateTime),
		UpdatedAt:    util.FormatDateToStringWithDash(updatedUser.UpdateTime),
	}

	util.ComposeJSONResponse(w, http.StatusOK, userDTO)
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
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
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
		if errors.IsForbiddenError(err) {
			util.ComposeJSONResponse(w, http.StatusForbidden, err)
			return
		}
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}
