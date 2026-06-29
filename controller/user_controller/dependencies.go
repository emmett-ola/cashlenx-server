package user_controller

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
)

var (
	getProfileUser           = user_service.GetService
	updateProfileUser        = user_service.UpdateProfileService
	getUserConfiguration     = user_service.GetConfigurationService
	upsertUserConfiguration  = user_service.UpsertConfigurationService
	createUserByAdmin        = user_service.CreateService
	getUserByID              = user_service.GetUserByObjectId
	updateUserByAdmin        = user_service.UpdateService
	deleteUserByAdmin        = user_service.DeleteService
	changeUserPassword       = user_service.ChangePasswordService
	deleteCurrentUser        = user_service.DeleteService
	requestUserPasswordReset = user_service.RequestPasswordReset
	confirmUserPasswordReset = user_service.ConfirmPasswordReset
)

type updateProfileUserFunc func(string, model.UserProfileUpdateRequest) (model.UserEntity, error)
