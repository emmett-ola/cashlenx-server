package user_cmd

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/service/user_service"
)

var (
	getUserProfile         = user_service.GetService
	updateUserProfile      = user_service.UpdateProfileService
	getUserConfig          = user_service.GetConfigurationService
	upsertUserConfig       = user_service.UpsertConfigurationService
	changeUserPassword     = user_service.ChangePasswordService
	requestUserEmailChange = user_service.RequestEmailChange
	confirmUserEmailChange = user_service.ConfirmEmailChange
	deleteUserAccount      = user_service.DeleteService
	exportUserData         = manage_service.UserExportData
	importUserData         = manage_service.UserImportData
)

type updateUserProfileFunc func(string, model.UserProfileUpdateRequest) (model.UserEntity, error)
