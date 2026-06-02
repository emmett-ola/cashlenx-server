package user_cmd

import (
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/user_service"
)

var (
	getUserProfile    = user_service.GetService
	updateUserProfile = user_service.UpdateProfileService
)

type updateUserProfileFunc func(string, model.UserProfileUpdateRequest) (model.UserEntity, error)
